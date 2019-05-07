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

package transaction

import (
	"bytes"
	"errors"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

var (
	ErrDuplicatedTx = errors.New("Duplicated transaction")
	ErrNotComeTime  = errors.New("Transaction processing time has not come")
	ErrExpiredTime  = errors.New("Transaction processing time is expired")
	ErrEarlyTime    = utils.WithBan(errors.New("Early transaction time"))
	ErrEmptyKey     = utils.WithBan(errors.New("KeyID is empty"))
)

// InsertInLogTx is inserting tx in log
func InsertInLogTx(t *Transaction, blockID int64) error {
	ltx := &model.LogTransaction{Hash: t.TxHash, Block: blockID}
	if err := ltx.Create(t.DbTransaction); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("insert logged transaction")
		return utils.ErrInfo(err)
	}
	return nil
}

// CheckLogTx checks if this transaction exists
// And it would have successfully passed a frontal test
func CheckLogTx(txHash []byte, transactions, txQueue bool) error {
	logTx := &model.LogTransaction{}
	found, err := logTx.GetByHash(txHash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting log transaction by hash")
		return utils.ErrInfo(err)
	}
	if found {
		log.WithFields(log.Fields{"tx_hash": txHash, "type": consts.DuplicateObject}).Error("double tx in log transactions")
		return ErrDuplicatedTx
	}

	if transactions {
		// check for duplicate transaction
		tx := &model.Transaction{}
		_, err := tx.GetVerified(txHash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting verified transaction")
			return utils.ErrInfo(err)
		}
		if len(tx.Hash) > 0 {
			log.WithFields(log.Fields{"tx_hash": tx.Hash, "type": consts.DuplicateObject}).Error("double tx in transactions")
			return ErrDuplicatedTx
		}
	}

	if txQueue {
		// check for duplicate transaction from queue
		qtx := &model.QueueTx{}
		found, err := qtx.GetByHash(nil, txHash)
		if found {
			log.WithFields(log.Fields{"tx_hash": txHash, "type": consts.DuplicateObject}).Error("double tx in queue")
			return ErrDuplicatedTx
		}
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting transaction from queue")
			return utils.ErrInfo(err)
		}
	}

	return nil
}

// DeleteQueueTx deletes a transaction from the queue
func DeleteQueueTx(dbTransaction *model.DbTransaction, hash []byte) error {
	delQueueTx := &model.QueueTx{Hash: hash}
	err := delQueueTx.DeleteTx(dbTransaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction from queue")
		return utils.ErrInfo(err)
	}
	// Because we process transactions with verified=0 in queue_parser_tx, after processing we need to delete them
	_, err = model.DeleteTransactionIfUnused(dbTransaction, hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction if unused")
		return utils.ErrInfo(err)
	}
	return nil
}

func MarkTransactionBad(dbTransaction *model.DbTransaction, hash []byte, errText string) error {
	if hash == nil {
		return nil
	}
	model.MarkTransactionUsed(dbTransaction, hash)
	if len(errText) > 255 {
		errText = errText[:255]
	}

	// set loglevel as error because default level setups to "error"
	log.WithFields(log.Fields{"type": consts.BadTxError, "tx_hash": hash, "error": errText}).Error("tx marked as bad")

	// looks like there is not hash in queue_tx in this moment
	qtx := &model.QueueTx{}
	_, err := qtx.GetByHash(dbTransaction, hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting tx by hash from queue")
	}

	if qtx.FromGate == 0 {
		m := &model.TransactionStatus{}
		err = m.SetError(dbTransaction, errText, hash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("setting transaction status error")
			return utils.ErrInfo(err)
		}
	}
	err = DeleteQueueTx(dbTransaction, hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction from queue")
		return utils.ErrInfo(err)
	}

	return nil
}

// TxParser writes transactions into the queue
func ProcessQueueTransaction(dbTransaction *model.DbTransaction, hash, binaryTx []byte, myTx bool) error {
	t, err := UnmarshallTransaction(bytes.NewBuffer(binaryTx), true)
	if err != nil {
		MarkTransactionBad(dbTransaction, hash, err.Error())
		return err
	}

	if err = t.Check(time.Now().Unix(), true); err != nil {
		if err != ErrEarlyTime {
			MarkTransactionBad(dbTransaction, hash, err.Error())
			return err
		}
		return nil
	}

	if t.TxKeyID == 0 {
		errStr := "undefined keyID"
		MarkTransactionBad(dbTransaction, hash, errStr)
		return errors.New(errStr)
	}

	tx := &model.Transaction{}
	_, err = tx.Get(hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting transaction by hash")
		return utils.ErrInfo(err)
	}
	counter := tx.Counter
	counter++
	_, err = model.DeleteTransactionByHash(hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction by hash")
		return utils.ErrInfo(err)
	}

	// put with verified=1
	newTx := &model.Transaction{
		Hash:     hash,
		Data:     binaryTx,
		Type:     int8(t.TxType),
		KeyID:    t.TxKeyID,
		Counter:  counter,
		Verified: 1,
		HighRate: tx.HighRate,
	}
	err = newTx.Create()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating new transaction")
		return utils.ErrInfo(err)
	}

	// remove transaction from the queue (with verified=0)
	err = DeleteQueueTx(dbTransaction, hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting transaction from queue")
		return utils.ErrInfo(err)
	}

	return nil
}

// AllTxParser parses new transactions
func ProcessTransactionsQueue(dbTransaction *model.DbTransaction) error {
	all, err := model.GetAllUnverifiedAndUnusedTransactions()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all unverified and unused transactions")
		return err
	}
	for _, data := range all {
		err := ProcessQueueTransaction(dbTransaction, data.Hash, data.Data, false)
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("transaction parsed successfully")
	}
	return nil
}
