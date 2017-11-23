package model

import (
	"database/sql"
	"fmt"

	"github.com/AplaProject/go-apla/packages/converter"
)

// SingleResult is a structure for the single result
type SingleResult struct {
	result []byte
	err    error
}

// Single is retrieving single result
func Single(query string, args ...interface{}) *SingleResult {
	var result []byte
	err := DBConn.Raw(query, args...).Row().Scan(&result)
	switch {
	case err == sql.ErrNoRows:
		return &SingleResult{[]byte(""), nil}
	case err != nil:
		return &SingleResult{[]byte(""), fmt.Errorf("%s in query %s %s", err, query, args)}
	}
	return &SingleResult{result, nil}
}

// Int64 converts bytes to int64
func (r *SingleResult) Int64() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return converter.BytesToInt64(r.result), nil
}

// Int converts bytes to int
func (r *SingleResult) Int() (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	return converter.BytesToInt(r.result), nil
}

// Float64 converts string to float64
func (r *SingleResult) Float64() (float64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return converter.StrToFloat64(string(r.result)), nil
}

// String returns string
func (r *SingleResult) String() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return string(r.result), nil
}

// Bytes returns []byte
func (r *SingleResult) Bytes() ([]byte, error) {
	if r.err != nil {
		return []byte(""), r.err
	}
	return r.result, nil
}

// OneRow is storing one row result
type OneRow struct {
	result map[string]string
	err    error
}

// String is extracts result from OneRow as string
func (r *OneRow) String() (map[string]string, error) {
	if r.err != nil {
		return r.result, r.err
	}
	return r.result, nil
}

// Bytes is extracts result from OneRow as []byte
func (r *OneRow) Bytes() (map[string][]byte, error) {
	result := make(map[string][]byte)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = []byte(v)
	}
	return result, nil
}

// Int64 is extracts result from OneRow as int64
func (r *OneRow) Int64() (map[string]int64, error) {
	result := make(map[string]int64)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = converter.StrToInt64(v)
	}
	return result, nil
}

// Float64 is extracts result from OneRow as float64
func (r *OneRow) Float64() (map[string]float64, error) {
	result := make(map[string]float64)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = converter.StrToFloat64(v)
	}
	return result, nil
}

// Int is extracts result from OneRow as int
func (r *OneRow) Int() (map[string]int, error) {
	result := make(map[string]int)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = converter.StrToInt(v)
	}
	return result, nil
}

// GetAllTransaction is retrieve all query result rows
func GetAllTransaction(transaction *DbTransaction, query string, countRows int, args ...interface{}) ([]map[string]string, error) {
	var result []map[string]string
	rows, err := GetDB(transaction).Raw(query, args...).Rows()
	if err != nil {
		return result, fmt.Errorf("%s in query %s %s", err, query, args)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return result, fmt.Errorf("%s in query %s %s", err, query, args)
	}
	// Make a slice for the values
	values := make([][]byte /*sql.RawBytes*/, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	r := 0
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return result, fmt.Errorf("%s in query %s %s", err, query, args)
		}

		// Now do something with the data.
		// Here we just print each column as a string.
		var value string
		rez := make(map[string]string)
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			rez[columns[i]] = value
		}
		result = append(result, rez)
		r++
		if countRows != -1 && r >= countRows {
			break
		}
	}
	if err = rows.Err(); err != nil {
		return result, fmt.Errorf("%s in query %s %s", err, query, args)
	}
	return result, nil
}

// GetAll returns all transaction
func GetAll(query string, countRows int, args ...interface{}) ([]map[string]string, error) {
	return GetAllTransaction(nil, query, countRows, args)
}

// GetAllTx returns all tx's
func GetAllTx(transaction *DbTransaction, query string, countRows int, args ...interface{}) ([]map[string]string, error) {
	return GetAllTransaction(transaction, query, countRows, args)
}

// GetOneRowTransaction returns one row from transactions
func GetOneRowTransaction(transaction *DbTransaction, query string, args ...interface{}) *OneRow {
	result := make(map[string]string)
	all, err := GetAllTransaction(transaction, query, 1, args...)
	if err != nil {
		return &OneRow{result, fmt.Errorf("%s in query %s %s", err, query, args)}
	}
	if len(all) == 0 {
		return &OneRow{result, nil}
	}
	return &OneRow{all[0], nil}
}

// GetOneRow returns one row
func GetOneRow(query string, args ...interface{}) *OneRow {
	return GetOneRowTransaction(nil, query, args...)
}
