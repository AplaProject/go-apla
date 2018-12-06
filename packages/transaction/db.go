package transaction

import (
	"errors"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/queue"
)

var ErrDuplicatedTx = errors.New("Duplicated transaction")
var ErrEarlyTime = errors.New("Early transaction time")

// AllTxParser parses new transactions
func ProcessTransactionsQueue() error {
	return queue.ValidateTxQueue.ProcessItems(func(tx *blockchain.Transaction) error {
		if err := CheckTransaction(tx); err != nil {
			hash, err2 := tx.Hash()
			if err2 != nil {
				return err
			}
			blockchain.SetTransactionError(nil, hash, err.Error())
			return err
		}

		if tx.Header.KeyID == 0 {
			errStr := "undefined keyID"
			hash, err := tx.Hash()
			if err != nil {
				return err
			}
			blockchain.SetTransactionError(nil, hash, errStr)
			return errors.New(errStr)
		}

		if err := blockchain.InsertTxToProcess(nil, tx); err != nil {
			return err
		}
		return nil
	})
}
