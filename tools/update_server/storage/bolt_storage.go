package storage

import (
	"encoding/json"

	"github.com/GenesisKernel/go-genesis/tools/update_server/model"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type BoltStorage struct {
	bolt          *bolt.DB
	binaryStorage BinaryStorage

	open bool
}

var bucketName = []byte("builds")

// NewBoltStorage is creating bolt storage
func NewBoltStorage(binaryStorage BinaryStorage, filename string) (BoltStorage, error) {
	var err error
	var db BoltStorage

	db.binaryStorage = binaryStorage
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

func (db *BoltStorage) Get(build model.Build) (model.Build, error) {
	var fb model.Build
	if build.String() == "" {
		return fb, errors.Errorf("wrong system")
	}

	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		ub := b.Get([]byte(build.String()))

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

	// doesn't exists
	if fb.String() == "" {
		return fb, nil
	}

	body, err := db.binaryStorage.GetBinary(fb)
	if err != nil {
		return fb, errors.Wrapf(err, "retrieving build body from storage")
	}

	fb.Body = body
	return fb, nil
}

func (db *BoltStorage) Add(build model.Build) error {
	aeb, err := db.Get(build)
	if err != nil {
		return err
	}

	if aeb.String() != "" {
		return errors.Errorf("version %s already exists in storage", build.String())
	}

	if build.String() == "" {
		return errors.Errorf("wrong system")
	}

	err = db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		jb, err := json.Marshal(build)
		if err != nil {
			return err
		}

		err = b.Put([]byte(build.String()), jb)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "saving build to registry")
	}

	err = db.binaryStorage.SaveBuild(build)
	if err != nil {
		err = db.deleteFromRegistry(build)
		if err != nil {
			return errors.Wrapf(err, "removing build from registry after failed writing to filesystem")
		}

		return errors.Wrapf(err, "saving binary to filesystem")
	}

	return nil
}

func (db *BoltStorage) Delete(build model.Build) error {
	err := db.binaryStorage.DeleteBinary(build)
	if err != nil {
		return errors.Wrapf(err, "deleting binary from storage")
	}

	err = db.deleteFromRegistry(build)
	if err != nil {
		return errors.Wrapf(err, "deleting build from registry")
	}

	return nil
}

func (db *BoltStorage) deleteFromRegistry(build model.Build) error {
	if build.String() == "" {
		return errors.Errorf("wrong system")
	}

	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete([]byte(build.String()))
	})
}
