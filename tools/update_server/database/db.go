package database

import (
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type Database struct {
	storage *bolt.DB
	open    bool
}

var bucketName = []byte("updateDB")

func NewDatabase(filename string) (Database, error) {
	var err error
	var db Database

	db.storage, err = bolt.Open(filename, 0600, nil)
	if err != nil {
		return db, errors.Wrapf(err, "opening boltdb file storage")
	}

	db.open = true
	// We need to check all buckets to availability before doing some stuff
	err = db.storage.Update(func(tx *bolt.Tx) error {
		_, err := tx.Bucket(bucketName).CreateBucketIfNotExists(bucketName)
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

func (db *Database) GetVersionsList() ([]string, error) {
	var result []string
	err := db.storage.View(func(tx *bolt.Tx) error {
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

func (db *Database) GetBinary(version string) ([]byte, error) {
	var binary []byte
	err := db.storage.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		binary = b.Get([]byte(version))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return binary, nil
}

func (db *Database) AddBinary(binary []byte, version string) error {
	return db.storage.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		err := b.Put([]byte(version), binary)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *Database) DeleteBinary(version string) error {
	return db.storage.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete([]byte(version))
	})
}
