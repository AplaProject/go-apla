package types

import (
	"database/sql/driver"
)

type RegistryType int8

const (
	RegistryTypeDefault RegistryType = iota
	RegistryTypePrimary
)

type Registry struct {
	Name      string // ex table Name
	Ecosystem *Ecosystem
	Type      RegistryType
}

type MetadataRegistryReader interface {
	Get(registry *Registry, pkValue string, out interface{}) error
	Walk(registry *Registry, field string, fn func(jsonRow string) bool) error
}

type BlockchainContext interface {
	GetBlockHash() []byte
	GetTransactionHash() []byte
}

type MetadataRegistryWriter interface {
	Insert(ctx BlockchainContext, registry *Registry, pkValue string, value interface{}) error
	Update(ctx BlockchainContext, registry *Registry, pkValue string, newValue interface{}) error

	AddIndex(indexes ...Index) error

	driver.Tx
}

type MetadataRegistryReaderWriter interface {
	MetadataRegistryReader
	MetadataRegistryWriter
}

// MetadataRegistryStorage provides a read or read-write transactions for metadata registry
type MetadataRegistryStorage interface {
	MetadataRegistryReader

	// Write/Read transaction. Must be closed by calling Commit() or Rollback() when done.
	Begin() MetadataRegistryReaderWriter

	// Rollback is rollback all block transactions
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
