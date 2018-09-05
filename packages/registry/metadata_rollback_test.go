package registry

import (
	"testing"

	"encoding/json"

	"fmt"

	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/yddmat/memdb"
)

type teststruct struct {
	Key    int
	Value1 string
	Value2 []byte
}

func TestMetadataRollbackSaveState(t *testing.T) {
	txMock := &kv.MockTransaction{}
	mr := metadataRollback{tx: txMock, counter: counter{txCounter: make(map[string]uint64)}}

	registry := &types.Registry{
		Name:      "keys",
		Ecosystem: &types.Ecosystem{Name: "aaa"},
	}

	block, tx := []byte("123"), []byte("321")

	s := state{Counter: 1, RegistryName: registry.Name, Ecosystem: registry.Ecosystem.Name, Key: "1"}
	jstate, err := json.Marshal(s)
	require.Nil(t, err)
	txMock.On("Set", fmt.Sprintf(writePrefix, string(block), 1, string(tx)), string(jstate)).Return(nil)
	require.Nil(t, mr.saveState(block, tx, registry, "1", ""))
	fmt.Println(mr.counter.txCounter[string(block)])
	require.Equal(t, mr.counter.txCounter[string(block)], uint64(1))

	structValue := teststruct{
		Key:    666,
		Value1: "stringvalue",
		Value2: make([]byte, 20),
	}
	jsonValue, err := json.Marshal(structValue)
	require.Nil(t, err)
	s = state{Counter: 2, RegistryName: registry.Name, Ecosystem: registry.Ecosystem.Name, Value: string(jsonValue), Key: "2"}
	jstate, err = json.Marshal(s)
	require.Nil(t, err)
	txMock.On("Set", fmt.Sprintf(writePrefix, string(block), 2, string(tx)), string(jstate)).Return(nil)
	require.Nil(t, mr.saveState(block, tx, registry, "2", string(jsonValue)))
	require.Equal(t, mr.counter.txCounter[string(block)], uint64(2))

	s = state{Counter: 3, RegistryName: registry.Name, Ecosystem: registry.Ecosystem.Name, Value: "", Key: "3"}
	jstate, err = json.Marshal(s)
	require.Nil(t, err)
	txMock.On("Set", fmt.Sprintf(writePrefix, string(block), 3, string(tx)), string(jstate)).Return(errors.New("testerr"))
	require.Error(t, mr.saveState(block, tx, registry, "3", ""))
	require.Equal(t, mr.counter.txCounter[string(block)], uint64(2))
}

func TestMetadataRollbackSaveRollback(t *testing.T) {
	mDb, err := memdb.OpenDB("", false)
	require.Nil(t, err)
	db := kv.DatabaseAdapter{Database: *mDb}

	dbTx := db.Begin(true)
	mr := metadataRollback{tx: dbTx, counter: counter{txCounter: make(map[string]uint64)}}

	registry := &types.Registry{
		Name:      "key",
		Ecosystem: &types.Ecosystem{Name: "aaa"},
	}

	block := []byte("123")

	mtx := metadataTx{}
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

		_, err := dbTx.Get(fmt.Sprintf(writePrefix, string(block), key+1, string(tx)))
		require.Nil(t, err)
	}
	require.Nil(t, dbTx.Commit())

	dbTx = db.Begin(true)
	require.Nil(t, err)

	dbTx.AddIndex(kv.Index{
		Name: "rollback_tx",
		SortFn: func(a, b string) bool {
			return true
		},
		Pattern: fmt.Sprintf(searchPrefix, "*", "*", "*"),
	})

	mr = metadataRollback{tx: dbTx, counter: counter{txCounter: make(map[string]uint64)}}
	require.Nil(t, mr.rollbackState(block))

	// We are checking that all values are now at the previous state
	for key := range make([]int, 20) {
		// Emulating new value in database
		value, err := dbTx.Get(fmt.Sprintf(keyConvention, registry.Name, registry.Ecosystem.Name, strconv.Itoa(key)))
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
