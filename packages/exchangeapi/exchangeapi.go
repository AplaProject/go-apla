// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package exchangeapi

import (
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/boltdb/bolt"
	"github.com/op/go-logging"
)

var (
	boltDB *bolt.DB
	bucket = []byte(`Keys`)
	log    = logging.MustGetLogger("exchangeapi")
)

func init() {
	var err error
	boltDB, err = bolt.Open(*utils.BoltDir+"/exchangeapi.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//	NewKey()
}

func NewKey() ([]byte, error) {
	key := `23`
	err := boltDB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(bucket)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		if err := b.Put([]byte(key), []byte("test")); err != nil {
			return fmt.Errorf("put in bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	/*	var val []byte
		err = boltDB.View(func(tx *bolt.Tx) error {
			val = tx.Bucket(bucket).Get([]byte("foo"))
			return nil
		})
		fmt.Printf("The value of 'foo' is: %s %s\n", err, val)

		err = boltDB.View(func(tx *bolt.Tx) error {
			val = tx.Bucket(bucket).Get([]byte("23"))
			return nil
		})
		fmt.Printf("The value of '23' is: %s %s\n", err, val)*/
	return nil, nil
}
