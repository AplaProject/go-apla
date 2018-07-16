package kv

import (
	"github.com/pkg/errors"
	"github.com/tidwall/buntdb"
)

type BuntDatabase struct {
	db *buntdb.DB
}

func NewBuntDatabase(db *buntdb.DB) *BuntDatabase {
	return &BuntDatabase{db: db}
}

func (bd *BuntDatabase) Begin(writable bool) (Transaction, error) {
	tx, err := bd.db.Begin(writable)
	if err != nil {
		return nil, errors.Wrapf(err, "starting transaction (writable - %t)", writable)
	}

	transaction := BuntTransaction{}
	transaction.tx = tx
	return &transaction, nil
}

type BuntTransaction struct {
	tx *buntdb.Tx
}

func (bt *BuntTransaction) Rollback() error {
	return bt.tx.Rollback()
}

func (bt *BuntTransaction) Commit() error {
	return bt.tx.Commit()
}

func (bt *BuntTransaction) Insert(key, value string) error {
	_, _, err := bt.tx.Set(key, value, nil)
	if err != nil {
		return errors.Wrapf(err, "inserting %s %s", key, value)
	}

	return nil
}

func (bt *BuntTransaction) Get(key string) (string, error) {
	value, err := bt.Get(key)
	if err != nil {
		return "", errors.Wrapf(err, "retrieving %s value from storage", key)
	}

	return value, nil
}

func (bt *BuntTransaction) Walk(keyPattern string, fn func(value string) bool) error {
	return bt.tx.AscendKeys(keyPattern, func(key, value string) bool {
		return fn(value)
	})
}

//
//func (bt *BuntTransaction) getTransaction(writable bool) (*buntdb.Tx, error) {
//	if bt.inTx() {
//		return bt.tx, nil
//	}
//	tx, err := bt.dbConnection.Begin(writable)
//	if err != nil {
//		return tx, errors.Wrapf(err, "starting buntdb transaction (writable - %t)", writable)
//	}
//	return tx, nil
//}
//
//func (bt *BuntTransaction) inTx() bool {
//	return bt.tx != nil
//}
//
//func (bt *BuntTransaction) atomicOperationDone(localTx *buntdb.Tx) error {
//	if bt.tx == nil {
//		err := localTx.Commit()
//		if err != nil {
//			return errors.Wrapf(err, "committing inserting transaction")
//		}
//	}
//
//	return nil
//}
