package metadata

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	writePrefix  = "rollback_tx.%s.%d.%s"
	searchPrefix = "rollback_tx.%s.%s.%s"
)

type state struct {
	Transaction  string `json:"t"`
	Counter      uint64 `json:"c"`
	RegistryName string `json:"r"`
	Value        string `json:"v"`
	Key          string `json:"k"`
	Ecosystem    string `json:"e"`
}

type counter struct {
	txCounter map[string]uint64
	mu        sync.Mutex
}

func (c *counter) increment(key string) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.txCounter[key]++
	return c.txCounter[key]
}

func (c *counter) decrement(key string) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.txCounter[key]--
	return c.txCounter[key]
}

type rollback struct {
	tx  kv.Transaction
	ltx *leveldb.Transaction

	counter counter
}

func (mr *rollback) saveState(block, tx []byte, registry *types.Registry, pk, value string) error {
	key := string(block)

	counter := mr.counter.increment(key)

	s := state{Transaction: string(tx), Counter: counter, RegistryName: registry.Name, Key: pk, Ecosystem: registry.Ecosystem.Name, Value: value}
	jstate, err := json.Marshal(s)
	if err != nil {
		mr.counter.decrement(key)
		return err
	}

	kk := fmt.Sprintf(writePrefix, string(block), counter, string(tx))
	err = mr.ltx.Put([]byte(kk), jstate, nil)
	if err != nil {
		mr.counter.decrement(key)
		return err
	}

	return nil
}

// rollbackState is rollback all block transactions
func (mr *rollback) rollbackState(block []byte) error {
	txses, err := mr.getBlockStates(block)
	if err != nil {
		return errors.Wrapf(err, "retrieving block states")
	}

	for _, tx := range txses {
		key := fmt.Sprintf(keyConvention, tx.RegistryName, tx.Ecosystem, tx.Key)

		// rollback inserted row
		if tx.Value == "" {
			err := mr.tx.Delete(key)
			if err != nil {
				return errors.Wrapf(err, "deleting old row")
			}
		} else {
			// rollback updated row
			err := mr.tx.Delete(key)
			if err != nil {
				return errors.Wrapf(err, "deleting old row")
			}

			err = mr.tx.Set(key, tx.Value)
			if err != nil {
				return errors.Wrapf(err, "setting old value")
			}
		}
	}

	return nil
}

func (mr *rollback) removeState(block []byte) error {
	txses, err := mr.getBlockStates(block)
	if err != nil {
		return errors.Wrapf(err, "retrieving block states")
	}

	for _, tx := range txses {
		key := fmt.Sprintf(writePrefix, string(block), tx.Counter, tx.Transaction)
		if err := mr.tx.Delete(key); err != nil {
			return errors.Wrapf(err, "removing block state %s", key)
		}
	}

	return nil
}

func (mr *rollback) getBlockStates(block []byte) ([]state, error) {
	txses := make([]state, 0)
	var err error

	iter := mr.ltx.NewIterator(util.BytesPrefix([]byte(fmt.Sprintf("rollback_tx.%s", string(block)))), nil)
	for iter.Next() {
		state := state{}
		err = json.Unmarshal([]byte(iter.Value()), &state)
		if err != nil {
			return nil, errors.Wrapf(err, "retrieving block transactions")
		}

		txses = append(txses, state)
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return txses, errors.Wrapf(err, "retrieving block transactions")
	}

	sort.Slice(txses, func(i, j int) bool {
		return txses[i].Counter > txses[j].Counter
	})

	return txses, nil
}
