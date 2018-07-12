package registry

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
	"github.com/tidwall/buntdb"
)

const keyConvention = "%d.%s.%s"

type MetadataService struct {
	customRegistry *kv.BuntDBStorage
}

func NewRegistryManager(customStorage *buntdb.DB) *MetadataService {
	return &MetadataService{
		customRegistry: kv.NewBuntDBStorage(customStorage),
	}
}

func (m *MetadataService) Insert(registry *types.Registry, pkValue string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return errors.Wrapf(err, "marshalling struct to json")
	}

	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	err = m.customRegistry.Insert(key, string(jsonValue))
	if err != nil {
		return errors.Wrapf(err, "inserting value %s to %s registry", value, registry.Name)
	}

	return nil
}

func (m *MetadataService) Get(registry *types.Registry, pkValue string, out interface{}) error {
	key := fmt.Sprintf(keyConvention, registry.Ecosystem.ID, registry.Name, pkValue)

	value, err := m.customRegistry.Get(key)
	if err != nil {
		return errors.Wrapf(err, "retrieving %s from service registry", key)
	}

	err = json.Unmarshal([]byte(value), out)
	if err != nil {
		return errors.Wrapf(err, "unmarshalling value %s to struct", value)
	}

	return nil
}
