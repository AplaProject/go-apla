package registry

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

// metadataTx must be closed by calling Commit() or Rollback() when done
type metadataTx struct {
	db kv.Database
	tx kv.Transaction

	price priceCounter

	rollback *metadataRollback
	indexer  registryIndexer
}

func (m *metadataTx) Insert(ctx types.BlockchainContext, registry *types.Registry, pkValue string, value interface{}) error {
	if m.rollback != nil && (len(ctx.GetBlockHash()) == 0 || len(ctx.GetTransactionHash()) == 0) {
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

	if registry.Name == "ecosystem" {
		if err := m.indexer.addPrimaryValue(m.tx, pkValue); err != nil {
			return err
		}
	} else {
		if err := m.price.Add(Set, registry); err != nil {
			return err
		}
	}

	if m.rollback != nil {
		err = m.rollback.saveState(ctx.GetBlockHash(), ctx.GetTransactionHash(), registry, pkValue, "")
		if err != nil {
			return errors.Wrapf(err, "saving rollback info")
		}
	}

	return nil
}

func (m *metadataTx) Update(ctx types.BlockchainContext, registry *types.Registry, pkValue string, newValue interface{}) error {
	if m.rollback != nil && (len(ctx.GetBlockHash()) == 0 || len(ctx.GetTransactionHash()) == 0) {
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

	if registry.Name != "ecosystem" {
		if err := m.price.Add(Update, registry); err != nil {
			return err
		}
	}

	if m.rollback != nil {
		err = m.rollback.saveState(ctx.GetBlockHash(), ctx.GetTransactionHash(), registry, pkValue, old)
		if err != nil {
			return errors.Wrapf(err, "saving rollback info")
		}
	}

	return nil
}

func (m *metadataTx) Get(registry *types.Registry, pkValue string, out interface{}) error {
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

	if err := m.price.Add(Get, registry); err != nil {
		return err
	}

	return nil
}

func (m *metadataTx) Walk(registry *types.Registry, field string, fn func(value string) bool) error {
	if err := m.tx.Ascend(m.indexer.formatIndexName(registry, field), func(key, value string) bool {
		return fn(value)
	}); err != nil {
		return err
	}

	if err := m.price.AddWalk(registry, field); err != nil {
		return err
	}

	return nil
}

func (m *metadataTx) Rollback() error {
	err := m.tx.Rollback()
	if err != nil {
		return err
	}

	m.closeTx()
	return nil
}

func (m *metadataTx) Commit() error {
	err := m.tx.Commit()
	if err != nil {
		return err
	}

	m.closeTx()
	return nil
}

func (m *metadataTx) Price() int64 {
	return m.price.GetCurrentPrice()
}

func (m *metadataTx) Fill(name string, params map[string]interface{}) (types.RegistryModel, error) {
	r := model.GetRegistries()
	for _, registry := range r {
		model, ok := registry.(types.RegistryModel)
		if !ok {
			panic("registry must implementing RegistryModel interface")
		}

		if model.ModelName() == name {
			filled, err := model.CreateFromData(params)
			if err != nil {
				return nil, err
			}

			return filled, nil
		}
	}

	return nil, ErrWrongRegistry
}

func (m *metadataTx) closeTx() {
	m.tx = nil
}

func (m *metadataTx) prepareValue(registry *types.Registry, pkValue string, newValue interface{}) (string, string, error) {
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

func (m *metadataTx) formatKey(reg *types.Registry, pk string) (string, error) {
	if reg.Name == "ecosystem" {
		return fmt.Sprintf("%s.%s", reg.Name, pk), nil
	}

	if reg.Ecosystem == nil {
		return "", ErrWrongRegistry
	}

	return fmt.Sprintf(keyConvention, reg.Name, reg.Ecosystem.Name, pk), nil
}

type metadataStorage struct {
	db      kv.Database
	indexer registryIndexer

	rollback bool
	pricing  bool
}

func NewMetadataStorage(db kv.Database, indexes []types.Index, rollback bool, pricing bool) (types.MetadataRegistryStorage, error) {
	ms := &metadataStorage{
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

func (m *metadataStorage) Begin() types.MetadataRegistryReaderWriter {
	databaseTx := m.db.Begin(true)
	tx := &metadataTx{tx: databaseTx, indexer: m.indexer}

	if m.rollback {
		tx.rollback = &metadataRollback{tx: databaseTx, counter: counter{txCounter: make(map[string]uint64)}}
	}

	if m.pricing {
		tx.price = priceCounter{tx: databaseTx, indexer: m.indexer}
	}

	return tx
}

func (m *metadataStorage) Walk(registry *types.Registry, field string, fn func(value string) bool) error {
	tx := &metadataTx{tx: m.db.Begin(false)}
	defer tx.Rollback()
	return tx.Walk(registry, field, fn)
}

func (m *metadataStorage) Get(registry *types.Registry, pkValue string, out interface{}) error {
	tx := &metadataTx{tx: m.db.Begin(false)}
	defer tx.Rollback()
	return tx.Get(registry, pkValue, out)
}

func (m *metadataStorage) Rollback(block []byte) error {
	if !m.rollback {
		return ErrRollbackDisabled
	}

	databaseTx := m.db.Begin(true)
	rollback := &metadataRollback{tx: databaseTx, counter: counter{txCounter: make(map[string]uint64)}}

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
