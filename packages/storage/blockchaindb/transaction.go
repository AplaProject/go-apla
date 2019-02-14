package blockchaindb

import (
	"github.com/syndtr/goleveldb/leveldb"
)

type Transaction struct {
	*leveldb.Transaction
}

func (t *Transaction) Rollback() error {
	t.Discard()
	return nil
}

func (t *Transaction) SavePoint(_ string) error {
	return nil
}

func (t *Transaction) RollbackSavePoint(_ string) error {
	return nil
}

func (t *Transaction) ReleaseSavePoint(_ string) error {
	return nil
}
