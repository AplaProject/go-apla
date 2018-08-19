package kv

import "github.com/yddmat/memdb"

type DB struct {
	memdb.Database
}

func (db *DB) Begin(writable bool) Transaction {
	return Transaction(db.Database.Begin(writable))
}
