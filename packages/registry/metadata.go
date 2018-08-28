package registry

import (
	"encoding/json"
	"fmt"

	"sync"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/match"
	"github.com/yddmat/memdb"
)

const keyConvention = "%s.%s"

var (
	ErrUnknownContext   = errors.New("unknown writing operation context (block o/or hash empty)")
	ErrRollbackDisabled = errors.New("rollback is disabled")
)

// metadataTx must be closed by calling Commit() or Rollback() when done
type metadataTx struct {
	db      kv.Database
	tx      kv.Transaction
	durable bool

	rollback *metadataRollback

	currentBlockHash []byte
	currentTxHash    []byte
	stateMu          sync.RWMutex
}

func (m *metadataTx) Insert(registry *types.Registry, pkValue string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "marshalling struct to json")
	}

	key := fmt.Sprintf(keyConvention, registry.Name, pkValue)

	err = m.tx.Set(key, string(jsonValue))
	if err != nil {
		return errors.Wrapf(err, "inserting Value %s to %s registry", value, registry.Name)
	}

	if m.rollback != nil {
		m.stateMu.RLock()
		block := m.currentBlockHash
		tx := m.currentTxHash
		m.stateMu.RUnlock()

		if len(block) == 0 || len(tx) == 0 {
			return ErrUnknownContext
		}

		err = m.rollback.saveState(block, tx, registry, pkValue, "")
		if err != nil {
			return errors.Wrapf(err, "saving rollback info")
		}
	}

	return nil
}

func (m *metadataTx) Update(registry *types.Registry, pkValue string, newValue interface{}) error {
	jsonValue, err := json.Marshal(newValue)
	if err != nil {
		return errors.Wrapf(err, "marshalling struct to json")
	}

	key := fmt.Sprintf(keyConvention, registry.Name, pkValue)

	old, err := m.tx.Update(key, string(jsonValue))
	if err != nil {
		return errors.Wrapf(err, "inserting Value %s to %s registry", pkValue, registry.Name)
	}

	m.stateMu.RLock()
	block := m.currentBlockHash
	tx := m.currentTxHash
	m.stateMu.RUnlock()

	if len(block) == 0 || len(tx) == 0 {
		return ErrUnknownContext
	}

	err = m.rollback.saveState(block, tx, registry, pkValue, old)
	if err != nil {
		return errors.Wrapf(err, "saving rollback info")
	}

	return nil
}

func (m *metadataTx) Get(registry *types.Registry, pkValue string, out interface{}) error {
	err := m.refreshTx()
	if err != nil {
		return err
	}
	defer m.endRead()

	key := fmt.Sprintf(keyConvention, registry.Name, pkValue)

	value, err := m.tx.Get(key)
	if err != nil {
		return errors.Wrapf(err, "retrieving %s from databse", key)
	}

	err = json.Unmarshal([]byte(value), out)
	if err != nil {
		return errors.Wrapf(err, "unmarshalling Value %s to struct", value)
	}

	return nil
}

func (m *metadataTx) Walk(registry *types.Registry, index string, fn func(value string) bool) error {
	err := m.refreshTx()
	if err != nil {
		return err
	}
	defer m.endRead()

	prefix := fmt.Sprintf("%s.*", registry.Name)

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

func (m *metadataTx) AddIndex(index types.Index) {
	m.tx.AddIndex(index)
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
	db       kv.Database
	rollback bool
}

func NewMetadataStorage(db kv.Database, indexes []types.Index, rollback bool) (types.MetadataRegistryStorage, error) {
	if indexes != nil {
		if err := db.Begin(true).AddIndex(indexes...); err != nil {
			return nil, err
		}
	}

	return &metadataStorage{
		db:       db,
		rollback: rollback,
	}, nil
}

func (m *metadataStorage) Begin() types.MetadataRegistryReaderWriter {
	databaseTx := m.db.Begin(true)
	tx := &metadataTx{tx: databaseTx, durable: true}

	if m.rollback {
		tx.rollback = &metadataRollback{tx: databaseTx, txCounter: make(map[string]uint64)}
	}

	return tx
}

func (m *metadataStorage) Rollback(block []byte) error {
	if !m.rollback {
		return ErrRollbackDisabled
	}

	databaseTx := m.db.Begin(true)
	rollback := &metadataRollback{tx: databaseTx, txCounter: make(map[string]uint64)}

	err := rollback.rollbackState(block)
	if err != nil {
		rbErr := databaseTx.Rollback()
		log.WithFields(log.Fields{"type": consts.DBError, "error": rbErr}).Error("rollback metadata db")
		return err
	}

	err = databaseTx.Commit()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("commiting metadata db")
		return err
	}

	return nil
}

func (m *metadataStorage) Reader() types.MetadataRegistryReader {
	return &metadataTx{db: m.db}
}
