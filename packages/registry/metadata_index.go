package registry

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
)

var ErrEcosystemIndexNotFound = errors.New("main index (ecosystem) not found")
var ErrCreateIndexes = errors.New("cant create indexes")

type indexer struct {
	tx kv.Transaction
}

func (i *indexer) initPrimaryIndex(indexes []types.Index) error {
	var found bool
	for _, index := range indexes {
		if index.Registry == nil {
			return ErrWrongRegistry
		}

		if index.Registry.Type == types.RegistryTypePrimary {
			err := i.createIndex(index)
			if err != nil {
				return errors.New("cant init primary index")
			}

			found = true
			break
		}
	}

	if !found {
		return ErrEcosystemIndexNotFound
	}

	return nil
}

func (i *indexer) AddIndexes(init bool, indexes ...types.Index) error {
	if init {
		if err := i.initPrimaryIndex(indexes); err != nil {
			return err
		}
	}

	ecosystems := make([]string, 0)
	var err error
	i.tx.Ascend(i.formatIndexName(&types.Registry{Name: "ecosystem", Type: types.RegistryTypePrimary}, "name"), func(key, value string) bool {
		e := &model.Ecosystem{}
		if err = json.Unmarshal([]byte(value), e); err != nil {
			return false
		}
		ecosystems = append(ecosystems, e.Name)
		return true
	})

	if err != nil {
		return errors.Wrapf(err, "retrieving all primary entities")
	}

	prepeared := make([]types.Index, 0)
	for _, index := range indexes {
		if index.Registry == nil {
			return ErrWrongRegistry
		}

		if index.Registry.Type != types.RegistryTypeDefault {
			continue
		}

		for _, ecosystem := range ecosystems {
			if index.Field == "" {
				return errors.New("unknown field")
			}

			r := *index.Registry
			r.Ecosystem = &types.Ecosystem{Name: ecosystem}
			prepeared = append(prepeared, types.Index{
				Registry: &r,
				Field:    index.Field,
				SortFn:   index.SortFn,
			})
		}
	}

	for _, p := range prepeared {
		fmt.Println(p.Registry.Ecosystem.Name)
	}

	if err := i.createIndex(prepeared...); err != nil {
		return ErrCreateIndexes
	}

	return nil
}

func (i *indexer) createIndex(indexes ...types.Index) error {
	kvIndexes := make([]kv.Index, 0)
	for _, index := range indexes {
		kvIndexes = append(kvIndexes, kv.Index{
			Name:    i.formatIndexName(index.Registry, index.Field),
			SortFn:  index.SortFn,
			Pattern: i.formatIndexPattern(index.Registry),
		})
	}

	return i.tx.AddIndex(kvIndexes...)
}

func (i *indexer) formatIndexPattern(reg *types.Registry) string {
	switch reg.Type {
	case types.RegistryTypeDefault:
		return fmt.Sprintf("%s.%s.*", reg.Name, reg.Ecosystem.Name)
	case types.RegistryTypePrimary:
		return fmt.Sprintf("%s.*", reg.Name)
	}

	panic("unknown registry")
}

func (i *indexer) formatIndexName(reg *types.Registry, field string) string {
	switch reg.Type {
	case types.RegistryTypeDefault:
		return fmt.Sprintf("%s.%s.%s", reg.Name, field, reg.Ecosystem.Name)
	case types.RegistryTypePrimary:
		return fmt.Sprintf("%s.%s", reg.Name, field)
	}

	panic("unknown registry")
}
