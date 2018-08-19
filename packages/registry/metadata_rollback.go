package registry

import (
	"fmt"
	"sync"

	"encoding/json"

	"github.com/GenesisKernel/go-genesis/packages/storage/kv"
	"github.com/GenesisKernel/go-genesis/packages/types"
)

const prefix = "rollback_tx.%s.%s"

type state struct {
	c uint64 // operation counter
	r string // registry name
	e int64  // ecosystem ID
	v string // value
	k string // primary key
}

type MetadataRollback struct {
	tx kv.Transaction

	txCounter map[string]uint64
	mu        sync.Mutex
}

func (mr *MetadataRollback) saveDocumentState(block, tx []byte, registry *types.Registry, pk, value string) error {
	key := string(block) + string(tx)
	mr.mu.Lock()
	defer mr.mu.Unlock()

	counter := mr.txCounter[key]
	counter++

	state := state{c: counter, r: registry.Name, e: registry.Ecosystem.ID, v: value}
	jstate, err := json.Marshal(state)
	if err != nil {
		return err
	}

	err = mr.tx.Set(fmt.Sprintf(prefix, string(block), string(tx)), string(jstate))
	if err != nil {
		return err
	}

	mr.txCounter[key] = counter
	return nil
}
