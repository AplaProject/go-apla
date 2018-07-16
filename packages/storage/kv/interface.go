//go:generate sh -c "mockery -inpkg -name DB -print > file.tmp && mv file.tmp kv_storage_mock.go"
package kv

import "database/sql/driver"

// TODO Delete, Update, Find, Transactions
type Database interface {
	Begin(writeable bool) (Transaction, error)
}

type Transaction interface {
	driver.Tx

	Insert(key, value string) error
	Get(key string) (string, error)
	Walk(keyPattern string, fn func(value string) bool) error
}
