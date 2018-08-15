package transaction

import (
	"errors"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

var ErrDuplicatedTx = errors.New("Duplicated transaction")

// InsertInLogTx is inserting tx in log
func InsertInLogTx(transaction *model.DbTransaction, binaryTx []byte, time int64) error {
	txHash, err := crypto.Hash(binaryTx)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Fatal("hashing binary tx")
	}
	ltx := &model.LogTransaction{Hash: txHash, Time: time}
	err = ltx.Create(transaction)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("insert logged transaction")
		return utils.ErrInfo(err)
	}
	return nil
}

// CheckLogTx checks if this transaction exists
// And it would have successfully passed a frontal test
func CheckLogTx(txBinary []byte, transactions, txQueue bool) error {
	searchedHash, err := crypto.Hash(txBinary)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal(err)
	}
	logTx := &model.LogTransaction{}
	found, err := logTx.GetByHash(searchedHash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting log transaction by hash")
		return utils.ErrInfo(err)
	}
	if found {
		log.WithFields(log.Fields{"tx_hash": searchedHash, "type": consts.DuplicateObject}).Error("double tx in log transactions")
		return ErrDuplicatedTx
	}

	if transactions {
		// check for duplicate transaction
		tx := &model.Transaction{}
		_, err := tx.GetVerified(searchedHash)
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
		found, err := qtx.GetByHash(nil, searchedHash)
		if found {
			log.WithFields(log.Fields{"tx_hash": searchedHash, "type": consts.DuplicateObject}).Error("double tx in queue")
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
	log.WithFields(log.Fields{"type": consts.BadTxError, "tx_hash": string(hash), "error": errText}).Error("tx marked as bad")

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
	// get parameters for "struct" transactions
	txType, keyID := GetTxTypeAndUserID(binaryTx)

	header, err := CheckTransaction(binaryTx)
	if err != nil {
		MarkTransactionBad(dbTransaction, hash, err.Error())
		return err
	}

	if !( /*txType > 127 ||*/ consts.IsStruct(int(txType))) {
		if header == nil {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("tx header is nil")
			return utils.ErrInfo(errors.New("header is nil"))
		}
		keyID = header.KeyID
	}

	if keyID == 0 {
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
		Type:     int8(txType),
		KeyID:    keyID,
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
