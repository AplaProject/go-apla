package block

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/protocols"
	"github.com/GenesisKernel/go-genesis/packages/transaction"
	"github.com/GenesisKernel/go-genesis/packages/transaction/custom"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// Block is storing block data
type PlayableBlock struct {
	Header       blockchain.BlockHeader
	PrevHeader   *blockchain.BlockHeader
	Hash         []byte
	PrevHash     []byte
	MrklRoot     []byte
	BinData      []byte
	Transactions []*transaction.Transaction
	SysUpdate    bool
	GenBlock     bool // it equals true when we are generating a new block
	StopCount    int  // The count of good tx in the block
}

func (b PlayableBlock) String() string {
	return fmt.Sprintf("header: %s, prevHeader: %s", b.Header, b.PrevHeader)
}

// GetLogger is returns logger
func (b PlayableBlock) GetLogger() *log.Entry {
	return log.WithFields(log.Fields{"block_id": b.Header.BlockID, "block_time": b.Header.Time, "block_wallet_id": b.Header.KeyID,
		"block_state_id": b.Header.EcosystemID, "block_version": b.Header.Version})
}

// PlayBlockSafe is inserting block safely
func (b *PlayableBlock) PlaySafe() error {
	logger := b.GetLogger()
	dbTransaction, err := model.StartTransaction()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting db transaction")
		return err
	}

	err = b.Play(dbTransaction)
	NodePrivateKey, _, err := utils.GetNodeKeys()
	if b.GenBlock && b.StopCount > 0 {
		doneTx := b.Transactions[:b.StopCount]
		transactions := [][]byte{}
		for _, tr := range doneTx {
			transactions = append(transactions, tr.TxFullData)
		}
		if err != nil || len(NodePrivateKey) < 1 {
			log.WithFields(log.Fields{"type": consts.NodePrivateKeyFilename, "error": err}).Error("reading node private key")
			return err
		}
		bBlock := &blockchain.Block{
			Header:       &b.Header,
			Transactions: transactions,
			PrevHash:     b.PrevHash,
		}
		mrklRoot, err := bBlock.GetMrklRoot()
		if err != nil {
			return err
		}
		bData, err := bBlock.Marshal(NodePrivateKey)
		if err != nil {
			return err
		}

		b.BinData = bData
		b.Transactions = doneTx
		b.MrklRoot = mrklRoot
		b.SysUpdate = false
		err = nil
	} else if err != nil {
		dbTransaction.Rollback()
		if b.GenBlock && b.StopCount == 0 {
			if err == ErrLimitStop {
				err = ErrLimitTime
			}
			transaction.MarkTransactionBad(nil, b.Transactions[0].TxHash, err.Error())
		}
		return err
	}

	blockHash, err := crypto.DoubleHash(b.BinData)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("double hashing block data")
		return err
	}
	if err := blockchain.InsertBlock(blockHash, b.ToBlockchainBlock(), NodePrivateKey); err != nil {
		dbTransaction.Rollback()
		return err
	}

	dbTransaction.Commit()
	if b.SysUpdate {
		b.SysUpdate = false
		if err = syspar.SysUpdate(nil); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
			return err
		}
	}
	return nil
}

func (b *PlayableBlock) readPreviousBlockFromBlockchainTable() error {
	if b.Header.BlockID == 1 {
		b.PrevHeader = &blockchain.BlockHeader{}
		return nil
	}

	var err error
	b.PrevHeader, err = GetBlockDataFromBlockChain(b.Hash)
	if err != nil {
		return utils.ErrInfo(fmt.Errorf("can't get block %d", b.Header.BlockID-1))
	}
	return nil
}

