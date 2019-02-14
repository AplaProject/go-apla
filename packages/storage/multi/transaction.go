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

func NewTransaction() Transaction {
	return Transaction{
		current: make(map[string]types.DBTransaction),
	}
}

type Transaction struct {
	mu sync.Mutex

	savePoint string
	current   map[string]types.DBTransaction
}

func (t *Transaction) rollback() (err error) {
	for key, tr := range t.current {
		if errTr := tr.Rollback(); errTr != nil {
			err = errors.WithMessage(errTr, key)
		}
	}
	return
}

func (t *Transaction) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.rollback()
}

func (t *Transaction) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for key, tr := range t.current {
		if err := tr.Commit(); err != nil {
			t.rollback()
			return errors.WithMessage(err, key)
		}
	}
	return nil
}

func (t *Transaction) SavePoint(savePoint string) (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.savePoint) > 0 {
		return ErrActiveSavePoint
	}

	for key, tr := range t.current {
		if errTr := tr.SavePoint(savePoint); errTr != nil {
			err = errors.WithMessage(errTr, key)
		}
	}
	t.savePoint = savePoint
	return
}

func (t *Transaction) ReleaseSavePoint() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.savePoint) == 0 {
		return ErrNotExistsSavePoint
	}

	for key, tr := range t.current {
		if errTr := tr.ReleaseSavePoint(t.savePoint); errTr != nil {
			err = errors.WithMessage(errTr, key)
		}
	}
	t.savePoint = nopeSavePoint
	return
}

func (t *Transaction) RollbackSavePoint() (err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for key, tr := range t.current {
		if errTr := tr.RollbackSavePoint(t.savePoint); errTr != nil {
			err = errors.WithMessage(errTr, key)
		}
	}

	t.savePoint = nopeSavePoint
	return
}

func (t *Transaction) Set(key string, tr types.DBTransaction) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.current[key] = tr
}

func (t *Transaction) Get(key string) types.DBTransaction {
	return t.current[key]
}
