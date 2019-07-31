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

package queryBuilder

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
)

// query="SELECT ,,,id,amount,\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\"ecosystem\"
// FROM \"1_keys\" \nWHERE  AND id = -6752330173818123413 AND ecosystem = '1'\n"

// fields="[+amount]"
// values="[2912910000000000000]"

// whereF="[id]"
// whereV="[-6752330173818123413]"

type TestKeyTableChecker struct {
	Val bool
}

func (tc TestKeyTableChecker) IsKeyTable(tableName string) bool {
	return tc.Val
}
func TestSqlFields(t *testing.T) {
	qb := smartQueryBuilder{
		Entry:        log.WithFields(log.Fields{"mod": "test"}),
		table:        "1_keys",
		Fields:       []string{"+amount"},
		FieldValues:  []interface{}{2912910000000000000},
		WhereFields:  []string{"id"},
		WhereValues:  []string{"-6752330173818123413"},
		KeyTableChkr: TestKeyTableChecker{true},
	}

	fields, err := qb.GetSQLSelectFieldsExpr()
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(fields)
}
