package kv

import "github.com/yddmat/memdb"

type DB struct {
	memdb.Database
}

func (db *DB) Begin(writable bool) Transaction {
	return &Tx{Transaction: *db.Database.Begin(writable)}
}

type Index struct {
	memdb.Index
}

type Tx struct {
	memdb.Transaction
}

func (tx *Tx) AddIndex(index *Index) {
	tx.Transaction.AddIndex(&index.Index)
}