func (b *PlayableBlock) Play(dbTransaction *model.DbTransaction) error {
	logger := b.GetLogger()
	limits := NewLimits(b)

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

	for curTx, t := range b.Transactions {
		var (
			err error
		)
		t.DbTransaction = dbTransaction
		t.Rand = randBlock

		blockchain.IncrementTxAttemptCount(t.TxHash)
		err = dbTransaction.Savepoint(curTx)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("using savepoint")
			return err
		}
		_, err = t.Play()
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

		blockchain.DecrementTxAttemptCount(t.TxHash)

		if t.SysUpdate {
			b.SysUpdate = true
			t.SysUpdate = false
		}
	}
	return nil
}

// CheckBlock is checking block
func (b *PlayableBlock) Check() error {
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

		if err := t.Check(b.Header.Time); err != nil {
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
func (b *PlayableBlock) CheckHash() (bool, error) {
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
			return false, utils.ErrInfo(fmt.Errorf("err: %v / block.PrevHeader.BlockID: %d /  block.PrevHeader.Hash: %x / ", err, b.PrevHeader.BlockID, b.PrevHash))
		}

		return resultCheckSign, nil
	}

	return true, nil
}

func (b PlayableBlock) ForSha() string {
	return fmt.Sprintf("%d,%x,%s,%d,%d,%d,%d",
		b.Header.BlockID, b.PrevHash, b.MrklRoot, b.Header.Time,
		b.Header.EcosystemID, b.Header.KeyID, b.Header.NodePosition)
}

// ForSign from 128 bytes to 512 bytes. Signature of TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, state_id, MRKL_ROOT
func (b PlayableBlock) ForSign() string {
	return fmt.Sprintf("0,%v,%x,%v,%v,%v,%v,%s",
		b.Header.BlockID, b.PrevHash, b.Header.Time, b.Header.EcosystemID,
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
func ProcessBlockWherePrevFromBlockchainTable(data []byte, checkSize bool) (*PlayableBlock, error) {
	if checkSize && int64(len(data)) > syspar.GetMaxBlockSize() {
		log.WithFields(log.Fields{"check_size": checkSize, "size": len(data), "max_size": syspar.GetMaxBlockSize(), "type": consts.ParameterExceeded}).Error("binary block size exceeds max block size")
		return nil, utils.ErrInfo(fmt.Errorf(`len(binaryBlock) > variables.Int64["max_block_size"]`))
	}

	blockModel := &blockchain.Block{}
	if err := blockModel.Unmarshal(data); err != nil {
		return nil, err
	}
	hash, err := crypto.DoubleHash(data)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("Hashing block data")
		return nil, err
	}
	block, err := FromBlockchainBlock(blockModel, hash)
	if err != nil {
		return nil, err
	}
	block.BinData = data

	if err := block.readPreviousBlockFromBlockchainTable(); err != nil {
		return nil, err
	}

	return block, nil
}

func FromBlockchainBlock(b *blockchain.Block, hash []byte) (*PlayableBlock, error) {
	transactions := make([]*transaction.Transaction, 0)
	for _, tx := range b.Transactions {
		bufTransaction := bytes.NewBuffer(tx)
		t, err := transaction.UnmarshallTransaction(bufTransaction)
		if err != nil {
			if t != nil && t.TxHash != nil {
				transaction.MarkTransactionBad(t.DbTransaction, t.TxHash, err.Error())
			}
			return nil, fmt.Errorf("parse transaction error(%s)", err)
		}
		t.BlockData = b.Header
		transactions = append(transactions, t)
	}
	return &PlayableBlock{
		Header:       *b.Header,
		Transactions: transactions,
		Hash:         hash,
		PrevHash:     b.PrevHash,
		MrklRoot:     b.MrklRoot,
	}, nil
}

func (b *PlayableBlock) ToBlockchainBlock() *blockchain.Block {
	blockchainBlock := &blockchain.Block{
		Header:   &b.Header,
		PrevHash: b.PrevHash,
	}
	txs := [][]byte{}
	for _, tx := range b.Transactions {
		txs = append(txs, tx.TxFullData)
	}
	blockchainBlock.Transactions = txs
	return blockchainBlock
}
