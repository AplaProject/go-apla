package storage

import (
	"encoding/json"

	"github.com/AplaProject/go-apla/tools/update_server/model"
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

func (db *BoltStorage) GetVersionsList() ([]string, error) {
	var result []string
	err := db.bolt.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketName).Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			result = append(result, string(k))
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
	if binary.GetSystem() == "" {
		return fb, errors.Errorf("wrong system")
	}

	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		ub := b.Get([]byte(binary.GetSystem()))

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

	if aeb.GetSystem() != "" {
		return errors.Errorf("version %s already exists in storage", binary.GetSystem())
	}

	if binary.GetSystem() == "" {
		return errors.Errorf("wrong system")
	}

	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		jb, err := json.Marshal(binary)
		if err != nil {
			return err
		}

		err = b.Put([]byte(binary.GetSystem()), jb)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *BoltStorage) Delete(binary model.Build) error {
	if binary.GetSystem() == "" {
		return errors.Errorf("wrong system")
	}

	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete([]byte(binary.GetSystem()))
	})
}
