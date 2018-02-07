package storage

import (
	"encoding/json"

	"github.com/GenesisCommunity/go-genesis/tools/update_server/model"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type BoltStorage struct {
	bolt *bolt.DB
	open bool
}

var bucketName = []byte("builds")

// NewBoltStorage is creating bolt storage
func NewBoltStorage(filename string) (BoltStorage, error) {
	var err error
	var db BoltStorage

	db.bolt, err = bolt.Open(filename, 0600, nil)
	if err != nil {
		return db, errors.Wrapf(err, "opening boltdb file storage")
	}

	db.open = true
	// We need to check all buckets to availability before doing some stuff
	err = db.bolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return errors.Wrapf(err, "creating bucket")
		}
		return nil
	})
	if err != nil {
		return db, errors.Wrapf(err, "updating storage")
	}

	return db, nil
}

func (db *BoltStorage) GetVersionsList() ([]model.Version, error) {
	var result []model.Version
	err := db.bolt.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketName).Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			kv, err := model.NewVersion(string(k))
			if err != nil {
				return err
			}

			result = append(result, kv)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *BoltStorage) Get(binary model.Build) (model.Build, error) {
	var fb model.Build
	if binary.String() == "" {
		return fb, errors.Errorf("wrong system")
	}

	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		ub := b.Get([]byte(binary.String()))

		if ub != nil {
			err := json.Unmarshal(ub, &fb)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fb, err
	}
	return fb, nil
}

func (db *BoltStorage) Add(binary model.Build) error {
	aeb, err := db.Get(binary)
	if err != nil {
		return err
	}

	if aeb.String() != "" {
		return errors.Errorf("version %s already exists in storage", binary.String())
	}

	if binary.String() == "" {
		return errors.Errorf("wrong system")
	}

	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		jb, err := json.Marshal(binary)
		if err != nil {
			return err
		}

		err = b.Put([]byte(binary.String()), jb)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *BoltStorage) Delete(binary model.Build) error {
	if binary.String() == "" {
		return errors.Errorf("wrong system")
	}

	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete([]byte(binary.String()))
	})
}
