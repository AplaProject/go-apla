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

package apiv2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
)

func TestGetUID(t *testing.T) {
	var ret getUIDResult
	err := sendGet(`getuid`, nil, &ret)
	if err != nil {
		var v map[string]string
		json.Unmarshal([]byte(err.Error()[4:]), &v)
		if v[`error`] == `E_NOTINSTALLED` {
			var instRes installResult
			err := sendPost(`install`, &url.Values{`db_port`: {`5432`}, `db_host`: {`localhost`},
				`type`: {`PRIVATE_NET`}, `db_name`: {`apla`}, `log_level`: {`ERROR`},
				`db_pass`: {`postgres`}, `db_user`: {`postgres`}}, &instRes)
			if err != nil {
				t.Error(err)
				return
			}
		} else {
			t.Error(err)
			return
		}
	}
	gAuth = ret.Token
	priv, pub, err := crypto.GenHexKeys()
	if err != nil {
		t.Error(err)
		return
	}
	sign, err := crypto.Sign(priv, ret.UID)
	if err != nil {
		t.Error(err)
		return
	}
	form := url.Values{"pubkey": {pub}, "signature": {hex.EncodeToString(sign)}}
	var lret loginResult
	err = sendPost(`login`, &form, &lret)
	if err != nil {
		t.Error(err)
		return
	}
	gAuth = lret.Token
	var ref refreshResult
	err = sendPost(`refresh`, &url.Values{"token": {lret.Refresh}}, &ref)
	if err != nil {
		t.Error(err)
		return
	}
	gAuth = ref.Token
}

func TestHashID(t *testing.T) {
	err := model.GormInit(`postgres`, `postgres`, `v2`) // v2 - specify your database
	if err != nil {
		t.Error(err)
	}
	model.DBConn.Exec(`DROP SEQUENCE IF EXISTS "hashid_id_seq" CASCADE;
		CREATE SEQUENCE "hashid_id_seq" START WITH 1;
		DROP TABLE IF EXISTS "hashid"; CREATE TABLE "hashid" (
		"id" bigint  NOT NULL default nextval('hashid_id_seq'),
		"hash" bytea  NOT NULL DEFAULT '',
		"name" character varying(255) NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "hashid" ADD CONSTRAINT "hashid_pkey" PRIMARY KEY (id);
		ALTER SEQUENCE "hashid_id_seq" owned by "hashid".id;
		CREATE INDEX "hashid_index_hash" ON "hashid" (hash);`)
	start := time.Now()

	for i := 0; i < 100000; i++ {
		hash, err := crypto.Hash([]byte(fmt.Sprintf(`My name %d`, i+1)))
		if err != nil {
			t.Error(err)
		}
		model.DBConn.Exec(`INSERT INTO hashid (hash,name) VALUES(?,?)`,
			hash, fmt.Sprintf(`My name %d`, i+1))
	}
	fmt.Println(`Time: `, time.Now().Sub(start))
}

func TestMaxID(t *testing.T) {
	err := model.GormInit(`postgres`, `postgres`, `v2`) // v2 - specify your database
	if err != nil {
		t.Error(err)
	}

	model.DBConn.Exec(`DROP TABLE IF EXISTS "maxid"; CREATE TABLE "maxid" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "maxid" ADD CONSTRAINT "maxid_pkey" PRIMARY KEY (id);
    `)
	start := time.Now()
	for i := 0; i < 100000; i++ {
		var id int64
		if i > 0 {
			id = int64(i)
		}
		model.DBConn.Exec(`INSERT INTO maxid (id,name) VALUES(?,?)`, id+1, fmt.Sprintf(`My name %d`, id+1))
	}
	fmt.Println(`Time: `, time.Now().Sub(start))

}
