package multi

import (
	"sync"

	"github.com/AplaProject/go-apla/packages/storage/undo"
	"github.com/AplaProject/go-apla/packages/types"
)

type Database interface {
	Begin(u *undo.Stack) (types.DBTransaction, error)
}

type MultiStorage struct {
	mu   sync.Mutex
	db   map[string]Database
	undo *undo.Storage
}

func (ms *MultiStorage) Add(key string, db Database) {
	ms.mu.Lock()
	ms.db[key] = db
	ms.mu.Unlock()
}

func (ms *MultiStorage) Begin() (*MultiTransaction, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	mt := newMultiTransaction()
	for key, db := range ms.db {
		// TODO: set current tx
		undo := ms.undo.NewStack(key, "100500")
		tr, err := db.Begin(undo)
		if err != nil {
			mt.rollback()
			return nil, err
		}
		mt.tr[key] = tr
	}

	return mt, nil
}

func (ms *MultiStorage) UndoSave() {
	ms.undo.Save()
}

func NewMultiStorage() *MultiStorage {
	return &MultiStorage{
		db:   make(map[string]Database),
		undo: undo.NewStorage(),
	}
}
