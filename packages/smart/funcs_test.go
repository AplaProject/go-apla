package smart

import (
	"testing"

	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/registry/metadata"
	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	"github.com/GenesisKernel/memdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/buntdb"
)

func getMetaStorage(t *testing.T) types.MetadataRegistryStorage {
	db, err := memdb.OpenDB("", false)
	require.Nil(t, err)
	storage, err := metadata.NewStorage(&kv.DatabaseAdapter{Database: *db}, []types.Index{
		{
			Registry: &types.Registry{Name: model.KeySchema{}.ModelName()},
			Name:     "amount",
			SortFn:   buntdb.IndexJSON("amount"),
		},
		{
			Name:     "name",
			Registry: &types.Registry{Name: model.Ecosystem{}.ModelName(), Type: types.RegistryTypePrimary},
			SortFn:   buntdb.IndexJSON("name"),
		},
	}, false, false)
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

func TestMetadbSelectWithID(t *testing.T) {
	storage := getMetaStorage(t)
	reg := &types.Registry{Name: model.KeySchema{}.ModelName(), Ecosystem: &types.Ecosystem{Name: "1"}}
	rw := storage.Begin()

	cases := []model.KeySchema{
		{ID: 1, Amount: "100"},
		{ID: 2, Amount: "200"},
		{ID: 3, Amount: "300"},
		{ID: 4, Amount: "400"},
		{ID: 5, Amount: "500"},
	}

	for _, c := range cases {
		require.Nil(t, rw.Insert(nil, reg, strconv.FormatInt(c.ID, 10), c))
	}

	sc := &SmartContract{
		MetaDb:  rw,
		TxSmart: tx.SmartContract{Header: tx.Header{EcosystemID: 1}},
	}

	_, data, err := metadbSelect(sc, model.KeySchema{}.ModelName(), []string{"amount", "blocked"}, 4, []interface{}{}, 0, 2, 1, nil)

	require.Nil(t, err)
	require.Equal(t, []interface{}{map[string]interface{}{
		"amount":  "400",
		"blocked": false,
	}}, data)
}

//func TestMetadbSelectWhere(t *testing.T) {
//	storage := getMetaStorage(t)
//	reg := &types.Registry{Name: model.KeySchema{}.ModelName(), Ecosystem: &types.Ecosystem{Name: "1"}}
//	rw := storage.Begin()
//
//	cases := []model.KeySchema{
//		{ID: 1, Amount: "101", Blocked: true},
//		{ID: 2, Amount: "102", Blocked: true},
//		{ID: 3, Amount: "103", Blocked: true},
//		{ID: 4, Amount: "400"},
//		{ID: 5, Amount: "500"},
//		{ID: 6, Amount: "600"},
//	}
//
//	for _, c := range cases {
//		require.Nil(t, rw.Insert(nil, reg, strconv.FormatInt(c.ID, 10), c))
//	}
//
//	sc := &SmartContract{
//		MetaDb:  rw,
//		TxSmart: tx.SmartContract{Header: tx.Header{EcosystemID: 1}},
//	}
//
//	_, data, err := metadbSelect(sc, model.KeySchema{}.ModelName(), []string{"id", "amount", "blocked"}, 0, []interface{}{}, 0, 0, 1, map[string]interface{}{
//		"id": "{\"$eq\": 4}",
//	})
//}
