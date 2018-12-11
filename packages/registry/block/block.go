package block

import (
	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Registry struct {
	ldb  blockchain.LevelDBGetterPutterDeleter
	undo types.StateStorage
}

func NewBlockRegistry(ldb blockchain.LevelDBGetterPutterDeleter, undo types.StateStorage) *Registry {
	return &Registry{ldb: ldb, undo: undo}
}

func (bc *Registry) Get(key []byte, ro *opt.ReadOptions) ([]byte, error) {
	return bc.ldb.Get(key, ro)
}

func (bc *Registry) Put(key, value []byte, wo *opt.WriteOptions) error {
	err := bc.undo.Save(types.State{DBType: types.DBTypeBlockChain, Key: string(key)})
	if err != nil {
		return err
	}

	return bc.ldb.Put(key, value, wo)
}

func (bc *Registry) Delete(key []byte, wo *opt.WriteOptions) error {
	previous, err := bc.ldb.Get(key, nil)
	if err != nil {
		return err
	}

	err = bc.undo.Save(types.State{DBType: types.DBTypeBlockChain, Key: string(key), Value: string(previous)})
	if err != nil {
		return err
	}

	return bc.ldb.Delete(key, wo)
}

func (bc *Registry) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return bc.ldb.NewIterator(slice, ro)
}
