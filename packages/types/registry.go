package types

type RegistryType int8
type RegistryAction int8

const (
	// Metadata ecosystem registry (e.g. pages, keys)
	RegistryTypeMetadata RegistryType = iota + 1
)

const (
	RegistryActionRead RegistryAction = iota + 1
	RegistryActionInsert
	RegistryActionUpdate
	RegistryActionNewField // ex new_column
)

type Registry struct {
	Name      string // ex table name
	Ecosystem *Ecosystem
	Type      RegistryType
}

type MetadataStorage interface {
	Insert(registry *Registry, pkValue string, value interface{}) error
	Get(registry *Registry, pkValue string, out interface{}) error

	Find(registry *Registry, findFunc func(value interface{}) bool) (interface{}, error)
	FindMany(registry *Registry, findFunc func(value interface{}) error) ([]interface{}, error)
}

type RegistryAccessor interface {
	// TODO move SmartContract into types package to prevent circular dependency
	//CanAccess(contract *smart.SmartContract, registry *Registry, action RegistryAction) (bool, error)
}
