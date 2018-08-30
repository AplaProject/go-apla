package registry

import (
	"fmt"
	"sync"

	"encoding/json"

	"sort"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
	"github.com/tidwall/match"
)

const (
	writePrefix  = "rollback_tx.%s.%d.%s"
	searchPrefix = "rollback_tx.%s.%s.%s"
)

type state struct {
	Counter      uint64 `json:"c"`
	RegistryName string `json:"r"`
	Value        string `json:"v"`
	Key          string `json:"k"`
	Ecosystem    uint64 `json:"e"`
}

type metadataRollback struct {
	tx kv.Transaction

	txCounter map[string]uint64
	mu        sync.Mutex
}

func (mr *metadataRollback) saveState(block, tx []byte, registry *types.Registry, pk, value string) error {
	key := string(block)
	mr.mu.Lock()
	defer mr.mu.Unlock()

	counter := mr.txCounter[key]
	counter++

	s := state{Counter: counter, RegistryName: registry.Name, Key: pk, Ecosystem: registry.Ecosystem.ID, Value: value}
	jstate, err := json.Marshal(s)
	if err != nil {
		return err
	}

	kk := fmt.Sprintf(writePrefix, string(block), counter, string(tx))
	err = mr.tx.Set(kk, string(jstate))
	if err != nil {
		return err
	}

	mr.txCounter[key] = counter
	return nil
}

// rollbackState is rollback all block transactions
func (mr *metadataRollback) rollbackState(block []byte) error {
	txses := make([]state, 0)
	var err error
	mr.tx.Ascend("rollback", func(key, value string) bool {
		if match.Match(key, fmt.Sprintf(searchPrefix, string(block), "*", "*")) {
			state := state{}
			err = json.Unmarshal([]byte(value), &state)
			if err != nil {
				err = errors.Wrapf(err, "retrieving block transactions")
				return false
			}

			txses = append(txses, state)
		}

		return true
	})

	sort.Slice(txses, func(i, j int) bool {
		return txses[i].Counter > txses[j].Counter
	})

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
