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
	Update(registry *Registry, pkValue string, newValue interface{}) error
	Delete(registry *Registry, pkValue string) error
}

type MetadataRegistry interface {
	MetadataRegistryReader
	MetadataRegistryWriter
	driver.Tx
}

type MetadataRegistryProvider interface {
	// Transaction must be closed by calling Commit() (writable) or Rollback() (writable/readable) when done
	Begin(writable bool) (MetadataRegistry, error)
}

type RegistryAccessor interface {
	// TODO move SmartContract into types package to prevent circular dependency
	//CanAccess(contract *smart.SmartContract, registry *Registry, action RegistryAction) (bool, error)
}
