package metadata

import (
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceCounter(t *testing.T) {
	txMock := &kv.MockTransaction{}
	idxer := &mockRegistryIndexer{}
	pc := priceCounter{tx: txMock, indexer: idxer}

	for i := 0; i < 10; i++ {
		reg := &types.Registry{Name: "key", Ecosystem: &types.Ecosystem{Name: "abc"}}
		idxer.On("getIndexes", reg).Return([]types.Index{
			{Field: "amount"},
		})
		idxer.On("formatIndexName", reg, "amount").Return("blah")
		txMock.On("Len", "blah").Return(10000, nil)
		require.Nil(t, pc.Add(Set, reg))
	}

	assert.Equal(t, int64(20), pc.GetCurrentPrice())
}

func TestPriceEmptyIndexes(t *testing.T) {
	txMock := &kv.MockTransaction{}
	indexer := &mockRegistryIndexer{}
	pc := priceCounter{tx: txMock, indexer: indexer}

	reg := &types.Registry{Name: "key", Ecosystem: &types.Ecosystem{Name: "abc"}}
	indexer.On("getIndexes", reg).Return(make([]types.Index, 0))

	assert.Panics(t, func() {
		pc.Add(Set, reg)
	}, "registry %s doesn't have any indexes", reg.Name)
}
