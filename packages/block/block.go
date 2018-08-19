package block

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/protocols"
	"github.com/GenesisKernel/go-genesis/packages/transaction"
	"github.com/GenesisKernel/go-genesis/packages/transaction/custom"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	"github.com/GenesisKernel/go-genesis/packages/types"
	log "github.com/sirupsen/logrus"
)

// Block is storing block data
type Block struct {
	Header       utils.BlockData
	PrevHeader   *utils.BlockData
	MrklRoot     []byte
	BinData      []byte
	Transactions []*transaction.Transaction
	SysUpdate    bool
	GenBlock     bool // it equals true when we are generating a new block
	StopCount    int  // The count of good tx in the block
}

func (b Block) String() string {
	return fmt.Sprintf("header: %s, prevHeader: %s", b.Header, b.PrevHeader)
}

// GetLogger is returns logger
func (b Block) GetLogger() *log.Entry {
	return log.WithFields(log.Fields{"block_id": b.Header.BlockID, "block_time": b.Header.Time, "block_wallet_id": b.Header.KeyID,
		"block_state_id": b.Header.EcosystemID, "block_hash": b.Header.Hash, "block_version": b.Header.Version})
}

// PlaySafe is inserting block safely
func (b *Block) PlaySafe() error {
	logger := b.GetLogger()
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting db transaction")
		return err
	}
	metaDbTx := model.MetadataRegistry.Begin()

	err = b.Play(dbTransaction, metaDbTx)
	if b.GenBlock && b.StopCount > 0 {
		doneTx := b.Transactions[:b.StopCount]
		trData := make([][]byte, 0, b.StopCount)
		for _, tr := range doneTx {
			trData = append(trData, tr.TxFullData)
		}
		NodePrivateKey, _, err := utils.GetNodeKeys()
		if err != nil || len(NodePrivateKey) < 1 {
			log.WithFields(log.Fields{"type": consts.NodePrivateKeyFilename, "error": err}).Error("reading node private key")
			return err
		}

		newBlockData, err := MarshallBlock(&b.Header, trData, b.PrevHeader.Hash, NodePrivateKey)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("marshalling new block")
			return err
		}

		isFirstBlock := b.Header.BlockID == 1
		nb, err := UnmarshallBlock(bytes.NewBuffer(newBlockData), isFirstBlock)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("parsing new block")
			return err
		}
		b.BinData = newBlockData
		b.Transactions = nb.Transactions
		b.MrklRoot = nb.MrklRoot
		b.SysUpdate = nb.SysUpdate
		err = nil
	} else if err != nil {
		dbTransaction.Rollback()
		rbErr := metaDbTx.Rollback()
		if rbErr != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("rollback metadb transaction")
		}
		if b.GenBlock && b.StopCount == 0 {
			if err == ErrLimitStop {
				err = ErrLimitTime
			}
			transaction.MarkTransactionBad(nil, b.Transactions[0].TxHash, err.Error())
		}
		return err
	}

	if err := UpdBlockInfo(dbTransaction, b); err != nil {
		dbTransaction.Rollback()
		return err
	}

	if err := InsertIntoBlockchain(dbTransaction, b); err != nil {
		dbTransaction.Rollback()
		return err
	}

	dbTransaction.Commit()
	if err := metaDbTx.Commit(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("commiting metadb transaction")
		return err
	}
	if b.SysUpdate {
		b.SysUpdate = false
		if err = syspar.SysUpdate(nil); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
			return err
		}
	}
	return nil
}

func (b *Block) readPreviousBlockFromBlockchainTable() error {
	if b.Header.BlockID == 1 {
		b.PrevHeader = &utils.BlockData{}
		return nil
	}

	var err error
	b.PrevHeader, err = GetBlockDataFromBlockChain(b.Header.BlockID - 1)
	if err != nil {
		return utils.ErrInfo(fmt.Errorf("can't get block %d", b.Header.BlockID-1))
	}
	return nil
}

