package storage

import (
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type BoltStorage struct {
	bolt *bolt.DB
	open bool
}

var bucketName = []byte("updateDB")

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

func (db *BoltStorage) GetBinary(version string) ([]byte, error) {
	var binary []byte
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		binary = b.Get([]byte(version))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return binary, nil
}

func (db *BoltStorage) AddBinary(binary []byte, version string) error {
	if len(binary) == 0 {
		return errors.Errorf("empty binary")
	}
	if len(version) == 0 {
		return errors.Errorf("empty version")
	}

	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		err := b.Put([]byte(version), binary)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *BoltStorage) DeleteBinary(version string) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete([]byte(version))
	})
}
