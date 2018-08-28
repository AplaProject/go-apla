package registry_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/registry"
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
				Name: "key",
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
				Name: "key",
			},
			pkValue: "1",
			value:   make(chan int),

			err: true,
		},
	}

	for _, c := range cases {
		db, err := newKvDB(false)
		require.Nil(t, err)

		reg, err := registry.NewMetadataStorage(db, nil, true)
		require.Nil(t, err)

		metadataTx := reg.Begin()
		metadataTx.SetBlockHash([]byte("123"))
		metadataTx.SetTxHash([]byte("321"))
		require.Nil(t, err, c.testname)

		err = metadataTx.Insert(&c.registry, c.pkValue, c.value)
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

func TestMetadataTx_benchmark(t *testing.T) {
	rollbacks := false
	persist := false
	db, err := newKvDB(persist)
	require.Nil(t, err)
	fmt.Println("Database persistence:", persist)
	fmt.Println("Rollbacks:", persist)

	storage, err := registry.NewMetadataStorage(db, nil, rollbacks)
	require.Nil(t, err)

	metadataTx := storage.Begin()
	type key struct {
		ID        int64
		PublicKey []byte
		Amount    int64
		Deleted   bool
		Blocked   bool
		Ecosystem int
	}

	reg := types.Registry{
		Name: "key",
	}

	count := 100000

	insertStart := time.Now()
	for i := 0; i < count; i++ {
		id := rand.Int63()
		err := metadataTx.Insert(
			&reg,
			strconv.FormatInt(id, 10),
			key{
				ID:        id,
				PublicKey: make([]byte, 64),
				Amount:    rand.Int63(),
				Ecosystem: rand.Intn(10000),
			},
		)

		if err != nil {
			metadataTx.Commit()
			metadataTx = storage.Begin()
			err = nil
		}

		require.Nil(t, err)
	}
	require.Nil(t, metadataTx.Commit())
	fmt.Println("Inserted", count, "keys:", time.Since(insertStart))

	indexStart := time.Now()
	metadataTx = storage.Begin()
	metadataTx.AddIndex(types.Index{Name: "test", Registry: &types.Registry{Name: "key"}, SortFn: func(a, b string) bool {
		return gjson.Get(a, "amount").Less(gjson.Get(b, "amount"), false)
	}})
	require.Nil(t, metadataTx.Commit())
	fmt.Println("Creating and fill 'amount' index by", count, "keys:", time.Since(indexStart))

	readonlyTx := storage.Reader()
	var topAmount int64
	ecosystem := 666
	ecosystems := make(map[int]struct{})
	walkingStart := time.Now()
	require.Nil(t, readonlyTx.Walk(&reg, "test", func(jsonRow string) bool {
		k := key{}
		require.Nil(t, json.Unmarshal([]byte(jsonRow), &k))
		if k.Ecosystem == ecosystem && topAmount < k.Amount {
			topAmount = k.Amount
		}

		ecosystems[k.Ecosystem] = struct{}{}
		return true
	}))

	fmt.Println("Finded top amount of", count, "keys (", len(ecosystems), "ecosystems ):", time.Since(walkingStart))

	secondWriting := time.Now()
	writeTx := storage.Begin()
	writeTx.SetBlockHash([]byte("123"))
	writeTx.SetTxHash([]byte("321"))
	// Insert more values
	for i := -count; i < 0; i++ {
		id := rand.Int63()
		err := writeTx.Insert(
			&reg,
			strconv.FormatInt(id, 10),
			key{
				ID:        id,
				PublicKey: make([]byte, 64),
				Amount:    rand.Int63(),
			},
		)

		require.Nil(t, err)
	}
	require.Nil(t, writeTx.Commit())
	fmt.Println("Inserted", count, "more keys ( with indexes ):", time.Since(secondWriting))
}
