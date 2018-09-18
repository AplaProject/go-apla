package registry

import (
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestMetadataIndex(t *testing.T) {
	db, err := newKvDB(false)
	require.Nil(t, err)

	tx := db.Begin(true)
	require.Nil(t, tx.Set("key.aaa.1", "{\"amount\":10}"))
	require.Nil(t, tx.Set("key.aaa.2", "{\"amount\":15}"))
	require.Nil(t, tx.Set("key.bbb.3", "{\"amount\":20}"))

	require.Nil(t, tx.Set("role.aaa.admin", "{\"name\":\"admin\"}"))
	require.Nil(t, tx.Set("role.ccc.admin", "{\"name\":\"admin\"}"))
	require.Nil(t, tx.Set("role.ccc.user", "{\"name\":\"user\"}"))
	require.Nil(t, tx.Set("role.ccc.moderator", "{\"name\":\"moderator\"}"))
	require.Nil(t, tx.Set("role.bbb.admin", "{\"name\":\"admin\"}"))
	require.Nil(t, tx.Set("role.ddd.admin", "{\"name\":\"dddadmin\"}"))

	require.Nil(t, tx.Set("ecosystem.aaa", "{\"name\":\"aaa\"}"))
	require.Nil(t, tx.Set("ecosystem.bbb", "{\"name\":\"bbb\"}"))
	require.Nil(t, tx.Set("ecosystem.ccc", "{\"name\":\"ccc\"}"))
	require.Nil(t, tx.Commit())

	idxs := []types.Index{
		{
			Field:    "amount",
			Registry: &types.Registry{Name: "key"},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "amount").Less(gjson.Get(b, "amount"), false)
			},
		},
		{
			Field:    "name",
			Registry: &types.Registry{Name: "role"},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "role").Less(gjson.Get(b, "role"), false)
			},
		},
		{
			Field:    "name",
			Registry: &types.Registry{Name: "ecosystem", Type: types.RegistryTypePrimary},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
			},
		},
	}

	msrw, err := NewMetadataStorage(db, idxs, false)
	require.Nil(t, err)

	mtx := msrw.Begin()

	keys := make(map[string][]string, 0)
	for _, ecosys := range []string{"aaa", "bbb", "ccc"} {
		require.Nil(t, mtx.Walk(&types.Registry{Name: "key", Ecosystem: &types.Ecosystem{Name: ecosys}}, "amount", func(jsonRow string) bool {
			keys[ecosys] = append(keys[ecosys], jsonRow)
			return true
		}))
	}
	assert.Equal(t, map[string][]string{
		"aaa": {"{\"amount\":10}", "{\"amount\":15}"},
		"bbb": {"{\"amount\":20}"},
	}, keys)

	roles := make(map[string][]string, 0)
	for _, ecosys := range []string{"aaa", "bbb", "ccc"} {
		require.Nil(t, mtx.Walk(&types.Registry{Name: "role", Ecosystem: &types.Ecosystem{Name: ecosys}}, "name", func(jsonRow string) bool {
			roles[ecosys] = append(roles[ecosys], jsonRow)
			return true
		}))
	}
	assert.Equal(t, map[string][]string{
		"aaa": {"{\"name\":\"admin\"}"},
		"bbb": {"{\"name\":\"admin\"}"},
		"ccc": {"{\"name\":\"admin\"}", "{\"name\":\"moderator\"}", "{\"name\":\"user\"}"},
	}, roles)

	require.Error(t, mtx.Walk(&types.Registry{Name: "role", Ecosystem: &types.Ecosystem{Name: "ddd"}}, "name", func(jsonRow string) bool {
		return true
	}))

	assert.Nil(t, mtx.Insert(nil, &types.Registry{
		Name:      "ecosystem",
		Ecosystem: &types.Ecosystem{Name: "ecosystem"},
	}, "ddd", model.Ecosystem{Name: "ddd"}))

	dddRoles := make([]string, 0)
	require.Nil(t, mtx.Walk(&types.Registry{Name: "role", Ecosystem: &types.Ecosystem{Name: "ddd"}}, "name", func(jsonRow string) bool {
		dddRoles = append(dddRoles, jsonRow)
		return true
	}))

	assert.Equal(t, []string{"{\"name\":\"dddadmin\"}"}, dddRoles)
}
