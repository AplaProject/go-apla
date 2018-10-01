package transaction

import (
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/queue"
)

var ErrDuplicatedTx = errors.New("Duplicated transaction")

// TxParser writes transactions into the queue
func ProcessQueueTransaction(tx *blockchain.Transaction) error {
	// get parameters for "struct" transactions
	keyID := tx.Header.KeyID
	if err := CheckTransaction(tx); err != nil {
		hash, err2 := tx.Hash()
		if err2 != nil {
			return err
		}
		blockchain.SetTransactionError(hash, err.Error())
		return err
	}

	if keyID == 0 {
		errStr := "undefined keyID"
		hash, err := tx.Hash()
		if err != nil {
			return err
		}
		blockchain.SetTransactionError(hash, errStr)
		return errors.New(errStr)
	}

	if err := queue.ProcessTxQueue.Enqueue(tx); err != nil {
		return err
	}
	return nil
}

// AllTxParser parses new transactions
func ProcessTransactionsQueue() error {
	return queue.ValidateTxQueue.ProcessItems(func(tx *blockchain.Transaction) error {
		if err := ProcessQueueTransaction(tx); err != nil {
			return err
		}
		return nil
	})
}
