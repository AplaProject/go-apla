package registry_test

import (
	"testing"

	"os"

	"path/filepath"

	"math/rand"

	"strconv"

	"time"

	"fmt"

	"encoding/json"

	"github.com/GenesisKernel/go-genesis/packages/registry"
	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testModel struct {
	Id     int
	Field  string
	Field2 []byte
}

func newBadger() (kv.Database, error) {
	path := filepath.Join(os.TempDir(), "badger")

	err := os.RemoveAll(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	opts := badger.DefaultOptions
	opts.Dir, opts.ValueDir = path, path

	return badger.Open(opts)
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
				Ecosystem: &types.Ecosystem{ID: 1},
				Type:      types.RegistryTypeMetadata,
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
				Ecosystem: &types.Ecosystem{ID: 1},
				Type:      types.RegistryTypeMetadata,
			},
			pkValue: "1",
			value:   make(chan int),

			err: true,
		},
	}

	for _, c := range cases {
		db, err := newBadger()
		require.Nil(t, err)

		reg := registry.NewMetadataStorage(db)
		metadataTx := reg.Begin()
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

func TestMetadataTx_1millionKeys(t *testing.T) {
	db, err := newBadger()
	require.Nil(t, err)

	storage := registry.NewMetadataStorage(db)
	metadataTx := storage.Begin()

	type key struct {
		ID        int64
		PublicKey []byte
		Amount    int64
		Deleted   bool
		Blocked   bool
	}

	reg := types.Registry{
		Name:      "key",
		Ecosystem: &types.Ecosystem{ID: 1},
		Type:      types.RegistryTypeMetadata,
	}

	insertStart := time.Now()
	for i := 0; i < 1000000; i++ {
		id := rand.Int63()
		err := metadataTx.Insert(
			&reg,
			strconv.FormatInt(id, 10),
			key{
				ID:        id,
				PublicKey: make([]byte, 64),
				Amount:    rand.Int63(),
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
	fmt.Println("Inserted 1 keys mln in", time.Since(insertStart).Seconds())

	readonlyTx := storage.Begin()
	walkingStart := time.Now()
	var topAmount int64
	require.Nil(t, readonlyTx.Walk(&reg, "", func(jsonRow string) bool {
		k := key{}
		require.Nil(t, json.Unmarshal([]byte(jsonRow), &k))
		if topAmount < k.Amount {
			topAmount = k.Amount
		}
		return true
	}))
	fmt.Println("Finded top amount of 1 mln keys", "in", time.Since(walkingStart))

	secondWriting := time.Now()
	writeTx := storage.Begin()
	// Insert 10 more values
	for i := -10; i < 0; i++ {
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
	fmt.Println("Inserted 10 keys to 1 mln in", time.Since(secondWriting))

}
