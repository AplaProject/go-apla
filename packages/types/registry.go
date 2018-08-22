package types

import (
	"database/sql/driver"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
)

type RegistryType int8
type RegistryAction int8

//const (
// Metadata ecosystem registry (e.g. pages, keys)
//RegistryTypeMetadata RegistryType = iota + 1
//)

//const (
//	RegistryActionRead RegistryAction = iota + 1
//	RegistryActionInsert
//	RegistryActionUpdate
//	RegistryActionNewField // ex new_column
//)

type Registry struct {
	Name      string // ex table name
	Ecosystem *Ecosystem
	//Type      RegistryType
}

type MetadataRegistryReader interface {
	Get(registry *Registry, pkValue string, out interface{}) error
	Walk(registry *Registry, index string, fn func(jsonRow string) bool) error
}

type MetadataRegistryWriter interface {
	Insert(registry *Registry, pkValue string, value interface{}) error
	Update(registry *Registry, pkValue string, newValue interface{}) error

	AddIndex(index *kv.Index)

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

type RegistryAccessor interface {
	// TODO move SmartContract into types package to prevent circular dependency
	//CanAccess(contract *smart.SmartContract, registry *Registry, action RegistryAction) (bool, error)
}
