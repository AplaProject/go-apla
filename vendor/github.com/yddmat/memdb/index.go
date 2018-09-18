package memdb

import (
	"errors"

	"github.com/tidwall/btree"
	"github.com/tidwall/match"
)

const btreeDegrees = 64

var (
	ErrEmptyIndex   = errors.New("index name is empty")
	ErrIndexExists  = errors.New("index already exists")
	ErrUnknownIndex = errors.New("unknown index")
)

type Index struct {
	name    string
	pattern string
	tree    *btree.BTree
	sortFn  func(a, b string) bool
}

func NewIndex(name, pattern string, sortFn func(a, b string) bool) *Index {
	i := new(Index)
	i.tree = btree.New(btreeDegrees, i)
	i.pattern = pattern
	i.name = name
	i.sortFn = sortFn
	return i
}

func (idx *Index) insert(item btree.Item) {
	idx.tree.ReplaceOrInsert(item)
}

func (idx *Index) remove(item btree.Item) {
	idx.tree.Delete(item)
}

// Indexes is not thread-safe
type Indexes struct {
	storage map[string]*Index
}

func newIndexer() *Indexes {
	return &Indexes{
		storage: make(map[string]*Index),
	}
}

func (idxer *Indexes) AddIndex(index *Index) error {
	if index.name == "" {
		return ErrEmptyIndex
	}

	if _, ok := idxer.storage[index.name]; ok {
		return ErrIndexExists
	}

	idxer.storage[index.name] = index

	return nil
}

func (idxer *Indexes) RemoveIndex(name string) error {
	if name == "" {
		return ErrEmptyIndex
	}

	delete(idxer.storage, name)

	return nil
}

func (idxer *Indexes) GetIndex(name string) *Index {
	for indexName, index := range idxer.storage {
		if name == indexName {
			return index
		}
	}

	return nil
}

func (idxer *Indexes) Has(name string) bool {
	return idxer.GetIndex(name) != nil
}

// TODO rename to ReplaceOrInsert
func (idxer *Indexes) Insert(item *item, to ...string) {
	for _, index := range idxer.storage {
		if idxer.fit(index.name, to) && match.Match(string(item.key), index.pattern) {
			index.insert(item)
		}
	}
}

func (idxer *Indexes) Remove(item *item, from ...string) {
	for _, index := range idxer.storage {
		if idxer.fit(index.name, from) && match.Match(string(item.key), index.pattern) {
			index.remove(item)
		}
	}
}

func (idxer *Indexes) fit(current string, indexes []string) bool {
	if len(indexes) == 0 {
		return true
	}

	for _, v := range indexes {
		if v == current {
			return true
		}
	}

	return false
}

func (idxer *Indexes) Copy() *Indexes {
	newIndexer := newIndexer()

	for _, oldIdx := range idxer.storage {
		newIdx := NewIndex(oldIdx.name, oldIdx.pattern, oldIdx.sortFn)
		newIdx.tree = oldIdx.tree.Clone()

		err := newIndexer.AddIndex(newIdx)
		if err != nil {
			panic(err)
		}
	}

	return newIndexer
}
