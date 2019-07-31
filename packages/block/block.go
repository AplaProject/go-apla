// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package block

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/notificator"
	"github.com/AplaProject/go-apla/packages/protocols"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/transaction/custom"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/AplaProject/go-apla/packages/utils"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	ErrIncorrectRollbackHash = errors.New("Rollback hash doesn't match")
	ErrEmptyBlock            = errors.New("Block doesn't contain transactions")

	errTxAttempts = "The limit of attempts has been reached"
)

// Block is storing block data
type Block struct {
	Header            utils.BlockData
	PrevHeader        *utils.BlockData
	PrevRollbacksHash []byte
	MrklRoot          []byte
	BinData           []byte
	Transactions      []*transaction.Transaction
	SysUpdate         bool
	GenBlock          bool // it equals true when we are generating a new block
	Notifications     []types.Notifications
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

	inputTx := b.Transactions[:]
	err = b.Play(dbTransaction)
	if err != nil {
		dbTransaction.Rollback()
		if b.GenBlock && len(b.Transactions) == 0 {
			if err == ErrLimitStop {
				err = ErrLimitTime
			}
			BadTxForBan(inputTx[0].TxHeader.KeyID)
			transaction.MarkTransactionBad(nil, inputTx[0].TxHash, err.Error())
		}
		return err
	}

	if b.GenBlock {
		if len(b.Transactions) == 0 {
			dbTransaction.Commit()
			return ErrEmptyBlock
		} else if len(inputTx) != len(b.Transactions) {
			if err = b.repeatMarshallBlock(); err != nil {
				dbTransaction.Rollback()
				return err
			}
		}
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
	if b.SysUpdate {
		b.SysUpdate = false
		if err = syspar.SysUpdate(nil); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
			return err
		}
	}

	for _, q := range b.Notifications {
		q.Send()
	}
	return nil
}

func (b *Block) repeatMarshallBlock() error {
	trData := make([][]byte, 0, len(b.Transactions))
	for _, tr := range b.Transactions {
		trData = append(trData, tr.TxFullData)
	}
	NodePrivateKey, _, err := utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) < 1 {
		log.WithFields(log.Fields{"type": consts.NodePrivateKeyFilename, "error": err}).Error("reading node private key")
		return err
	}

	newBlockData, err := MarshallBlock(&b.Header, trData, b.PrevHeader, NodePrivateKey)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("marshalling new block")
		return err
	}

	nb, err := UnmarshallBlock(bytes.NewBuffer(newBlockData), true)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("parsing new block")
		return err
	}
	b.BinData = newBlockData
	b.Transactions = nb.Transactions
	b.MrklRoot = nb.MrklRoot
	b.SysUpdate = nb.SysUpdate
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
		return errors.Wrapf(err, "Can't get block %d", b.Header.BlockID-1)
	}
	return nil
}

func (b *Block) Play(dbTransaction *model.DbTransaction) error {
	logger := b.GetLogger()
	if _, err := model.DeleteUsedTransactions(dbTransaction); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("delete used transactions")
		return err
	}

	limits := NewLimits(b)
	rand := utils.NewRand(b.Header.Time)
	var timeLimit int64
	if b.GenBlock {
		timeLimit = syspar.GetMaxBlockGenerationTime()
	}

	proccessedTx := make([]*transaction.Transaction, 0, len(b.Transactions))
	defer func() {
		if b.GenBlock {
			b.Transactions = proccessedTx
		}
	}()

	for curTx, t := range b.Transactions {
		var (
			msg string
			err error
		)
		t.DbTransaction = dbTransaction
		t.Rand = rand.BytesSeed(t.TxHash)
		t.Notifications = notificator.NewQueue()

		var attempts int64
		if b.GenBlock {
			attempts, err = model.IncrementTxAttemptCount(t.TxHash)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("increment attempts")
				return err
			}
			if attempts >= consts.MaxTXAttempt {
				transaction.MarkTransactionBad(t.DbTransaction, t.TxHash, errTxAttempts)
				continue
			}
		}
		err = dbTransaction.Savepoint(curTx)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("using savepoint")
			return err
		}
		var flush []smart.FlushInfo
		t.GenBlock = b.GenBlock
		t.TimeLimit = timeLimit
		msg, flush, err = t.Play()
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

			errRoll := dbTransaction.RollbackSavepoint(curTx)
			if errRoll != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("rolling back to previous savepoint")
				return errRoll
			}
			if b.GenBlock && err != nil {
				if err == ErrLimitStop {
					if curTx == 0 {
						return err
					}
					break
				}
				if strings.Contains(err.Error(), script.ErrVMTimeLimit.Error()) { // very heavy tx
					err = ErrLimitTime
				}
			}
			// skip this transaction
			transaction.MarkTransactionBad(t.DbTransaction, t.TxHash, err.Error())
			if t.SysUpdate {
				if err := syspar.SysUpdate(t.DbTransaction); err != nil {
					log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
				}
				t.SysUpdate = false
			}

			if b.GenBlock {
				continue
			}

			return err
		}
		err = dbTransaction.ReleaseSavepoint(curTx)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("releasing savepoint")
		}
		if b.GenBlock {
			if err = model.DecrementTxAttemptCount(t.TxHash); err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": t.TxHash}).Error("decrement attempts")
			}
		}
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
		if err := transaction.InsertInLogTx(t, b.Header.BlockID); err != nil {
			return utils.ErrInfo(err)
		}

		if t.Notifications.Size() > 0 {
			b.Notifications = append(b.Notifications, t.Notifications)
		}

		proccessedTx = append(proccessedTx, t)
	}

	return nil
}

