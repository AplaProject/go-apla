// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package block

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/notificator"
	"github.com/AplaProject/go-apla/packages/protocols"
	"github.com/AplaProject/go-apla/packages/queue"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/storage"
	"github.com/AplaProject/go-apla/packages/storage/multi"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/transaction/custom"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

// Block is storing block data
type PlayableBlock struct {
	Header        blockchain.BlockHeader
	PrevHeader    *blockchain.BlockHeader
	Hash          []byte
	PrevHash      []byte
	MrklRoot      []byte
	BinData       []byte
	Transactions  []*transaction.Transaction
	SysUpdate     bool
	GenBlock      bool // it equals true when we are generating a new block
	StopCount     int  // The count of good tx in the block
	Notifications []smart.NotifyInfo
	LDBTX         *leveldb.Transaction
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
func (b *PlayableBlock) PlaySafe(txs []*blockchain.Transaction) error {
	logger := b.GetLogger()

	mtr, err := storage.NewMultiTransaction()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Starting multi transaction")
		return err
	}

	dbTransaction, err := model.StartTransaction()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("starting db transaction")
		return err
	}
	ldbtx, err := blockchain.DB.OpenTransaction()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("starting transaction")
		return err
	}
	b.LDBTX = ldbtx

	// metaTx := model.MetaStorage.Begin(true)

	err = b.Play(dbTransaction, txs, ldbtx, mtr)
	storage.M.UndoSave()

	if b.GenBlock && b.StopCount > 0 {
		doneTx := b.Transactions[:b.StopCount]
		txs = []*blockchain.Transaction{}
		txHashes := [][]byte{}
		for _, tr := range doneTx {
			bTx, err := tr.ToBlockchainTransaction()
			if err != nil {
				return err
			}
			txHash, err := bTx.Hash()
			if err != nil {
				return err
			}
			txHashes = append(txHashes, txHash)
			txs = append(txs, bTx)
		}
		bBlock := &blockchain.Block{
			Header:   &b.Header,
			TxHashes: txHashes,
			PrevHash: b.PrevHash,
		}
		mrklRoot, err := bBlock.GetMrklRoot()
		bData, err := bBlock.Marshal()
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
		ldbtx.Discard()
		mtr.Rollback()

		if b.GenBlock && b.StopCount == 0 {
			if err == ErrLimitStop {
				err = ErrLimitTime
			}
			BadTxForBan(b.Transactions[0].TxHeader.KeyID)
			bTx, err := b.Transactions[0].ToBlockchainTransaction()
			if err != nil {
				return err
			}
			hash, err := bTx.Hash()
			if err != nil {
				return err
			}
			blockchain.SetTransactionError(ldbtx, hash, err.Error())
		}
		return err
	}

	bBlock, _, err := b.ToBlockchainBlock()
	if err != nil {
		return err
	}
	if err := bBlock.Insert(ldbtx, txs); err != nil {
		dbTransaction.Rollback()
		ldbtx.Discard()
		mtr.Rollback()
		return err
	}

	// TODO double phase commit
	dbTransaction.Commit()
	ldbtx.Commit()
	mtr.Commit()

	b.LDBTX = nil
	if b.SysUpdate {
		b.SysUpdate = false
		if err = syspar.SysUpdate(nil); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
			return err
		}
	}
	for _, item := range b.Notifications {
		if item.Roles {
			notificator.UpdateRolesNotifications(item.EcosystemID, item.List)
		} else {
			notificator.UpdateNotifications(item.EcosystemID, item.List)
		}
	}
	if err := queue.SendBlockQueue.Enqueue(bBlock); err != nil {
		return err
	}
	return nil
}

func (b *PlayableBlock) readPreviousBlockFromBlockchainTable() error {
	if b.Header.BlockID == 1 {
		b.PrevHeader = &blockchain.BlockHeader{}
		return nil
	}

	var err error
	b.PrevHeader, err = GetBlockDataFromBlockChain(nil, b.PrevHash)
	if err != nil {
		return utils.ErrInfo(fmt.Errorf("can't get block %d", b.Header.BlockID-1))
	}
	return nil
}

