package model

import "github.com/GenesisKernel/go-genesis/packages/types"

var registries = []interface{}{
	KeySchema{}, Ecosystem{},
}

func GetIndexes() []types.Index {
	indexes := make([]types.Index, 0)
	for _, registry := range registries {
		indexer, ok := registry.(types.Indexer)
		if !ok {
			panic("converting registry to Indexer interface")
		}

		indexes = append(indexes, indexer.GetIndexes()...)
	}

	return indexes
}
