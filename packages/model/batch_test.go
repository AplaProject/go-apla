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
	"testing"

	"github.com/stretchr/testify/require"
)

type TestBatchModel struct {
	ID   int64
	Name string
}

func (m TestBatchModel) TableName() string {
	return "test_batch"
}

func (m TestBatchModel) FieldValue(fieldName string) (interface{}, error) {
	switch fieldName {
	case "id":
		return m.ID, nil
	case "name":
		return m.Name, nil
	default:
		return nil, fmt.Errorf("Unknown field %s of TestBatchModel", fieldName)
	}
}

func TestPrepareQuery(t *testing.T) {
	slice := []BatchModel{
		TestBatchModel{ID: 1, Name: "first"},
		TestBatchModel{ID: 2, Name: "second"},
	}

	query, args, err := prepareQuery(slice, []string{"id", "name"})
	require.NoError(t, err)

	checkQuery := `INSERT INTO "test_batch" (id,name) VALUES (?,?),(?,?)`
	checkArgs := []interface{}{int64(1), "first", int64(2), "second"}

	require.Equal(t, checkQuery, query)
	require.Equal(t, checkArgs, args)
}
