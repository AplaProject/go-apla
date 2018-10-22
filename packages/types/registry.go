//go:generate sh -c "mockery -inpkg -name MetadataRegistryReaderWriter -print > file.tmp && mv file.tmp metadata_registry_rw_mock.go"

package types

import "encoding/json"

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

type RegistryModel interface {
	ModelName() string
	GetPrimaryKey() string
	CreateFromData(data map[string]interface{}) (RegistryModel, error)
	UpdateFromData(model RegistryModel, data map[string]interface{}) error
	GetData() map[string]interface{}
	json.Unmarshaler
}

type Pricer interface {
	Price() int64
}

type Converter interface {
	CreateFromParams(name string, params map[string]interface{}) (RegistryModel, error)
	UpdateFromParams(name string, value RegistryModel, params map[string]interface{}) error
}

type MetadataRegistryReader interface {
	Get(registry *Registry, pkValue string, out interface{}) error
	GetModel(registry *Registry, pkValue string) (RegistryModel, error)
	Walk(registry *Registry, field string, fn func(jsonRow string) bool) error
}

type MetadataRegistryWriter interface {
	Insert(ctx BlockchainContext, registry *Registry, pkValue string, value interface{}) error
	Update(ctx BlockchainContext, registry *Registry, pkValue string, newValue interface{}) error

	Commit() error
	Rollback() error
}

type MetadataRegistryReaderWriter interface {
	MetadataRegistryReader
	MetadataRegistryWriter
	Pricer
	Converter
	RegistryState
}

type BlockchainContext interface {
	GetBlockHash() []byte
	GetTransactionHash() []byte
}

type RegistryState interface {
	// Rollback is rollback all block transactions
	RollbackBlock(block []byte) error

	// CleanBlockState is removing all rollbacks generated for block transactions
	CleanBlockState(block []byte) error
}

// MetadataRegistryStorage provides a read or read-write transactions for metadata registry
type MetadataRegistryStorage interface {
	// Storage provides unpaid reading
	MetadataRegistryReader

	RegistryState

	// Write/Read transaction. Must be closed by calling Commit() or Rollback() when done.
	Begin() MetadataRegistryReaderWriter
}

type Index struct {
	Registry *Registry
	Name     string
	SortFn   func(a, b string) bool
}

type Indexer interface {
	GetIndexes() []Index
}
