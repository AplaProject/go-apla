// Metadata storage integration tests
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"math/rand"
	"time"

	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/storage/kv"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/GenesisKernel/memdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
)

type testModel struct {
	Id     int
	Field  string
	Field2 []byte
}

func newKvDB(persist bool) (kv.Database, error) {
	if persist {
		if err := os.Remove("test.db"); err != nil {
			return nil, err
		}
	}
	db, err := memdb.OpenDB("test.db", persist)
	if err != nil {
		return nil, err
	}

	return &kv.DatabaseAdapter{Database: *db}, nil
}

func TestMetadataTx_RW(t *testing.T) {
	cases := []struct {
		testname string

		registry types.Registry
		pkValue  string
		value    interface{}

		expJson string
		err     bool
	}{
		{
			testname: "insert-good",
			registry: types.Registry{
				Name:      "key",
				Ecosystem: &types.Ecosystem{Name: "abc"},
			},
			pkValue: "1",
			value: testModel{
				Id:     1,
				Field:  "testfield",
				Field2: make([]byte, 10),
			},

			err: false,
		},

		{
			testname: "insert-bad-1",
			registry: types.Registry{
				Name:      "key",
				Ecosystem: &types.Ecosystem{Name: "abc"},
			},
			pkValue: "1",
			value:   make(chan int),

			err: true,
		},
	}

	for _, c := range cases {
		db, err := newKvDB(false)
		require.Nil(t, err)

		reg, err := NewStorage(db, []types.Index{
			{
				Registry: &types.Registry{Name: "key"},
				Name:     "amount",
				SortFn: func(a, b string) bool {
					return gjson.Get(b, "amount").Less(gjson.Get(a, "amount"), false)
				},
			},
			{
				Name:     "name",
				Registry: &types.Registry{Name: "ecosystem", Type: types.RegistryTypePrimary},
				SortFn: func(a, b string) bool {
					return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
				},
			},
		}, false, true)
		require.Nil(t, err)

		metadataTx := reg.Begin(nil)
		require.Nil(t, err, c.testname)

		err = metadataTx.Insert(nil, &types.Registry{Name: model.Ecosystem{}.ModelName()}, "abc", model.Ecosystem{
			Name: "abc",
		})
		require.Nil(t, err, c.testname)

		err = metadataTx.Insert(nil, &c.registry, c.pkValue, c.value)
		if c.err {
			assert.Error(t, err)
			continue
		}

		assert.Nil(t, err)

		saved := testModel{}
		err = metadataTx.Get(&c.registry, c.pkValue, &saved)
		require.Nil(t, err)

		assert.Equal(t, c.value, saved, c.testname)
	}
}

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
			Name:     "amount",
			Registry: &types.Registry{Name: "key"},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "amount").Less(gjson.Get(b, "amount"), false)
			},
		},
		{
			Name:     "name",
			Registry: &types.Registry{Name: "role"},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "role").Less(gjson.Get(b, "role"), false)
			},
		},
		{
			Name:     "name",
			Registry: &types.Registry{Name: "ecosystem", Type: types.RegistryTypePrimary},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
			},
		},
	}

	msrw, err := NewStorage(db, idxs, false, true)
	require.Nil(t, err)

	mtx := msrw.Begin(nil)

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
		Name:      model.Ecosystem{}.ModelName(),
		Ecosystem: &types.Ecosystem{Name: "ecosystem"},
	}, "ddd", model.Ecosystem{Name: "ddd"}))

	dddRoles := make([]string, 0)
	require.Nil(t, mtx.Walk(&types.Registry{Name: "role", Ecosystem: &types.Ecosystem{Name: "ddd"}}, "name", func(jsonRow string) bool {
		dddRoles = append(dddRoles, jsonRow)
		return true
	}))

	assert.Equal(t, []string{"{\"name\":\"dddadmin\"}"}, dddRoles)
}

