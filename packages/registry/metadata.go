package registry

import (
	"encoding/json"
	"fmt"

	"sync"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
	"github.com/tidwall/match"
	"github.com/yddmat/memdb"
)

const keyConvention = "%d.%s.%s"

// metadataTx must be closed by calling Commit() or Rollback() when done
type metadataTx struct {
	db      kv.Database
	tx      kv.Transaction
	durable bool

	rollback *MetadataRollback

	currentBlockHash []byte
	currentTxHash    []byte
	stateMu          sync.RWMutex
}

func (m *metadataTx) Insert(registry *types.Registry, pkValue string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "marshalling struct to json")
	}

	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	err = m.tx.Set(key, string(jsonValue))
	if err != nil {
		return errors.Wrapf(err, "inserting value %s to %s registry", value, registry.Name)
	}

	m.stateMu.RLock()
	block := m.currentBlockHash
	tx := m.currentTxHash
	m.stateMu.RUnlock()

	err = m.rollback.saveDocumentState(block, tx, registry, pkValue, "")
	if err != nil {
		return errors.Wrapf(err, "saving rollback info")
	}

	return nil
}

func (m *metadataTx) Update(registry *types.Registry, pkValue string, newValue interface{}) error {
	jsonValue, err := json.Marshal(newValue)
	if err != nil {
		return errors.Wrapf(err, "marshalling struct to json")
	}

	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	err = m.tx.Update(key, string(jsonValue))
	if err != nil {
		return errors.Wrapf(err, "inserting value %s to %s registry", pkValue, registry.Name)
	}

	// TODO save rollback info

	return nil
}

func (m *metadataTx) Get(registry *types.Registry, pkValue string, out interface{}) error {
	err := m.refreshTx()
	if err != nil {
		return err
	}
	defer m.endRead()

	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	value, err := m.tx.Get(key)
	if err != nil {
		return errors.Wrapf(err, "retrieving %s from databse", key)
	}

	err = json.Unmarshal([]byte(value), out)
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

	return m.tx.Ascend(index, func(key, value string) bool {
		if match.Match(key, prefix) {
			return fn(value)
		}

		return true
	})
}

func (m *metadataTx) Rollback() error {
	tx := m.tx
	m.tx = nil
	return tx.Rollback()
}

func (m *metadataTx) Commit() error {
	err := m.tx.Commit()
	m.tx = nil
	return err
}

func (m *metadataTx) SetTxHash(txHash []byte) {
	m.stateMu.Lock()
	m.currentTxHash = txHash
	m.stateMu.Unlock()
}

func (m *metadataTx) SetBlockHash(blockHash []byte) {
	m.stateMu.Lock()
	m.currentBlockHash = blockHash
	m.stateMu.Unlock()
}

func (m *metadataTx) refreshTx() error {
	if m.durable {
		if m.tx == nil {
			return memdb.ErrTxClosed
		}

		return nil
	}

	// Non-durable transaction can only be readable. All writable transaction called directly from Begin()
	// and must be committed/rollback manually. So here we create readonly tx without providing any choice
	m.tx = m.db.Begin(false)

	return nil
}

func (m *metadataTx) endRead() error {
	if !m.durable {
		if err := m.tx.Commit(); err != nil {
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
	databaseTx := m.db.Begin(true)
	return &metadataTx{tx: databaseTx, rollback: &MetadataRollback{tx: databaseTx, txCounter: make(map[string]uint64)}, durable: true}
}

func (m *metadataStorage) Reader() types.MetadataRegistryReader {
	return &metadataTx{}
}
