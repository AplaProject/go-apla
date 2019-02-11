package multi

import (
	"sync"

	"github.com/AplaProject/go-apla/packages/types"
	"github.com/pkg/errors"
)

const nopeSavePoint = ""

var (
	ErrActiveSavePoint    = errors.New("Previous save point is active")
	ErrNotExistsSavePoint = errors.New("Save point is not exists")
)

func newMultiTransaction() *MultiTransaction {
	return &MultiTransaction{
		tr: make(map[string]types.DBTransaction),
	}
}

type MultiTransaction struct {
	mu        sync.Mutex
	savePoint string
	tr        map[string]types.DBTransaction
}

func (mt *MultiTransaction) rollback() (err error) {
	for key, tr := range mt.tr {
		if err1 := tr.Rollback(); err1 != nil {
			err = errors.WithMessage(err1, key)
		}
	}
	return
}

func (mt *MultiTransaction) Rollback() (err error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	err = mt.rollback()
	mt.tr = make(map[string]types.DBTransaction)
	return
}

func (mt *MultiTransaction) Commit() (err error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	for key, t := range mt.tr {
		if err1 := t.Commit(); err1 != nil {
			mt.rollback()
			err = errors.WithMessage(err1, key)
		}
	}

	return
}

func (mt *MultiTransaction) SavePoint(id string) (err error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if len(mt.savePoint) > 0 {
		return ErrActiveSavePoint
	}

	for key, t := range mt.tr {
		if err1 := t.SavePoint(id); err1 != nil {
			err = errors.WithMessage(err1, key)
		}
	}

	mt.savePoint = id
	return
}

func (mt *MultiTransaction) ReleaseSavePoint() (err error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if len(mt.savePoint) == 0 {
		return ErrNotExistsSavePoint
	}

	for key, tr := range mt.tr {
		if err1 := tr.ReleaseSavePoint(mt.savePoint); err1 != nil {
			err = errors.WithMessage(err1, key)
		}
	}

	mt.savePoint = nopeSavePoint
	return
}

func (mt *MultiTransaction) RollbackSavePoint() (err error) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	for key, tr := range mt.tr {
		if err1 := tr.RollbackSavePoint(mt.savePoint); err1 != nil {
			err = errors.WithMessage(err1, key)
		}
	}

	mt.savePoint = nopeSavePoint
	return
}

func (mt *MultiTransaction) Get(key string) types.DBTransaction {
	return mt.tr[key]
}
