// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package model

import (
	"fmt"

	"github.com/gobuffalo/fizz"
	"github.com/gobuffalo/fizz/translators"
)

var _ fizz.Translator = (*translators.Postgres)(nil)
var pgt = translators.NewPostgres()

/*
DROP TABLE IF EXISTS "1_keys"; CREATE TABLE "1_keys" (
	"id" bigint  NOT NULL DEFAULT '0',
	"pub" bytea  NOT NULL DEFAULT '',
	"amount" decimal(30) NOT NULL DEFAULT '0' CHECK (amount >= 0),
	"maxpay" decimal(30) NOT NULL DEFAULT '0' CHECK (maxpay >= 0),
	"deposit" decimal(30) NOT NULL DEFAULT '0' CHECK (deposit >= 0),
	"multi" bigint NOT NULL DEFAULT '0',
	"deleted" bigint NOT NULL DEFAULT '0',
	"blocked" bigint NOT NULL DEFAULT '0',
	"ecosystem" bigint NOT NULL DEFAULT '1',
	"account" char(24) NOT NULL
	);
	ALTER TABLE ONLY "1_keys" ADD CONSTRAINT "1_keys_pkey" PRIMARY KEY (ecosystem,id);
*/
func testFizz() {
	res, err := fizz.AString(`drop_table("1_keys", {"if_exists": true})`, pgt)
	res, err = fizz.AString(`sql("DROP TABLE IF EXISTS \"1_keys\";")
	create_table("users") {
		t.Column("id", "bigint", {primary: true})
		t.Column("email", "bigint", {"default": "0"})
		t.Column("twitter_handle", "string", {"size": 50})
		t.Column("age", "bigint", {"default_raw": "'0' CHECK (amount > 0)"})
		t.Column("admin", "bytea", {})
		t.Column("company_id", "uuid", {"default_raw": "uuid_generate_v1()"})
		t.Column("bio", "text", {"null": true})
		t.Column("joined_at", "timestamp", {})
		t.Index("email", {"unique": true})
	  }
	  add_index("table_name", "column_name", {"unique": true})`, pgt)
	fmt.Println(`POSTGRES`, err, res)
}
