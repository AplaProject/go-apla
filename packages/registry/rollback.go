package registry

import (
	"sync"

	"encoding/json"

	"sort"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/pkg/errors"
)

type innerState struct {
	types.State
	Counter uint `json:"c"`
}

type counter struct {
	txCounter map[string]uint
	mu        sync.Mutex
}

func (c *counter) increment(key string) uint {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.txCounter[key]++
	return c.txCounter[key]
}

func (c *counter) decrement(key string) uint {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.txCounter[key]--
	return c.txCounter[key]
}

type undoLog struct {
	ldb     blockchain.LevelDBGetterPutterDeleter
	counter counter
}

func newUndoLog(store blockchain.LevelDBGetterPutterDeleter) types.StateStorage {
	return &undoLog{ldb: store, counter: counter{txCounter: make(map[string]uint)}}
}

func (rs *undoLog) Save(state types.State) error {
	inner := innerState{State: state}
	inner.Counter = rs.counter.increment("default")

	data, err := json.Marshal(inner)
	if err != nil {
		return err
	}
	err = rs.ldb.Put(inner.Transaction, data, nil)
	if err != nil {
		return err
	}
	return nil
}

// TODO simplify this by removing second cycle
func (rs *undoLog) Get() ([]types.State, error) {
	txses := make([]innerState, 0)
	var err error

	iter := rs.ldb.NewIterator(nil, nil)
	for iter.Next() {
		inner := innerState{}
		err = json.Unmarshal([]byte(iter.Value()), &inner)
		if err != nil {
			return nil, errors.Wrapf(err, "retrieving block transactions")
		}

		txses = append(txses, inner)
	}

	iter.Release()
	err = iter.Error()
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving block transactions")
	}

	sort.Slice(txses, func(i, j int) bool {
		return txses[i].Counter > txses[j].Counter
	})

	states := make([]types.State, 0)
	for _, tx := range txses {
		states = append(states, tx.State)
	}

	return states, nil
}