func TestMetadataMultipleIndex(t *testing.T) {
	db, err := newKvDB(false)
	require.Nil(t, err)

	tx := db.Begin(true)
	cases := []struct {
		key string
		model.KeySchema
	}{
		{key: "keys.aaa.1", KeySchema: model.KeySchema{ID: 1, Amount: "10"}},
		{key: "keys.aaa.2", KeySchema: model.KeySchema{ID: 2, Amount: "20"}},
		{key: "keys.aaa.3", KeySchema: model.KeySchema{ID: 3, Amount: "19"}},
		{key: "keys.aaa.4", KeySchema: model.KeySchema{ID: 4, Amount: "20", Blocked: true}},
		{key: "keys.aaa.5", KeySchema: model.KeySchema{ID: 5, Amount: "10", Blocked: true}},
		{key: "keys.bbb.6", KeySchema: model.KeySchema{ID: 6, Amount: "10", Blocked: true}},
	}

	require.Nil(t, tx.Set("ecosystems.aaa", "{\"name\":\"aaa\"}"))
	require.Nil(t, tx.Set("ecosystems.bbb", "{\"name\":\"bbb\"}"))

	for _, c := range cases {
		j, err := json.Marshal(c.KeySchema)
		require.Nil(t, err)
		require.Nil(t, tx.Set(c.key, string(j)))
	}

	require.Nil(t, tx.Commit())

	idxs := []types.Index{
		{
			Name:     "amount_blocked",
			Registry: &types.Registry{Name: model.KeySchema{}.ModelName()},
			SortFn:   memdb.СompositeIndex(buntdb.IndexJSON("blocked"), buntdb.IndexJSON("amount")),
		},
		{
			Name:     "amount",
			Registry: &types.Registry{Name: model.KeySchema{}.ModelName()},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "amount").Less(gjson.Get(b, "amount"), false)
			},
		},
		{
			Name:     "name",
			Registry: &types.Registry{Name: model.Ecosystem{}.ModelName(), Type: types.RegistryTypePrimary},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
			},
		},
	}

	msrw, err := NewStorage(db, idxs, false, true)
	require.Nil(t, err)

	mtx := msrw.Begin(nil)
	got := make([]model.KeySchema, 0)
	err = mtx.Walk(&types.Registry{Name: model.KeySchema{}.ModelName(), Ecosystem: &types.Ecosystem{Name: "aaa"}}, "amount_blocked", func(jsonRow string) bool {
		k := model.KeySchema{}
		require.Nil(t, json.Unmarshal([]byte(jsonRow), &k))
		got = append(got, k)
		return true
	})
	require.Nil(t, err)

	assert.Equal(t, []model.KeySchema{
		cases[0].KeySchema, cases[2].KeySchema, cases[1].KeySchema, cases[4].KeySchema, cases[3].KeySchema,
	}, got)
}

