package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const keyConvention = "%s.%s.%s"

var (
	ErrUnknownContext   = errors.New("unknown writing operation context")
	ErrWrongRegistry    = errors.New("wrong registry")
	ErrRollbackDisabled = errors.New("rollback is disabled")
)

// tx must be closed by calling Commit() or Rollback() when done
type tx struct {
	db kv.Database
	tx kv.Transaction

	pricing      bool
	priceCounter priceCounter

	saveState bool
	rollback  rollback

	indexer registryIndexer
}

func (m *tx) Insert(ctx types.BlockchainContext, registry *types.Registry, pkValue string, value interface{}) error {
	if m.saveState && (len(ctx.GetBlockHash()) == 0 || len(ctx.GetTransactionHash()) == 0) {
		return ErrUnknownContext
	}

	key, jsonValue, err := m.prepareValue(registry, pkValue, value)
	if err != nil {
		return err
	}

	err = m.tx.Set(key, string(jsonValue))
	if err != nil {
		return errors.Wrapf(err, "inserting value %s to %s registry", value, registry.Name)
	}

	n := model.Ecosystem{}.ModelName()
	if registry.Name == n {
		if err := m.indexer.addPrimaryValue(m.tx, pkValue); err != nil {
			return err
		}
	} else if m.pricing {
		if err := m.priceCounter.Add(Set, registry); err != nil {
			return err
		}
	}

	if m.saveState && registry.Name != "ecosystems" {
		err = m.rollback.saveState(ctx.GetBlockHash(), ctx.GetTransactionHash(), registry, pkValue, "")
		if err != nil {
			return errors.Wrapf(err, "saving rollback info")
		}
	}

	return nil
}

func (m *tx) Update(ctx types.BlockchainContext, registry *types.Registry, pkValue string, newValue interface{}) error {
	if m.saveState && (len(ctx.GetBlockHash()) == 0 || len(ctx.GetTransactionHash()) == 0) {
		return ErrUnknownContext
	}

	key, jsonValue, err := m.prepareValue(registry, pkValue, newValue)
	if err != nil {
		return err
	}

	old, err := m.tx.Update(key, string(jsonValue))
	if err != nil {
		return errors.Wrapf(err, "inserting value %s to %s registry", pkValue, registry.Name)
	}

	if registry.Name != "ecosystems" && m.pricing {
		if err := m.priceCounter.Add(Update, registry); err != nil {
			return err
		}
	}

	if m.saveState && registry.Name != "ecosystems" {
		err = m.rollback.saveState(ctx.GetBlockHash(), ctx.GetTransactionHash(), registry, pkValue, old)
		if err != nil {
			return errors.Wrapf(err, "saving rollback info")
		}
	}

	return nil
}

func (m *tx) Get(registry *types.Registry, pkValue string, out interface{}) error {
	key, err := m.formatKey(registry, pkValue)
	if err != nil {
		return err
	}

	value, err := m.tx.Get(key)
	if err != nil {
		return errors.Wrapf(err, "retrieving %s from database", key)
	}

	err = json.Unmarshal([]byte(value), out)
	if err != nil {
		return errors.Wrapf(err, "unmarshalling value %s to struct", value)
	}

	if m.pricing {
		if err := m.priceCounter.Add(Get, registry); err != nil {
			return err
		}
	}

	return nil
}

