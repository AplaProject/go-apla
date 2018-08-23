package kv

import "github.com/yddmat/memdb"

type DatabaseAdapter struct {
	memdb.Database
}

func (db *DatabaseAdapter) Begin(writable bool) Transaction {
	return &TransactionAdapter{Transaction: *db.Database.Begin(writable)}
}

type IndexAdapter struct {
	memdb.Index
}

type TransactionAdapter struct {
	memdb.Transaction
}

func (tx *TransactionAdapter) AddIndex(index *IndexAdapter) {
	tx.Transaction.AddIndex(&index.Index)
}
