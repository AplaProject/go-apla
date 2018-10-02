package model

import "github.com/GenesisKernel/go-genesis/packages/types"

var registries = []types.RegistryModel{
	&KeySchema{},
	&Ecosystem{},
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

func IsMetaRegistry(name string) bool {
	for _, registry := range registries {
		r, ok := registry.(types.RegistryModel)
		if !ok {
			panic("converting registry to Namer interface")
		}

		if r.ModelName() == name {
			return true
		}
	}

	return false
}

func GetRegistries() []types.RegistryModel {
	return registries
}

func GetRegistry(name string) types.RegistryModel {
	for _, registry := range registries {
		if registry.ModelName() == name {
			return registry
		}
	}

	return nil
}
