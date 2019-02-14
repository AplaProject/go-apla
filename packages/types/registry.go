//go:generate sh -c "mockery -inpkg -name MetadataRegistryReaderWriter -print > file.tmp && mv file.tmp metadata_registry_rw_mock.go"

package types

import (
	"encoding/json"

	"github.com/AplaProject/go-apla/packages/blockchain"
)

type MetaRegistryType int8

const (
	RegistryTypeDefault MetaRegistryType = iota
	RegistryTypePrimary
)

type Registry struct {
	Name      string // ex table Name
	Ecosystem *Ecosystem
	Type      MetaRegistryType
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
	StateApplier
}

type BlockchainContext interface {
	GetBlockHash() []byte
	GetTransactionHash() []byte
}

// MetadataRegistryStorage provides a read or read-write transactions for metadata registry
type MetadataRegistryStorage interface {
	// Storage provides unpaid reading
	MetadataRegistryReader

	// Write/Read transaction. Must be closed by calling Commit() or Rollback() when done.
	Begin(saver StateStorage) MetadataRegistryReaderWriter
}
type Index struct {
	Registry *Registry
	Name     string
	SortFn   func(a, b string) bool
}

type Indexer interface {
	GetIndexes() []Index
}

type DBType int8

const (
	DBTypeMeta DBType = iota
	DBTypeUsers
	DBTypeBlockChain
)

type State struct {
	Transaction []byte `json:"t"`
	DBType      DBType `json:"d"`
	Key         string `json:"k"`
	Value       string `json:"v"`
}

type StateStorage interface {
	Save(State) error
	Get() ([]State, error)
}

type StateApplier interface {
	Apply(State) error
}

type MultiTransaction interface {
	GetMetadataRegistry() MetadataRegistryReaderWriter
	GetBlockChainRegistry() blockchain.LevelDBGetterPutterDeleter
	GetUsersRegistry() DBTransaction
}

type DBTransaction interface {
	Rollback() error
	Commit() error
	SavePoint(id string) error
	ReleaseSavePoint(id string) error
	RollbackSavePoint(id string) error
}

type SourceType int8

func (s SourceType) String() string {
	switch s {
	case SourceTypeContract:
		return "contract"
	case SourceTypeMem:
		return "mem"
	case SourceTypeRegistry:
		return "registry"
	default:
		return ""
	}
}

const (
	SourceTypeContract SourceType = iota
	SourceTypeMem
	SourceTypeRegistry
)

type UndoState struct {
	Type  SourceType `json:"type"`
	Table string     `json:"table,omitempty"`
	Key   string     `json:"key"`
	Value string     `json:"value,omitempty"`
}

type UndoStack interface {
	PushState(*UndoState)
	Stack() []*UndoState
	Reset()
}
