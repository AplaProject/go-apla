package memdb

import (
	"github.com/AplaProject/go-apla/packages/storage/undo"
	"github.com/AplaProject/go-apla/packages/types"

	"github.com/AplaProject/memdb"
)

type Transaction struct {
	tx   *memdb.Transaction
	undo *undo.Stack
}

func (t *Transaction) Get(key string) (*types.Map, error) {
	data, err := t.tx.Get(key)
	if err != nil {
		return nil, err
	}

	return toMap(data)
}

func (t *Transaction) Set(key string, val *types.Map) error {
	data, err := fromMap(val)
	if err != nil {
		return err
	}

	err = t.tx.Set(key, data)
	if err != nil {
		return err
	}

	t.undo.PushState(&undo.State{
		Key: key,
	})

	return nil
}

func (t *Transaction) Update(key string, val *types.Map) error {
	prevValue, err := t.tx.Get(key)
	if err != nil {
		return err
	}

	prev, err := toMap(prevValue)
	if err != nil {
		return err
	}

	mergeMap(prev, val)

	data, err := fromMap(prev)
	if err != nil {
		return err
	}

	_, err = t.tx.Update(key, data)
	if err != nil {
		return err
	}

	t.undo.PushState(&undo.State{
		Key:   key,
		Value: prevValue,
	})

	return nil
}

func (t *Transaction) IsFound(err error) bool {
	return memdb.ErrNotFound != err
}

func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

func (t *Transaction) SavePoint(tx string) error {
	t.undo.Reset(tx)
	return nil
}

func (t *Transaction) RollbackSavePoint(_ string) (err error) {
	stack := t.undo.Current()
	for i := len(stack) - 1; i > 0; i-- {
		err = t.Undo(stack[i])
		if err != nil {
			return
		}
	}
	return
}

func (t *Transaction) ReleaseSavePoint(_ string) error {
	t.undo.Release()
	return nil
}

func (t *Transaction) Undo(s *undo.State) (err error) {
	if len(s.Value) > 0 {
		_, err = t.tx.Update(s.Key, s.Value)
		return
	}
	return t.tx.Delete(s.Key)
}
