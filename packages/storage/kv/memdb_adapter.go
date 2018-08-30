package kv

import (
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
	"github.com/yddmat/memdb"
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

func (tx *TransactionAdapter) AddIndex(indexes ...types.Index) error {
	idxes := make([]*memdb.Index, 0)
	for _, idx := range indexes {
		if idx.Registry == nil {
			return errors.New("empty registry")
		}

		memdbIndex := memdb.NewIndex(
			idx.Name,
			fmt.Sprintf("%s.*", idx.Registry.Name),
			idx.SortFn,
		)
		idxes = append(idxes, memdbIndex)
	}
	return tx.Transaction.AddIndex(idxes...)
}
