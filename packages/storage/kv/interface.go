//go:generate sh -c "mockery -inpkg -name Database -print > file.tmp && mv file.tmp database_mock.go"
//go:generate sh -c "mockery -inpkg -name Transaction -print > file.tmp && mv file.tmp transaction_mock.go"

package kv

import (
	"io"

	"github.com/dgraph-io/badger"
)

// Database and Transaction interfaces currently fits only badger implementation
type Database interface {
	io.Closer

	// Starting read/read-write transaction
	NewTransaction(update bool) *badger.Txn
}

type Transaction interface {
	Set(key, val []byte) error
	Delete(key []byte) error
	Get(key []byte) (item *badger.Item, rerr error)

	NewIterator(opt badger.IteratorOptions) *badger.Iterator

	Commit(callback func(error)) error
	Discard()
}
