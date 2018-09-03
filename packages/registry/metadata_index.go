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
var ErrCreteIndexes = errors.New("cant create indexes")

type indexer struct {
	tx kv.Transaction
}

func (i *indexer) init(indexes []types.Index) error {
	var found bool
	for _, index := range indexes {
		if index.Registry == nil {
			return ErrWrongRegistry
		}

		if index.Registry.Name == "ecosystem" && index.Field == "name" {
			if err := i.tx.AddIndex(kv.Index{
				Name:    i.formatIndexName(index.Registry, index.Field),
				SortFn:  index.SortFn,
				Pattern: i.formatIndexPattern(index.Registry),
			}); err != nil {
				return errors.New("cant init ecosystem index")
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

func (i *indexer) addIndexes(indexes ...types.Index) error {
	ecosystems := make([]string, 0)
	i.tx.Ascend(i.formatIndexName(&types.Registry{Name: "ecosystem"}, "name"), func(key, value string) bool {
		e := &model.Ecosystem{}
		json.Unmarshal([]byte(value), e)
		ecosystems = append(ecosystems, e.Name)
		return true
	})

	kvIndexes := make([]kv.Index, 0)
	for _, index := range indexes {
		if index.Registry.Name == "ecosystem" && index.Field == "name" {
			continue
		}

		for _, ecosystem := range ecosystems {
			if index.Registry == nil {
				return ErrWrongRegistry
			}

			if index.Field == "" {
				return errors.New("unknown field")
			}

			index.Registry.Ecosystem = &types.Ecosystem{Name: ecosystem}
			kvIndexes = append(kvIndexes, kv.Index{
				Name:    i.formatIndexName(index.Registry, index.Field),
				SortFn:  index.SortFn,
				Pattern: i.formatIndexPattern(index.Registry),
			})
		}
	}

	if err := i.tx.AddIndex(kvIndexes...); err != nil {
		return ErrCreteIndexes
	}

	return nil
}

func (i *indexer) formatIndexPattern(reg *types.Registry) string {
	if reg.Name == "ecosystem" {
		return fmt.Sprintf("%s.*", reg.Name)
	}

	return fmt.Sprintf("%s.%s.*", reg.Name, reg.Ecosystem.Name)
}

func (i *indexer) formatIndexName(reg *types.Registry, field string) string {
	if reg.Name == "ecosystem" && field == "name" {
		return fmt.Sprintf("%s.%s", reg.Name, field)
	}

	return fmt.Sprintf("%s.%s.%s", reg.Name, field, reg.Ecosystem.Name)
}