func (m *tx) Get2(registry *types.Registry, pkValue string) (types.RegistryModel, error) {
	key, err := m.formatKey(registry, pkValue)
	if err != nil {
		return nil, err
	}

	value, err := m.tx.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving %s from database", key)
	}

	registries := model.GetRegistries()
	var out types.RegistryModel
	for _, r := range registries {
		if r.ModelName() == registry.Name {
			out = r
			break
		}
	}

	err = json.Unmarshal([]byte(value), &out)
	if err != nil {
		return nil, errors.Wrapf(err, "unmarshalling value %s to struct", value)
	}

	if m.pricing {
		if err := m.priceCounter.Add(Get, registry); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (m *tx) Walk(registry *types.Registry, field string, fn func(value string) bool) error {
	if err := m.tx.Ascend(m.indexer.formatIndexName(registry, field), func(key, value string) bool {
		return fn(value)
	}); err != nil {
		return err
	}

	if m.pricing {
		if err := m.priceCounter.AddWalk(registry, field); err != nil {
			return err
		}
	}

	return nil
}

func (m *tx) Rollback() error {
	err := m.tx.Rollback()
	if err != nil {
		return err
	}

	m.closeTx()
	return nil
}

func (m *tx) Commit() error {
	err := m.tx.Commit()
	if err != nil {
		return err
	}

	m.closeTx()
	return nil
}

func (m *tx) Price() int64 {
	if m.pricing {
		return m.priceCounter.GetCurrentPrice()
	}

	return 0
}

func (m *tx) CreateFromParams(name string, params map[string]interface{}) (types.RegistryModel, error) {
	return converter{}.createFromParams(name, params)
}

func (m *tx) UpdateFromParams(name string, value types.RegistryModel, params map[string]interface{}) error {
	return converter{}.updateFromParams(name, value, params)
}

func (m *tx) closeTx() {
	m.tx = nil
}

func (m *tx) prepareValue(registry *types.Registry, pkValue string, newValue interface{}) (string, string, error) {
	jsonValue, err := json.Marshal(newValue)
	if err != nil {
		return "", "", errors.Wrapf(err, "marshalling struct to json")
	}

	key, err := m.formatKey(registry, pkValue)
	if err != nil {
		return "", "", err
	}

	return key, string(jsonValue), nil
}

func (m *tx) formatKey(reg *types.Registry, pk string) (string, error) {
	if reg.Name == "ecosystems" {
		return fmt.Sprintf("%s.%s", reg.Name, pk), nil
	}

	if reg.Ecosystem == nil {
		return "", ErrWrongRegistry
	}

	return fmt.Sprintf(keyConvention, reg.Name, reg.Ecosystem.Name, pk), nil
}

type storage struct {
	db      kv.Database
	indexer registryIndexer

	rollback bool
	pricing  bool
}

func NewMetadataStorage(db kv.Database, indexes []types.Index, rollback bool, pricing bool) (types.MetadataRegistryStorage, error) {
	ms := &storage{
		db:       db,
		indexer:  newIndexer(indexes),
		rollback: rollback,
		pricing:  pricing,
	}

	kvTx := db.Begin(true)
	if err := ms.indexer.init(kvTx); err != nil {
		return nil, err
	}

	if err := kvTx.Commit(); err != nil {
		return nil, err
	}

	return ms, nil
}

func (m *storage) Begin() types.MetadataRegistryReaderWriter {
	databaseTx := m.db.Begin(true)
	tx := &tx{tx: databaseTx, indexer: m.indexer}

	if m.rollback {
		tx.saveState = true
		tx.rollback = rollback{tx: databaseTx, counter: counter{txCounter: make(map[string]uint64)}}
	}

	if m.pricing {
		tx.pricing = true
		tx.priceCounter = priceCounter{tx: databaseTx, indexer: m.indexer}
	}

	return tx
}

func (m *storage) Walk(registry *types.Registry, field string, fn func(value string) bool) error {
	tx := &tx{tx: m.db.Begin(false)}
	defer tx.Rollback()
	return tx.Walk(registry, field, fn)
}

func (m *storage) Get(registry *types.Registry, pkValue string, out interface{}) error {
	tx := &tx{tx: m.db.Begin(false)}
	defer tx.Rollback()
	return tx.Get(registry, pkValue, out)
}

func (m *storage) Get2(registry *types.Registry, pkValue string) (types.RegistryModel, error) {
	tx := &tx{tx: m.db.Begin(false)}
	defer tx.Rollback()
	return tx.Get2(registry, pkValue)
}

func (m *storage) Rollback(block []byte) error {
	if !m.rollback {
		return ErrRollbackDisabled
	}

	databaseTx := m.db.Begin(true)
	rollback := &rollback{tx: databaseTx, counter: counter{txCounter: make(map[string]uint64)}}

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

func (m *storage) Reader() types.MetadataRegistryReader {
	return &tx{db: m.db}
}