func (b *Block) Play(dbTransaction *model.DbTransaction, metaDb types.MetadataRegistryReaderWriter) error {
	logger := b.GetLogger()
	if _, err := model.DeleteUsedTransactions(dbTransaction); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("delete used transactions")
		return err
	}

	limits := NewLimits(b)
	metaDb.SetBlockHash(b.Header.Hash)

	txHashes := make([][]byte, 0, len(b.Transactions))
	for _, btx := range b.Transactions {
		txHashes = append(txHashes, btx.TxHash)
	}
	seed, err := crypto.CalcChecksum(bytes.Join(txHashes, []byte{}))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("calculating seed")
		return err
	}
	randBlock := rand.New(rand.NewSource(int64(seed)))

	storedTxes, err := model.GetTxesByHashlist(dbTransaction, txHashes)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting txes by hashlist")
		return err
	}

	for curTx, t := range b.Transactions {
		var (
			msg string
			err error
		)
		t.DbTransaction = dbTransaction
		metaDb.SetTxHash(t.TxHash)
		t.MetaDb = metaDb
		t.Rand = randBlock

		model.IncrementTxAttemptCount(dbTransaction, t.TxHash)
		err = dbTransaction.Savepoint(curTx)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("using savepoint")
			return err
		}

		if stx, ok := storedTxes[string(t.TxHash)]; ok {
			stx.Attempt++
			if stx.Attempt >= consts.MaxTXAttempt-1 {
				txString := fmt.Sprintf("tx_hash: %s, tx_data: %s, tx_attempt: %d", stx.Hash, stx.Data, stx.Attempt)
				log.WithFields(log.Fields{"type": consts.BadTxError, "tx_info": txString}).Error("tx attempts exceeded, transaction marked as bad")
			}
		}

		msg, err = t.Play()
		if err == nil && t.TxSmart != nil {
			err = limits.CheckLimit(t)
		}
		if err != nil {
			if err == custom.ErrNetworkStopping {
				return err
			}
			if b.GenBlock && err == ErrLimitStop {
				b.StopCount = curTx
			}

			errRoll := dbTransaction.RollbackSavepoint(curTx)
			if errRoll != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("rolling back to previous savepoint")
				return errRoll
			}
			if b.GenBlock && err == ErrLimitStop {
				if curTx == 0 {
					return err
				}
				break
			}
			// skip this transaction
			transaction.MarkTransactionBad(t.DbTransaction, t.TxHash, err.Error())
			if t.SysUpdate {
				if err = syspar.SysUpdate(t.DbTransaction); err != nil {
					log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
				}
				t.SysUpdate = false
			}
			continue
		}
		err = dbTransaction.ReleaseSavepoint(curTx)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("releasing savepoint")
		}

		model.DecrementTxAttemptCount(dbTransaction, t.TxHash)

		if t.SysUpdate {
			b.SysUpdate = true
			t.SysUpdate = false
		}

		if _, err := model.MarkTransactionUsed(t.DbTransaction, t.TxHash); err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("marking transaction used")
			return err
		}

		// update status
		ts := &model.TransactionStatus{}
		if err := ts.UpdateBlockMsg(t.DbTransaction, b.Header.BlockID, msg, t.TxHash); err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("updating transaction status block id")
			return err
		}
		if err := transaction.InsertInLogTx(t.DbTransaction, t.TxFullData, t.TxTime); err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

// CheckBlock is checking block
func (b *Block) Check() error {
	logger := b.GetLogger()
	// exclude blocks from future
	if b.Header.Time > time.Now().Unix() {
		logger.WithFields(log.Fields{"type": consts.ParameterExceeded}).Error("block time is larger than now")
		return utils.ErrInfo(fmt.Errorf("incorrect block time - block.Header.Time > time.Now().Unix()"))
	}
	if b.PrevHeader == nil || b.PrevHeader.BlockID != b.Header.BlockID-1 {
		if err := b.readPreviousBlockFromBlockchainTable(); err != nil {
			logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("block id is larger then previous more than on 1")
			return utils.ErrInfo(err)
		}
	}

	if b.Header.BlockID == 1 {
		return nil
	}

	// is this block too early? Allowable error = error_time
	if b.PrevHeader != nil {
		if b.Header.BlockID != b.PrevHeader.BlockID+1 {
			logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("block id is larger then previous more than on 1")
			return utils.ErrInfo(fmt.Errorf("incorrect block_id %d != %d +1", b.Header.BlockID, b.PrevHeader.BlockID))
		}

		// skip time validation for first block
		if b.Header.BlockID > 1 {

			exists, err := protocols.NewBlockTimeCounter().BlockForTimeExists(time.Unix(b.Header.Time, 0), int(b.Header.NodePosition))
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("calculating block time")
				return err
			}

			if exists {
				logger.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Warn("incorrect block time")
				return utils.ErrInfo(fmt.Errorf("incorrect block time %d", b.PrevHeader.Time))
			}
		}
	}

	// check each transaction
	txCounter := make(map[int64]int)
	txHashes := make(map[string]struct{})
	for _, t := range b.Transactions {
		hexHash := string(converter.BinToHex(t.TxHash))
		// check for duplicate transactions
		if _, ok := txHashes[hexHash]; ok {
			logger.WithFields(log.Fields{"tx_hash": hexHash, "type": consts.DuplicateObject}).Error("duplicate transaction")
			return utils.ErrInfo(fmt.Errorf("duplicate transaction %s", hexHash))
		}
		txHashes[hexHash] = struct{}{}

		// check for max transaction per user in one block
		txCounter[t.TxKeyID]++
		if txCounter[t.TxKeyID] > syspar.GetMaxBlockUserTx() {
			return utils.ErrInfo(fmt.Errorf("max_block_user_transactions"))
		}

		if err := t.Check(b.Header.Time, false); err != nil {
			return err
		}

	}

	result, err := b.CheckHash()
	if err != nil {
		return utils.ErrInfo(err)
	}
	if !result {
		logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect signature")
		return fmt.Errorf("incorrect signature / p.PrevBlock.BlockId: %d", b.PrevHeader.BlockID)
	}
	return nil
}

