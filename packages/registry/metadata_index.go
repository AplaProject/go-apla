package registry

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
)

var ErrPrimaryRegistryNotFound = errors.New("primary registry not found")
var ErrCreateIndexes = errors.New("cant create indexes")

type indexer struct {
	indexes      []types.Index
	primaryIndex string
}

func newIndexer(indexes []types.Index) *indexer {
	return &indexer{indexes: indexes}
}

func (i *indexer) init(tx kv.Transaction) error {
	var found bool
	for _, index := range i.indexes {
		if index.Registry == nil {
			return ErrWrongRegistry
		}

		if index.Registry.Type == types.RegistryTypePrimary {
			i.primaryIndex = i.formatIndexName(index.Registry, index.Field)

			err := i.writeIndex(tx, index)
			if err != nil {
				return errors.New("cant init primary index")
			}

			found = true
			break
		}
	}

	if !found {
		return ErrPrimaryRegistryNotFound
	}

	primaryValues := make([]string, 0)
	var err error
	tx.Ascend(i.primaryIndex, func(key, value string) bool {
		e := &model.Ecosystem{}
		if err = json.Unmarshal([]byte(value), e); err != nil {
			return false
		}
		primaryValues = append(primaryValues, e.Name)
		return true
	})

	if err != nil {
		return errors.Wrapf(err, "retrieving all primary entities")
	}

	newIndexes, err := i.makeIndexesForValues(primaryValues...)
	if err != nil {
		return err
	}

	if err := i.writeIndex(tx, newIndexes...); err != nil {
		return ErrCreateIndexes
	}

	return nil
}

func (i *indexer) makeIndexesForValues(primaryValues ...string) ([]types.Index, error) {
	prepeared := make([]types.Index, 0)
	for _, index := range i.indexes {
		if index.Registry == nil {
			return nil, ErrWrongRegistry
		}

		if index.Registry.Type != types.RegistryTypeDefault {
			continue
		}

		for _, value := range primaryValues {
			if index.Field == "" {
				return nil, errors.New("unknown field")
			}

			r := *index.Registry
			r.Ecosystem = &types.Ecosystem{Name: value}
			prepeared = append(prepeared, types.Index{
				Registry: &r,
				Field:    index.Field,
				SortFn:   index.SortFn,
			})
		}
	}

	return prepeared, nil
}

func (i *indexer) AddPrimaryValue(tx kv.Transaction, value string) error {
	newIndexes, err := i.makeIndexesForValues(value)
	if err != nil {
		return err
	}

	if err := i.writeIndex(tx, newIndexes...); err != nil {
		return ErrCreateIndexes
	}

	return nil
}

func (i *indexer) RemovePrimaryValue(tx kv.Transaction, value string) error {
	indexes, err := i.makeIndexesForValues(value)
	if err != nil {
		return err
	}

	for _, index := range indexes {
		if err := tx.RemoveIndex(i.formatIndexName(index.Registry, index.Field)); err != nil {
			return err
		}
	}

	return nil
}

func (i *indexer) writeIndex(tx kv.Transaction, indexes ...types.Index) error {
	kvIndexes := make([]kv.Index, 0)
	for _, index := range indexes {
		kvIndexes = append(kvIndexes, kv.Index{
			Name:    i.formatIndexName(index.Registry, index.Field),
			SortFn:  index.SortFn,
			Pattern: i.formatIndexPattern(index.Registry),
		})
	}

	return tx.AddIndex(kvIndexes...)
}

func (i indexer) formatIndexPattern(reg *types.Registry) string {
	switch reg.Type {
	case types.RegistryTypeDefault:
		return fmt.Sprintf("%s.%s.*", reg.Name, reg.Ecosystem.Name)
	case types.RegistryTypePrimary:
		return fmt.Sprintf("%s.*", reg.Name)
	}

	panic("unknown registry")
}

func (i indexer) formatIndexName(reg *types.Registry, field string) string {
	switch reg.Type {
	case types.RegistryTypeDefault:
		return fmt.Sprintf("%s.%s.%s", reg.Name, field, reg.Ecosystem.Name)
	case types.RegistryTypePrimary:
		return fmt.Sprintf("%s.%s", reg.Name, field)
	}

	panic("unknown registry")
}
