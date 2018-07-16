//go:generate sh -c "mockery -inpkg -name Database -print > file.tmp && mv file.tmp database_mock.go"
//go:generate sh -c "mockery -inpkg -name Transaction -print > file.tmp && mv file.tmp transaction_mock.go"

package kv

import (
	"database/sql/driver"

	"github.com/tidwall/buntdb"
)

// Database and Transaction interfaces currently fits only buntDB realisation
type Database interface {
	Begin(writeable bool) (*buntdb.Tx, error)
}

type Transaction interface {
	driver.Tx

	Get(key string, ignoreExpired ...bool) (val string, err error)
	Set(key, value string, opts *buntdb.SetOptions) (previousValue string, replaced bool, err error)
	Delete(key string) (val string, err error)

	AscendKeys(pattern string, iterator func(key, value string) bool) error
}
