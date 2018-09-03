package types

import (
	"database/sql/driver"
)

type Registry struct {
	Name      string // ex table Name
	Ecosystem *Ecosystem
}

type MetadataRegistryReader interface {
	Get(registry *Registry, pkValue string, out interface{}) error
	Walk(registry *Registry, field string, fn func(jsonRow string) bool) error
}

type MetadataRegistryWriter interface {
	Insert(registry *Registry, pkValue string, value interface{}) error
	Update(registry *Registry, pkValue string, newValue interface{}) error

	AddIndex(indexes ...Index) error

	driver.Tx

	SetTxHash(txHash []byte)
	SetBlockHash(blockHash []byte)
}

type MetadataRegistryReaderWriter interface {
	MetadataRegistryReader
	MetadataRegistryWriter
}

// MetadataRegistryStorage provides a read or read-write transactions for metadata registry
type MetadataRegistryStorage interface {
	// Write/Read transaction. Must be closed by calling Commit() or Rollback() when done.
	Begin() MetadataRegistryReaderWriter
	// Multiple read-only transactions can be opened even while write transaction is running
	Reader() MetadataRegistryReader

	Rollback(block []byte) error
}

type Index struct {
	Registry *Registry
	Field    string
	SortFn   func(a, b string) bool
}

type Indexer interface {
	GetIndexes() []Index
}
