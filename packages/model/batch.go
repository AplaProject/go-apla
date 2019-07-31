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
	"strings"
)

const maxBatchRows = 1000

// BatchModel allows bulk insert on BatchModel slice
type BatchModel interface {
	TableName() string
	FieldValue(fieldName string) (interface{}, error)
}

// BatchInsert create and execute batch queries from rows splitted by maxBatchRows and fields
func BatchInsert(rows []BatchModel, fields []string) error {
	queries, values, err := batchQueue(rows, fields)
	if err != nil {
		return err
	}

	for i := 0; i < len(queries); i++ {
		if err := DBConn.Exec(queries[i], values[i]...).Error; err != nil {
			return err
		}
	}

	return nil
}

func batchQueue(rows []BatchModel, fields []string) (queries []string, values [][]interface{}, err error) {
	for len(rows) > 0 {
		if len(rows) > maxBatchRows {
			q, vals, err := prepareQuery(rows[:maxBatchRows], fields)
			if err != nil {
				return queries, values, err
			}

			queries = append(queries, q)
			values = append(values, vals)
			rows = rows[maxBatchRows:]
			continue
		}

		q, vals, err := prepareQuery(rows, fields)
		if err != nil {
			return queries, values, err
		}

		queries = append(queries, q)
		values = append(values, vals)
		rows = nil
	}

	return
}

func prepareQuery(rows []BatchModel, fields []string) (query string, values []interface{}, err error) {
	valueTemplates := make([]string, 0, len(rows))
	values = make([]interface{}, 0, len(rows)*len(fields))
	query = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES `, rows[0].TableName(), strings.Join(fields, ","))

	rowQSlice := make([]string, 0, len(fields))
	for range fields {
		rowQSlice = append(rowQSlice, "?")
	}

	valueTemplate := fmt.Sprintf("(%s)", strings.Join(rowQSlice, ","))

	for _, row := range rows {
		valueTemplates = append(valueTemplates, valueTemplate)
		for _, field := range fields {
			val, err := row.FieldValue(field)
			if err != nil {
				return query, values, err
			}

			values = append(values, val)
		}
	}

	query += strings.Join(valueTemplates, ",")
	return
}