// CheckHash is checking hash
func (b *Block) CheckHash() (bool, error) {
	logger := b.GetLogger()
	if b.Header.BlockID == 1 {
		return true, nil
	}
	// check block signature
	if b.PrevHeader != nil {
		nodePublicKey, err := syspar.GetNodePublicKeyByPosition(b.Header.NodePosition)
		if err != nil {
			return false, utils.ErrInfo(err)
		}
		if len(nodePublicKey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node public key is empty")
			return false, utils.ErrInfo(fmt.Errorf("empty nodePublicKey"))
		}

		resultCheckSign, err := utils.CheckSign([][]byte{nodePublicKey}, b.ForSign(), b.Header.Sign, true)
		if err != nil {
			logger.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Error("checking block header sign")
			return false, utils.ErrInfo(fmt.Errorf("err: %v / block.PrevHeader.BlockID: %d /  block.PrevHeader.Hash: %x / ", err, b.PrevHeader.BlockID, b.PrevHeader.Hash))
		}

		return resultCheckSign, nil
	}

	return true, nil
}

func (b Block) ForSha() string {
	return fmt.Sprintf("%d,%x,%s,%d,%d,%d,%d",
		b.Header.BlockID, b.PrevHeader.Hash, b.MrklRoot, b.Header.Time,
		b.Header.EcosystemID, b.Header.KeyID, b.Header.NodePosition)
}

// ForSign from 128 bytes to 512 bytes. Signature of TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
func (b Block) ForSign() string {
	return fmt.Sprintf("0,%v,%x,%v,%v,%v,%v,%s",
		b.Header.BlockID, b.PrevHeader.Hash, b.Header.Time, b.Header.EcosystemID,
		b.Header.KeyID, b.Header.NodePosition, b.MrklRoot)
}

// InsertBlockWOForks is inserting blocks
func InsertBlockWOForks(data []byte, genBlock, firstBlock bool) error {
	block, err := ProcessBlockWherePrevFromBlockchainTable(data, !firstBlock)
	if err != nil {
		return err
	}
	block.GenBlock = genBlock
	if err := block.Check(); err != nil {
		return err
	}

	err = block.PlaySafe()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"block_id": block.Header.BlockID}).Debug("block was inserted successfully")
	return nil
}

// ProcessBlockWherePrevFromBlockchainTable is processing block with in table previous block
func ProcessBlockWherePrevFromBlockchainTable(data []byte, checkSize bool) (*Block, error) {
	if checkSize && int64(len(data)) > syspar.GetMaxBlockSize() {
		log.WithFields(log.Fields{"check_size": checkSize, "size": len(data), "max_size": syspar.GetMaxBlockSize(), "type": consts.ParameterExceeded}).Error("binary block size exceeds max block size")
		return nil, utils.ErrInfo(fmt.Errorf(`len(binaryBlock) > variables.Int64["max_block_size"]`))
	}

	buf := bytes.NewBuffer(data)
	if buf.Len() == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("buffer is empty")
		return nil, fmt.Errorf("empty buffer")
	}

	block, err := UnmarshallBlock(buf, !checkSize)
	if err != nil {
		return nil, err
	}
	block.BinData = data

	if err := block.readPreviousBlockFromBlockchainTable(); err != nil {
		return nil, err
	}

	return block, nil
}
