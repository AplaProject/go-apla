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
	valueArgs := make([]interface{}, 0, len(rows)*len(fields))
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

			valueArgs = append(valueArgs, val)
		}
	}

	query += strings.Join(valueTemplates, ",")
	return
}
