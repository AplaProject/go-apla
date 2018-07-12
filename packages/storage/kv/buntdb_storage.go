package kv

import (
	"github.com/pkg/errors"
	"github.com/tidwall/buntdb"
)

type BuntDBStorage struct {
	dbConnection *buntdb.DB
}

func NewBuntDBStorage(dbConnection *buntdb.DB) *BuntDBStorage {
	return &BuntDBStorage{dbConnection: dbConnection}
}

func (cr *BuntDBStorage) Insert(key, value string) error {
	err := cr.dbConnection.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, value, nil)
		return err
	})

	if err != nil {
		return errors.Wrapf(err, "inserting %s %s", key, value)
	}
	return nil
}

func (cr *BuntDBStorage) Get(key string) (string, error) {
	var value string
	var err error

	err = cr.dbConnection.View(func(tx *buntdb.Tx) error {
		value, err = tx.Get(key)
		if err != nil {
			return err
		}
		return err
	})
	return value, errors.Wrapf(err, "getting value from storage", key)
}