func TestRollbackSaveRollback(t *testing.T) {
	mDb, err := memdb.OpenDB("", false)
	require.Nil(t, err)
	db := kv.DatabaseAdapter{Database: *mDb}

	os.Remove("testrollback")
	ldb, err := leveldb.OpenFile("testrollback", nil)
	require.Nil(t, err)

	ltx, err := ldb.OpenTransaction()
	require.Nil(t, err)

	dbTx := db.Begin(true)
	mr := rollback{tx: dbTx, ltx: ltx, counter: counter{txCounter: make(map[string]uint64)}}

	registry := &types.Registry{
		Name:      "key",
		Ecosystem: &types.Ecosystem{Name: "aaa"},
	}

	block := []byte("123")

	mtx := tx{}
	for key := range make([]int, 20) {
		// Emulating new value in database
		mtx.formatKey(registry, strconv.Itoa(key))
		formatted, err := mtx.formatKey(registry, strconv.Itoa(key))
		require.Nil(t, err)
		dbTx.Set(formatted, "{\"result\":\"blah\"")

		tx := []byte(strconv.Itoa(key))
		tx = append(tx, []byte("blah")...)

		value := teststruct{
			Key:    key,
			Value1: "stringvalue" + strconv.Itoa(key),
			Value2: make([]byte, 20),
		}

		jsonValue, err := json.Marshal(value)
		require.Nil(t, err)
		// Save "old" state of record
		require.Nil(t, mr.saveState(block, tx, registry, strconv.Itoa(key), string(jsonValue)))
	}
	require.Nil(t, dbTx.Commit())

	dbTx = db.Begin(false)

	// We need to check that all previous states was saved to db
	for key := range make([]int, 20) {
		tx := []byte(strconv.Itoa(key))
		tx = append(tx, []byte("blah")...)

		_, err := ltx.Get([]byte(fmt.Sprintf(writePrefix, string(block), key+1, string(tx))), nil)
		require.Nil(t, err)
	}
	require.Nil(t, dbTx.Commit())

	dbTx = db.Begin(true)
	require.Nil(t, err)

	mr = rollback{tx: dbTx, ltx: ltx, counter: counter{txCounter: make(map[string]uint64)}}
	require.Nil(t, mr.rollbackState(block))

	// We are checking that all values are now at the previous state
	for key := range make([]int, 20) {
		value, err := ltx.Get([]byte(fmt.Sprintf(keyConvention, registry.Name, registry.Ecosystem.Name, strconv.Itoa(key))), nil)
		require.Nil(t, err)

		got := teststruct{}
		json.Unmarshal([]byte(value), &got)
		require.Equal(t, teststruct{
			Key:    key,
			Value1: "stringvalue" + strconv.Itoa(key),
			Value2: make([]byte, 20),
		}, got)
	}
}

func BenchmarkMetadataTx(b *testing.B) {
	rollbacks := false
	persist := true
	db, err := newKvDB(persist)
	require.Nil(b, err)
	fmt.Println("Database persistence:", persist)
	fmt.Println("Rollbacks:", rollbacks)

	storage, err := NewStorage(db, []types.Index{
		{
			Registry: &types.Registry{Name: "keys"},
			Name:     "amount",
			SortFn: func(a, b string) bool {
				return gjson.Get(b, "amount").Less(gjson.Get(a, "amount"), false)
			},
		},
		{
			Name:     "name",
			Registry: &types.Registry{Name: "ecosystems", Type: types.RegistryTypePrimary},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
			},
		},
	}, rollbacks, false)
	require.Nil(b, err)

	metadataTx := storage.Begin(nil)

	ecosystems := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "k"}
	for _, ecosystem := range ecosystems {
		err = metadataTx.Insert(nil, &types.Registry{
			Name: "ecosystem",
		}, ecosystem, model.Ecosystem{Name: ecosystem})
		require.Nil(b, err)
	}
	count := 30000

	insertStart := time.Now()
	ids := make(map[int64]string, 0)
	for i := 0; i < count; i++ {
		ecosystem := ecosystems[rand.Intn(9)]
		reg := types.Registry{
			Name:      "keys",
			Ecosystem: &types.Ecosystem{Name: ecosystem},
		}

		id := rand.Int63()
		strId := strconv.FormatInt(id, 10)
		err := metadataTx.Insert(
			nil,
			&reg,
			strId,
			model.KeySchema{
				ID:        id,
				PublicKey: make([]byte, 64),
				Amount:    strconv.FormatInt(rand.Int63(), 10),
			},
		)

		ids[id] = ecosystem
		require.Nil(b, err)
	}
	require.Nil(b, metadataTx.Commit())
	fmt.Println("Inserted", count, "keys:", time.Since(insertStart))

	metadataTx = storage.Begin(nil)
	updStart := time.Now()
	for id, ecosys := range ids {
		metadataTx.Update(
			nil,
			&types.Registry{
				Name:      "keys",
				Ecosystem: &types.Ecosystem{Name: ecosys},
			},
			strconv.FormatInt(id, 10),
			model.KeySchema{
				ID:     id,
				Amount: "0",
			},
		)
	}

	metadataTx.Commit()
	fmt.Println("Updated", count, "keys:", time.Since(updStart))
}
