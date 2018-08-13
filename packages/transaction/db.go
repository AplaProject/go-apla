package transaction

import (
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/queue"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

var ErrDuplicatedTx = errors.New("Duplicated transaction")

func MarkTransactionBad(dbTransaction *model.DbTransaction, hash []byte, errText string) error {
	if hash == nil {
		return nil
	}
	return blockchain.SetTransactionError(hash, errText)
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

	if !consts.IsStruct(int(txType)) {
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

	if err := blockchain.SetTransactionBinary(hash, binaryTx); err != nil {
		return err
	}
	return nil
}

// AllTxParser parses new transactions
func ProcessTransactionsQueue(dbTransaction *model.DbTransaction) error {
	for queue.ValidateTxQueue.Length() >= 0 {
		item, err := queue.ValidateTxQueue.Peek()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("Peek item from validate tx queue")
			return err
		}
		hash, err := crypto.Hash(item.Value)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("hashing value")
			return err
		}
		err = ProcessQueueTransaction(dbTransaction, hash, item.Value, false)
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("transaction parsed successfully")
		if _, err := queue.ValidateTxQueue.Dequeue(); err != nil {
			log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("Dequeuing from validate tx queue")
			return err
		}
	}
	return nil
}
