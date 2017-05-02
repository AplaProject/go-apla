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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	//	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/boltdb/bolt"
	"github.com/op/go-logging"
)

const (
	forpsw = `test string for decoding`
)

var (
	boltDB   *bolt.DB
	bucket   = []byte(`Keys`)
	settings = []byte(`Settings`)
	log      = logging.MustGetLogger("exchangeapi")
)

type DefaultApi struct {
	Error string `json:"error"`
}

func init() {
	var err error
	boltDB, err = bolt.Open(*utils.BoltDir+"/exchangeapi.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	var encTest []byte
	err = boltDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(settings)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	boltDB.View(func(tx *bolt.Tx) error {
		encTest = tx.Bucket(settings).Get([]byte("EncTest"))
		return nil
	})
	if len(*utils.ApiToken) > 0 {
		err = boltDB.Update(func(tx *bolt.Tx) error {
			err = tx.Bucket(settings).Put([]byte("Token"), []byte(*utils.ApiToken))
			return err
		})
	} else {
		err = boltDB.View(func(tx *bolt.Tx) error {
			*utils.ApiToken = string(tx.Bucket(settings).Get([]byte("Token")))
			return nil
		})
	}
	if err != nil {
		log.Fatal(fmt.Errorf(`BoltDB put/get token: %v`, err))
	}
	if len(*utils.BoltPsw) > 0 {
		if len(encTest) == 0 {
			err = boltDB.Update(func(tx *bolt.Tx) error {
				var encrypted []byte

				encrypted, err = encryptBytes([]byte(forpsw))
				if err != nil {
					return err
				}
				err = tx.Bucket(settings).Put([]byte("EncTest"), encrypted)
				return err
			})
			if err != nil {
				log.Fatal(fmt.Errorf(`BoltDB init: %v`, err))
			}
		} else {
			decrypted, err := decryptBytes(encTest)
			if err != nil {
				log.Fatal(fmt.Errorf(`Check BoltPsw: %v`, err))
			}
			if string(decrypted) != forpsw {
				log.Fatal(fmt.Errorf(`Wrong BoltPsw`))
			}
		}
	} else {
		if len(encTest) > 0 {
			log.Fatal(fmt.Errorf(`-boltPsw parameter must be specified`))
		}
	}
}

func encryptBytes(input []byte) (output []byte, err error) {
	pass := sha256.Sum256([]byte(*utils.BoltPsw))
	output, _, err = utils.EncryptCFB(input, pass[:], make([]byte, 16))
	output = output[16:]
	if err != nil {
		return
	}
	return
}

func decryptBytes(input []byte) (output []byte, err error) {
	pass := sha256.Sum256([]byte(*utils.BoltPsw))
	output, err = utils.DecryptCFB(make([]byte, 16), input, pass[:])
	return
}

func genNewKey() ([]byte, error) {
	if len(*utils.BoltPsw) == 0 {
		return nil, fmt.Errorf(`-boltPsw password is not defined`)
	}
	priv, pub := lib.GenKeys()

	privKey, err := hex.DecodeString(priv)
	if err != nil {
		return nil, err
	}
	pubKey, err := hex.DecodeString(pub)
	if err != nil {
		return nil, err
	}
	address := int64(lib.Address(pubKey))

	err = boltDB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		input, err := encryptBytes(privKey)
		if err != nil {
			return err
		}
		if err := b.Put([]byte(utils.Int64ToStr(address)), input); err != nil {
			return fmt.Errorf("put in bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func Api(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("exchangeapi Recovered", r)
			fmt.Println("exchangeapi Recovered", r)
		}
	}()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.ParseForm()

	token := r.FormValue("token")
	if len(*utils.ApiToken) > 0 && token != *utils.ApiToken {
		w.Write([]byte(`{"error": "Invalid token"}`))
		return
	}
	var ret interface{}
	switch r.URL.Path {
	case `/exchangeapi/newkey`:
		ret = newKey(r)
	case `/exchangeapi/send`:
		ret = send(r)
	default:
		ret = DefaultApi{`Unknown request`}
	}
	jsonData, err := json.Marshal(ret)
	if err != nil {
		ret = DefaultApi{err.Error()}
	}
	w.Write(jsonData)
}
