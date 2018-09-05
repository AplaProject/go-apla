package registry

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/yddmat/memdb"
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

		reg, err := NewMetadataStorage(db, []types.Index{
			{Field: "name", Registry: &types.Registry{Name: "ecosystem", Type: types.RegistryTypePrimary}},
		}, false)
		require.Nil(t, err)

		metadataTx := reg.Begin()
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

func BenchmarkMetadataTx(b *testing.B) {
	rollbacks := false
	persist := false
	db, err := newKvDB(persist)
	require.Nil(b, err)
	fmt.Println("Database persistence:", persist)
	fmt.Println("Rollbacks:", persist)

	storage, err := NewMetadataStorage(db, []types.Index{
		{
			Field:    "name",
			Registry: &types.Registry{Name: "ecosystem"},
			SortFn: func(a, b string) bool {
				return gjson.Get(a, "name").Less(gjson.Get(b, "name"), false)
			},
		},
	}, rollbacks)
	require.Nil(b, err)

	metadataTx := storage.Begin()

	ecosystems := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "k"}
	for _, ecosystem := range ecosystems {
		err = metadataTx.Insert(nil, &types.Registry{
			Name: "ecosystem",
		}, ecosystem, model.Ecosystem{Name: ecosystem})
		require.Nil(b, err)
	}
	count := 100000

	insertStart := time.Now()
	for i := 0; i < count; i++ {
		ecosystem := ecosystems[rand.Intn(9)]
		reg := types.Registry{
			Name:      "key",
			Ecosystem: &types.Ecosystem{Name: ecosystem},
		}

		id := rand.Int63()
		err := metadataTx.Insert(
			nil,
			&reg,
			strconv.FormatInt(id, 10),
			model.KeySchema{
				ID:        id,
				PublicKey: make([]byte, 64),
				Amount:    rand.Int63(),
			},
		)
		require.Nil(b, err)
	}
	require.Nil(b, metadataTx.Commit())
	fmt.Println("Inserted", count, "keys:", time.Since(insertStart))

	metadataTx = storage.Begin()

	indexStart := time.Now()
	require.Nil(b, metadataTx.AddIndex(types.Index{
		Registry: &types.Registry{Name: "key"},
		Field:    "amount",
		SortFn: func(a, b string) bool {
			return gjson.Get(b, "amount").Less(gjson.Get(a, "amount"), false)
		}},
	))
	fmt.Println("Creating and fill 'amount' index by", count, "keys:", time.Since(indexStart))

	require.Nil(b, metadataTx.Walk(&types.Registry{
		Name:      "key",
		Ecosystem: &types.Ecosystem{Name: "e"},
	}, "amount", func(jsonRow string) bool {
		k := model.KeySchema{}
		require.Nil(b, json.Unmarshal([]byte(jsonRow), &k))
		return false
	}))
}
