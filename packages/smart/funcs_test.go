package smart

import (
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/registry"
	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	"github.com/GenesisKernel/memdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func getMetaStorage(t *testing.T) types.MetadataRegistryStorage {
	db, err := memdb.OpenDB("", false)
	require.Nil(t, err)
	storage, err := registry.NewMetadataStorage(&kv.DatabaseAdapter{Database: *db}, []types.Index{
		{
			Registry: &types.Registry{Name: model.KeySchema{}.ModelName()},
			Name:     "amount",
			SortFn: func(a, b string) bool {
				return gjson.Get(b, "amount").Less(gjson.Get(a, "amount"), false)
			},
		},
		{
			Name:     "name",
			Registry: &types.Registry{Name: model.Ecosystem{}.ModelName(), Type: types.RegistryTypePrimary},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
			},
		},
	}, false, true)
	require.Nil(t, err)
	return storage
}

func TestDBUpdateExt2(t *testing.T) {
	storage := getMetaStorage(t)
	reg := &types.Registry{Name: model.KeySchema{}.ModelName(), Ecosystem: &types.Ecosystem{Name: "1"}}
	rw := storage.Begin()
	require.Nil(t, rw.Insert(nil, &types.Registry{Ecosystem: &types.Ecosystem{Name: "1"}, Type: types.RegistryTypePrimary}, "1", model.Ecosystem{Name: "1"}))
	require.Nil(t, rw.Insert(nil, reg, "1", model.KeySchema{ID: 1, PublicKey: []byte("pub"), Amount: "100"}))
	require.Nil(t, rw.Commit())
	rw = storage.Begin()

	sc := &SmartContract{
		MetaDb:  rw,
		TxSmart: tx.SmartContract{Header: tx.Header{EcosystemID: 1}},
	}

	_, err := DBUpdateExt(
		sc,
		model.KeySchema{}.ModelName(),
		"",
		"1",
		map[string]interface{}{
			"amount": "95",
		},
	)

	require.Nil(t, err)

	got := model.KeySchema{}
	assert.Nil(t, rw.Get(reg, "1", &got))
	assert.Equal(t, model.KeySchema{ID: 1, PublicKey: []byte("pub"), Amount: "95"}, got)
}

func TestDBInsert(t *testing.T) {
	storage := getMetaStorage(t)
	reg := &types.Registry{Name: model.KeySchema{}.ModelName(), Ecosystem: &types.Ecosystem{Name: "1"}}
	rw := storage.Begin()

	sc := &SmartContract{
		MetaDb:  rw,
		TxSmart: tx.SmartContract{Header: tx.Header{EcosystemID: 1}},
	}

	_, _, err := DBInsert(sc,
		model.KeySchema{}.ModelName(),
		map[string]interface{}{
			"id":        15,
			"publickey": []byte("pub"),
			"amount":    "95",
		},
	)

	require.Nil(t, err)

	got := model.KeySchema{}
	assert.Nil(t, rw.Get(reg, "15", &got))
	assert.Equal(t, model.KeySchema{ID: 15, PublicKey: []byte("pub"), Amount: "95"}, got)
}