var (
	ErrIcorrectBlockTime = utils.WithBan(errors.New("Incorrect block time"))
)

// CheckBlock is checking block
func (b *Block) Check() error {
	logger := b.GetLogger()
	// exclude blocks from future
	if b.Header.Time > time.Now().Unix() {
		logger.WithFields(log.Fields{"type": consts.ParameterExceeded}).Error("block time is larger than now")
		return ErrIcorrectBlockTime

	}
	if b.PrevHeader == nil || b.PrevHeader.BlockID != b.Header.BlockID-1 {
		if err := b.readPreviousBlockFromBlockchainTable(); err != nil {
			logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("block id is larger then previous more than on 1")
			return err
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
				return utils.WithBan(fmt.Errorf("Incorrect block time %d", b.PrevHeader.Time))
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
			return utils.WithBan(utils.ErrInfo(fmt.Errorf("max_block_user_transactions")))
		}

		if err := t.Check(b.Header.Time, false); err != nil {
			return errors.Wrap(err, "check transaction")
		}
	}

	result, err := b.CheckHash()
	if err != nil {
		return utils.WithBan(err)
	}
	if !result {
		logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect signature")
		return utils.WithBan(fmt.Errorf("incorrect signature / p.PrevBlock.BlockId: %d", b.PrevHeader.BlockID))
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

		signSource := b.Header.ForSign(b.PrevHeader, b.MrklRoot)

		resultCheckSign, err := utils.CheckSign(
			[][]byte{nodePublicKey},
			[]byte(signSource),
			b.Header.Sign,
			true)

		if err != nil {
			if err == crypto.ErrIncorrectSign {
				if !bytes.Equal(b.PrevRollbacksHash, b.PrevHeader.RollbacksHash) {
					return false, ErrIncorrectRollbackHash
				}
			}
			logger.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Error("checking block header sign")
			return false, utils.ErrInfo(fmt.Errorf("err: %v / block.PrevHeader.BlockID: %d /  block.PrevHeader.Hash: %x / ", err, b.PrevHeader.BlockID, b.PrevHeader.Hash))
		}

		return resultCheckSign, nil
	}

	return true, nil
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

var (
	ErrMaxBlockSize    = utils.WithBan(errors.New("Block size exceeds maximum limit"))
	ErrZeroBlockSize   = utils.WithBan(errors.New("Block size is zero"))
	ErrUnmarshallBlock = utils.WithBan(errors.New("Unmarshall block"))
)

// ProcessBlockWherePrevFromBlockchainTable is processing block with in table previous block
func ProcessBlockWherePrevFromBlockchainTable(data []byte, checkSize bool) (*Block, error) {
	if checkSize && int64(len(data)) > syspar.GetMaxBlockSize() {
		log.WithFields(log.Fields{"check_size": checkSize, "size": len(data), "max_size": syspar.GetMaxBlockSize(), "type": consts.ParameterExceeded}).Error("binary block size exceeds max block size")
		return nil, ErrMaxBlockSize
	}

	buf := bytes.NewBuffer(data)
	if buf.Len() == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("buffer is empty")
		return nil, ErrZeroBlockSize
	}

	block, err := UnmarshallBlock(buf, true)
	if err != nil {
		return nil, errors.Wrap(ErrUnmarshallBlock, err.Error())
	}
	block.BinData = data

	if err := block.readPreviousBlockFromBlockchainTable(); err != nil {
		return nil, err
	}

	return block, nil
}
