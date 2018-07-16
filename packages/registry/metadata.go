package registry

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
)

const keyConvention = "%d.%s.%s"

// metadataTx must be closed by calling Commit() or Rollback() when done
type metadataTx struct {
	tx kv.Transaction
}

func (m *metadataTx) Insert(registry *types.Registry, pkValue string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "marshalling struct to json")
	}

	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	_, _, err = m.tx.Set(key, string(jsonValue), nil)
	if err != nil {
		return errors.Wrapf(err, "inserting value %s to %s registry", value, registry.Name)
	}

	return nil
}

func (m *metadataTx) Update(registry *types.Registry, pkValue string, newValue interface{}) error {
	return m.Insert(registry, pkValue, newValue)
}

func (m *metadataTx) Delete(registry *types.Registry, pkValue string) error {
	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	_, err := m.tx.Delete(key)
	if err != nil {
		return errors.Wrapf(err, "deleting %s", key)
	}

	return nil
}

func (m *metadataTx) Get(registry *types.Registry, pkValue string, out interface{}) error {
	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	value, err := m.tx.Get(key)
	if err != nil {
		return errors.Wrapf(err, "retrieving %s from service registry", key)
	}

	err = json.Unmarshal([]byte(value), out)
	if err != nil {
		return errors.Wrapf(err, "unmarshalling value %s to struct", value)
	}

	return nil
}

func (m *metadataTx) Walk(registry *types.Registry, fn func(value string) bool) error {
	pattern := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, "*")

	err := m.tx.AscendKeys(pattern, func(key, value string) bool {
		return fn(value)
	})

	if err != nil {
		return errors.Wrapf(err, "walking through %s", pattern)
	}

	return nil
}

func (m *metadataTx) Rollback() error { return m.tx.Rollback() }

func (m *metadataTx) Commit() error { return m.tx.Commit() }

type metadataProvider struct {
	db kv.Database
}

func NewMetadataProvider(db kv.Database) types.MetadataRegistryProvider {
	return &metadataProvider{
		db: db,
	}
}

// Multiple read-only transactions can be opened at the same time but there can only be one read/write transaction at a time
func (m *metadataProvider) Begin(writable bool) (types.MetadataRegistry, error) {
	dbTx, err := m.db.Begin(writable)
	if err != nil {
		return nil, errors.Wrapf(err, "starting transaction")
	}

	return &metadataTx{tx: dbTx}, nil
}
