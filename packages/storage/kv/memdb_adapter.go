package kv

import (
	"github.com/GenesisKernel/memdb"
)

type DatabaseAdapter struct {
	memdb.Database
}

func (db *DatabaseAdapter) Begin(writable bool) Transaction {
	return &TransactionAdapter{Transaction: *db.Database.Begin(writable)}
}

type TransactionAdapter struct {
	memdb.Transaction
}

func (tx *TransactionAdapter) AddIndex(indexes ...Index) error {
	idxes := make([]*memdb.Index, 0)

	for _, idx := range indexes {
		memdbIndex := memdb.NewIndex(
			idx.Name,
			idx.Pattern,
			idx.SortFn,
		)

		idxes = append(idxes, memdbIndex)
	}

	return tx.Transaction.AddIndex(idxes...)
}
