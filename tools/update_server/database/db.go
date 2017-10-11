package database

import (
	"errors"

	"github.com/boltdb/bolt"
)

type Database struct {
	Database *bolt.DB
	open     bool
}

var bucketName = []byte("updateDB")

func (db *Database) Open(filename string) error {
	var err error
	db.Database, err = bolt.Open(filename, 0600, nil)
	if err != nil {
		return err
	}
	db.open = true
	return nil
}

func (db *Database) GetVersionsList() ([]string, error) {
	var result []string
	err := db.Database.View(func(tx *bolt.Tx) error {
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
	err := db.Database.View(func(tx *bolt.Tx) error {
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
	return db.Database.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return errors.New("can't create bucket")
		}

		err = b.Put([]byte(version), binary)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *Database) DeleteBinary(version string) error {
	return db.Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete([]byte(version))
	})
}