func (b *PlayableBlock) Play(dbTransaction *model.DbTransaction, txs []*blockchain.Transaction, ldbtx *leveldb.Transaction, mtr *multi.MultiTransaction) error {
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

	var counter uint64

	for curTx, t := range b.Transactions {
		var (
			err error
		)
		t.DbTransaction = dbTransaction
		t.Rand = randBlock
		t.MultiTr = mtr
		t.Counter = &counter

		blockchain.IncrementTxAttemptCount(ldbtx, t.TxHash)
		err = dbTransaction.Savepoint(curTx)
		mtr.SavePoint(fmt.Sprintf("%x", t.TxHash))

		fmt.Printf("Play %x\n", t.TxHash)

		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("using savepoint")
			return err
		}
		var flush []smart.FlushInfo
		msg, flush, err := t.Play()
		if err == nil && t.TxSmart != nil {
			err = limits.CheckLimit(t)
		}
		if err != nil {
			if flush != nil {
				for i := len(flush) - 1; i >= 0; i-- {
					finfo := flush[i]
					if finfo.Prev == nil {
						if finfo.ID != uint32(len(smart.GetVM().Children)-1) {
							logger.WithFields(log.Fields{"type": consts.ContractError, "value": finfo.ID,
								"len": len(smart.GetVM().Children) - 1}).Error("flush rollback")
						} else {
							smart.GetVM().Children = smart.GetVM().Children[:len(smart.GetVM().Children)-1]
							delete(smart.GetVM().Objects, finfo.Name)
						}
					} else {
						smart.GetVM().Children[finfo.ID] = finfo.Prev
						smart.GetVM().Objects[finfo.Name] = finfo.Info
					}
				}
			}
			if err == custom.ErrNetworkStopping {
				return err
			}
			if b.GenBlock && err == ErrLimitStop {
				b.StopCount = curTx
			}

			mtr.RollbackSavePoint()
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
			bTx, err2 := t.ToBlockchainTransaction()
			if err2 != nil {
				return err
			}
			hash, err2 := bTx.Hash()
			if err2 != nil {
				return err
			}
			blockchain.SetTransactionError(ldbtx, hash, err.Error())
			if t.SysUpdate {
				if err = syspar.SysUpdate(t.DbTransaction); err != nil {
					log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
				}
				t.SysUpdate = false
			}
			blockchain.SetTransactionError(ldbtx, hash, msg)
			continue
		}

		mtr.ReleaseSavePoint()
		err = dbTransaction.ReleaseSavepoint(curTx)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("releasing savepoint")
		}

		blockchain.DecrementTxAttemptCount(ldbtx, t.TxHash)

		if t.SysUpdate {
			b.SysUpdate = true
			t.SysUpdate = false
		}
		b.Notifications = append(b.Notifications, t.Notifications...)
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

		resultCheckSign, err := utils.CheckSign([][]byte{nodePublicKey}, []byte(b.ForSign()), b.Header.Sign, true)
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
func InsertBlockWOForks(block *blockchain.Block, txs []*blockchain.Transaction, genBlock, firstBlock bool) error {
	pBlock, err := ProcessBlockWherePrevFromBlockchainTable(block, txs, !firstBlock, nil)
	if err != nil {
		return err
	}
	pBlock.GenBlock = genBlock
	if err := pBlock.Check(); err != nil {
		return err
	}

	err = pBlock.PlaySafe(txs)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"block_id": pBlock.Header.BlockID}).Debug("block was inserted successfully")
	return nil
}

// ProcessBlockWherePrevFromBlockchainTable is processing block with in table previous block
func ProcessBlockWherePrevFromBlockchainTable(block *blockchain.Block, txs []*blockchain.Transaction, checkSize bool, ldbtx *leveldb.Transaction) (*PlayableBlock, error) {
	data, err := block.Marshal()
	if err != nil {
		return nil, err
	}
	if checkSize && int64(len(data)) > syspar.GetMaxBlockSize() {
		log.WithFields(log.Fields{"check_size": checkSize, "size": len(data), "max_size": syspar.GetMaxBlockSize(), "type": consts.ParameterExceeded}).Error("binary block size exceeds max block size")
		return nil, utils.ErrInfo(fmt.Errorf(`len(binaryBlock) > variables.Int64["max_block_size"]`))
	}

	hash, err := crypto.DoubleHash(data)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("Hashing block data")
		return nil, err
	}
	pBlock, err := FromBlockchainBlock(block, txs, hash, ldbtx)
	if err != nil {
		return nil, err
	}
	pBlock.BinData = data

	if err := pBlock.readPreviousBlockFromBlockchainTable(); err != nil {
		return nil, err
	}

	return pBlock, nil
}

func FromBlockchainBlock(b *blockchain.Block, txs []*blockchain.Transaction, hash []byte, ldbtx *leveldb.Transaction) (*PlayableBlock, error) {
	transactions := []*transaction.Transaction{}
	for _, tx := range txs {
		t, err := transaction.FromBlockchainTransaction(tx)
		if err != nil {
			if t != nil && t.TxHash != nil {
				blockchain.SetTransactionError(ldbtx, t.TxHash, err.Error())
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

func (b *PlayableBlock) ToBlockchainBlock() (*blockchain.Block, []*blockchain.Transaction, error) {
	txHashes := [][]byte{}
	transactions := []*blockchain.Transaction{}
	for _, tx := range b.Transactions {
		bTx, err := tx.ToBlockchainTransaction()
		if err != nil {
			return nil, nil, err
		}
		transactions = append(transactions, bTx)
		hash, err := bTx.Hash()
		if err != nil {
			return nil, nil, err
		}
		txHashes = append(txHashes, hash)
	}
	return &blockchain.Block{
		Header:   &b.Header,
		PrevHash: b.PrevHash,
		TxHashes: txHashes,
	}, transactions, nil
}
