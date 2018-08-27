package memdb

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/tidwall/btree"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrTxClosed      = errors.New("transaction closed")
	ErrTxNotWritable = errors.New("transaction is not writable")
)

type Transaction struct {
	writable bool

	db           *Database
	newIndexes   *Indexes
	pendingItems map[dbKey]struct{}
	mu           sync.RWMutex
}

func (tx *Transaction) Set(key, value string) error {
	k := dbKey(key)
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.db == nil {
		return ErrTxClosed
	}

	if !tx.writable {
		return ErrTxNotWritable
	}

	_, err := tx.getKey(k)
	if err != ErrNotFound {
		return ErrAlreadyExists
	}

	new := &item{key: k, value: value}
	tx.createItem(new)
	tx.newIndexes.Insert(new)

	return nil
}

func (tx *Transaction) Get(key string) (string, error) {
	if tx.db == nil {
		return "", ErrTxClosed
	}

	item, err := tx.getKey(dbKey(key))
	if err != nil {
		return "", err
	}

	return item.value, nil
}

func (tx *Transaction) Delete(key string) error {
	k := dbKey(key)
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.db == nil {
		return ErrTxClosed
	}

	if !tx.writable {
		return ErrTxNotWritable
	}

	item, err := tx.getKey(k)
	if err != nil {
		return err
	}

	tx.updateItem(k, nil, true)
	tx.newIndexes.Remove(&item)

	return nil
}

func (tx *Transaction) Update(key, value string) (string, error) {
	k := dbKey(key)
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.db == nil {
		return "", ErrTxClosed
	}

	if !tx.writable {
		return "", ErrTxNotWritable
	}

	old, err := tx.getKey(k)
	if err != nil {
		return "", err
	}

	update := &item{key: k, value: value}
	tx.updateItem(k, update, false)
	tx.newIndexes.Insert(update)

	return old.value, nil
}

func (tx *Transaction) AddIndex(indexes ...*Index) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.db == nil {
		return ErrTxClosed
	}

	if !tx.writable {
		return ErrTxNotWritable
	}

	rollbackInserted := func(inserted []string) {
		for _, idx := range inserted {
			if err := tx.newIndexes.RemoveIndex(idx); err != nil {
				panic(err)
			}
		}
	}

	inserted := make([]string, 0)
	for _, index := range indexes {
		err := tx.newIndexes.AddIndex(index)
		if err != nil {
			rollbackInserted(inserted)
			return err
		}

		inserted = append(inserted, index.name)
	}

	for _, key := range tx.db.items.keys() {
		revision, err := tx.getKey(key)
		if err != nil {
			rollbackInserted(inserted)
			return err
		}

		tx.newIndexes.Insert(&revision, inserted...)
	}

	return nil
}

func (tx *Transaction) Ascend(index string, iterator func(key, value string) bool) error {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if tx.db == nil {
		return ErrTxClosed
	}

	if index == "" {
		return ErrEmptyIndex
	}

	indexes := tx.db.indexes
	if tx.writable {
		indexes = tx.newIndexes
	}

	i := indexes.GetIndex(index)
	if i == nil {
		return ErrUnknownIndex
	}

	var curitem *item
	i.tree.Ascend(func(bitem btree.Item) bool {
		curitem = bitem.(*item)
		return iterator(string(curitem.key), curitem.value)
	})

	return nil
}

func (tx *Transaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.db == nil {
		return ErrTxClosed
	}

	db := tx.db
	tx.db = nil

	if tx.writable {
		save := make([]fileItem, 0)

		for key := range tx.pendingItems {
			dbItem := db.items.get(key)
			dbItem.Lock()

			if dbItem.pendingDeleted {
				if dbItem.current != nil {
					save = append(save, fileItem{item: item{key: key}, command: commandDEL})
				}
				dbItem.Unlock()
				db.items.remove(key)
				continue
			}

			// Delete old record
			if dbItem.current != nil {
				save = append(save, fileItem{item: item{key: key}, command: commandDEL})
			}

			save = append(save, fileItem{item: item{key: key, value: dbItem.pending.value}, command: commandSET})
			dbItem.current = dbItem.pending
			dbItem.pending = nil
			dbItem.Unlock()
		}

		tx.pendingItems = nil
		db.indexes = tx.newIndexes

		// Write to disk
		if db.persist {
			err := db.persistentStorage.write(save...)
			if err != nil {
				return err
			}
		}

		db.writeTx.Unlock()
	}

	return nil
}

func (tx *Transaction) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.db == nil {
		return ErrTxClosed
	}

	db := tx.db

	if tx.writable {
		tx.newIndexes = nil
		tx.pendingItems = nil
		db.writeTx.Unlock()
	}

	return nil
}

func (tx *Transaction) getKey(key dbKey) (item, error) {
	dbItem := tx.db.items.get(key)

	if dbItem == nil {
		return item{}, ErrNotFound
	}

	dbItem.RLock()
	defer dbItem.RUnlock()

	// Item doesn't created "yet"
	if !tx.writable && dbItem.current == nil || (tx.writable && dbItem.pendingDeleted) {
		return item{}, ErrNotFound
	}

	// Item was already updated at this transaction
	if tx.writable && !dbItem.pendingDeleted && dbItem.pending != nil {
		return *dbItem.pending, nil
	}

	return *dbItem.current, nil
}

func (tx *Transaction) createItem(item *item) {
	dbItem := &dbItem{key: item.key, pending: item}

	dbItem.Lock()
	tx.db.items.set(item.key, dbItem)
	tx.pendingItems[dbItem.key] = struct{}{}
	dbItem.Unlock()
}

func (tx *Transaction) updateItem(key dbKey, new *item, deleted bool) {
	dbItem := tx.db.items.get(key)

	dbItem.Lock()
	dbItem.pendingDeleted = deleted
	dbItem.pending = new
	tx.pendingItems[key] = struct{}{}
	dbItem.Unlock()
}
