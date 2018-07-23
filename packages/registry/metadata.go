package registry

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/tidwall/buntdb"
)

const keyConvention = "%d.%s.%s"

// metadataTx must be closed by calling Commit() or Rollback() when done
type metadataTx struct {
	db      kv.Database
	tx      kv.Transaction
	durable bool
}

func (m *metadataTx) Insert(registry *types.Registry, pkValue string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "marshalling struct to json")
	}

	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	err = m.tx.Set([]byte(key), jsonValue)
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

	err := m.tx.Delete([]byte(key))
	if err != nil {
		return errors.Wrapf(err, "deleting %s", key)
	}

	return nil
}

func (m *metadataTx) Get(registry *types.Registry, pkValue string, out interface{}) error {
	err := m.refreshTx()
	if err != nil {
		return err
	}
	defer m.endRead()

	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	value, err := m.tx.Get([]byte(key))
	if err != nil {
		return errors.Wrapf(err, "retrieving %s from databse", key)
	}

	rawValue, err := value.Value()
	if err != nil {
		return errors.Wrapf(err, "retrieving %s value from item", key)
	}

	err = json.Unmarshal(rawValue, out)
	if err != nil {
		return errors.Wrapf(err, "unmarshalling value %s to struct", value)
	}

	return nil
}

func (m *metadataTx) Walk(registry *types.Registry, index string, fn func(value string) bool) error {
	err := m.refreshTx()
	if err != nil {
		return err
	}
	defer m.endRead()

	prefix := fmt.Sprintf("%d.%s", registry.Ecosystem.ID, registry.Name)

	iterator := m.tx.NewIterator(badger.IteratorOptions{PrefetchValues: true, PrefetchSize: 1000})
	for iterator.Seek([]byte(prefix)); iterator.ValidForPrefix([]byte(prefix)); iterator.Next() {
		v, err := iterator.Item().Value()
		if err != nil {
			return errors.Wrapf(err, "iterating over %s", prefix)
		}

		if !fn(string(v)) {
			break
		}
	}

	return nil
}

func (m *metadataTx) Rollback() error {
	m.tx = nil
	m.tx.Discard()
	return nil
}

func (m *metadataTx) Commit() error {
	err := m.tx.Commit(nil)
	m.tx = nil
	return err
}

func (m *metadataTx) refreshTx() error {
	if m.durable {
		if m.tx == nil {
			return buntdb.ErrTxClosed
		}

		return nil
	}

	// Non-durable transaction can only be readable. All writable transaction called directly from Begin()
	// and must be committed/rollback manually. So here we create readonly tx without providing any choice
	m.tx = m.db.NewTransaction(false)

	return nil
}

func (m *metadataTx) endRead() error {
	if !m.durable {
		if err := m.tx.Commit(nil); err != nil {
			return errors.Wrapf(err, "ending read transaction")
		}

		m.tx = nil
	}

	return nil
}

type metadataStorage struct {
	db kv.Database
}

func NewMetadataStorage(db kv.Database) types.MetadataRegistryStorage {
	return &metadataStorage{
		db: db,
	}
}

func (m *metadataStorage) Begin() types.MetadataRegistryReaderWriter {
	return &metadataTx{tx: m.db.NewTransaction(true), durable: true}
}

func (m *metadataStorage) Reader() types.MetadataRegistryReader {
	return &metadataTx{}
}
