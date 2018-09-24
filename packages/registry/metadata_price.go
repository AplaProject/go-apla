package registry

import (
	"fmt"

	"sync"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
)

type operation int

const (
	Set operation = iota
	Update
	Walk
	Get
)

var price = map[operation]int64{
	Set:    1,
	Update: 1,
	Walk:   1,
	Get:    1,
}

var rowCoeff = map[operation]float64{
	Set:    0.0001,
	Update: 0.0001,
	Walk:   0.0001,
	Get:    0.0001,
}

type priceCounter struct {
	tx      kv.Transaction
	indexer registryIndexer

	price int64
	mu    sync.RWMutex
}

func (pc *priceCounter) Add(operation operation, registry *types.Registry) error {
	indexes := pc.indexer.getIndexes(registry)
	if len(indexes) == 0 {
		panic(fmt.Sprintf("registry %s doesn't have any indexes", registry.Name))
	}

	rows, err := pc.countRows(registry, indexes[0].Field)
	if err != nil {
		return err
	}

	pc.mu.Lock()
	pc.price += price[operation] + int64(rowCoeff[operation]*float64(rows))
	pc.mu.Unlock()
	return nil
}

func (pc *priceCounter) AddWalk(registry *types.Registry, field string) error {
	rows, err := pc.tx.Len(pc.indexer.formatIndexName(registry, field))
	if err != nil {
		return err
	}

	pc.mu.Lock()
	pc.price += price[Walk] + int64(rowCoeff[Walk]*float64(rows))
	pc.mu.Unlock()
	return nil
}

func (pc *priceCounter) GetCurrentPrice() int64 {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.price
}

func (pc *priceCounter) countRows(registry *types.Registry, field string) (int, error) {
	return pc.tx.Len(pc.indexer.formatIndexName(registry, field))
}
