package transaction

import (
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/queue"
)

var ErrDuplicatedTx = errors.New("Duplicated transaction")

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

		if err := queue.ProcessTxQueue.Enqueue(tx); err != nil {
			return err
		}
		return nil
	})
}
