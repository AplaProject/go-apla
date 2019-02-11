package memdb

import (
	"github.com/AplaProject/go-apla/packages/storage/undo"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/AplaProject/memdb"
)

type Storage struct {
	db *memdb.Database
}

func (s *Storage) Begin(undo *undo.Stack) (types.DBTransaction, error) {
	return &Transaction{
		tx:   s.db.Begin(true),
		undo: undo,
	}, nil
}

func (s *Storage) BeginRead() *Transaction {
	return &Transaction{
		tx: s.db.Begin(false),
	}
}

func NewStorage(path string) (*Storage, error) {
	db, err := memdb.OpenDB(path, true)
	if err != nil {
		return nil, err
	}
	return &Storage{db}, nil
}
