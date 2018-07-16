package types

import "database/sql/driver"

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

type MetadataRegistryReader interface {
	Get(registry *Registry, pkValue string, out interface{}) error
	Walk(registry *Registry, fn func(jsonRow string) bool) error
}

type MetadataRegistryWriter interface {
	Insert(registry *Registry, pkValue string, value interface{}) error
}

type MetadataRegistry interface {
	MetadataRegistryReader
	MetadataRegistryWriter
	driver.Tx
}

type MetadataRegistryStorage interface {
	Begin(writable bool) (MetadataRegistry, error)

	// Storage implements StorageTx's methods by wrapping each method in his own transaction
	//MetadataRegistryReader
	//MetadataRegistryWriter
}

type RegistryAccessor interface {
	// TODO move SmartContract into types package to prevent circular dependency
	//CanAccess(contract *smart.SmartContract, registry *Registry, action RegistryAction) (bool, error)
}
