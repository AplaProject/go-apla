package sql

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// SingleResult is a structure for the single result
type SingleResult struct {
	result []byte
	err    error
}

// Int64 converts bytes to int64
func (r *SingleResult) Int64() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return utils.BytesToInt64(r.result), nil
}

// Int converts bytes to int
func (r *SingleResult) Int() (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	return utils.BytesToInt(r.result), nil
}

// Float64 converts string to float64
func (r *SingleResult) Float64() (float64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return utils.StrToFloat64(string(r.result)), nil
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

// ListResult is a structure for the list result
type ListResult struct {
	result []string
	err    error
}

// Int64 converts all string values to int64
func (r *ListResult) Int64() ([]int64, error) {
	var result []int64
	if r.err != nil {
		return result, r.err
	}
	for _, v := range r.result {
		result = append(result, utils.StrToInt64(v))
	}
	return result, nil
}

// String return the slice of strings
func (r *ListResult) String() ([]string, error) {
	if r.err != nil {
		return r.result, r.err
	}
	return r.result, nil
}

type oneRow struct {
	result map[string]string
	err    error
}

func (r *oneRow) String() (map[string]string, error) {
	if r.err != nil {
		return r.result, r.err
	}
	return r.result, nil
}

func (r *oneRow) Bytes() (map[string][]byte, error) {
	result := make(map[string][]byte)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = []byte(v)
	}
	return result, nil
}

func (r *oneRow) Int64() (map[string]int64, error) {
	result := make(map[string]int64)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = utils.StrToInt64(v)
	}
	return result, nil
}

func (r *oneRow) Float64() (map[string]float64, error) {
	result := make(map[string]float64)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = utils.StrToFloat64(v)
	}
	return result, nil
}

func (r *oneRow) Int() (map[string]int, error) {
	result := make(map[string]int)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = utils.StrToInt(v)
	}
	return result, nil
}

// Single returns the single result of the query
func (db *DCDB) Single(query string, args ...interface{}) *SingleResult {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	var result []byte
	err := db.QueryRow(newQuery, newArgs...).Scan(&result)
	switch {
	case err == sql.ErrNoRows:
		return &SingleResult{[]byte(""), nil}
	case err != nil:
		return &SingleResult{[]byte(""), fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)}
	}
	if db.ConfigIni["sql_log"] == "1" {
		parent := utils.GetParent()
		log.Debug("SQL: %s / %v / %v", newQuery, newArgs, parent)
	}
	return &SingleResult{result, nil}
}

// GetMap returns the map of strings as the result of query
func (db *DCDB) GetMap(query string, name, value string, args ...interface{}) (map[string]string, error) {
	result := make(map[string]string)
	all, err := db.GetAll(query, -1, args...)
	if err != nil {
		return result, err
	}
	for _, v := range all {
		result[v[name]] = v[value]
	}
	return result, err
}

// GetList returns the result of the query as ListResult variable
func (db *DCDB) GetList(query string, args ...interface{}) *ListResult {
	var result []string
	all, err := db.GetAll(query, -1, args...)
	if err != nil {
		return &ListResult{result, err}
	}
	for _, v := range all {
		for _, v2 := range v {
			result = append(result, v2)
		}
	}
	return &ListResult{result, nil}
}

// GetAll returns the result of the query as slice of map[string]string
func (db *DCDB) GetAll(query string, countRows int, args ...interface{}) ([]map[string]string, error) {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	if db.ConfigIni["db_type"] == "postgresql" {
		query = ReplQ(query)
	}
	var result []map[string]string
	rows, err := db.Query(newQuery, newArgs...)
	if err != nil {
		return result, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
	}
	defer rows.Close()

	if db.ConfigIni["sql_log"] == "1" {
		parent := utils.GetParent()
		log.Debug("SQL: %s / %v / %v", newQuery, newArgs, parent)
	}
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return result, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
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
		//result[r] = make(map[string]string)

		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return result, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
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
		return result, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
	}
	return result, nil
}

// OneRow returns the result of the query as one row
func (db *DCDB) OneRow(query string, args ...interface{}) *oneRow {
	result := make(map[string]string)
	//log.Debug("%v", query, args)
	all, err := db.GetAll(query, 1, args...)
	//log.Debug("%v", all)
	if err != nil {
		return &oneRow{result, fmt.Errorf("%s in query %s %s", err, query, args)}
	}
	if len(all) == 0 {
		return &oneRow{result, nil}
	}
	return &oneRow{all[0], nil}
}

// QueryRows returns the result of the query
func (db *DCDB) QueryRows(query string, args ...interface{}) (*sql.Rows, error) {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	return db.Query(newQuery, newArgs...)
}

// ExecSQL executes the query
func (db *DCDB) ExecSQL(query string, args ...interface{}) error {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	//var res sql.Result
	var err error
	for {
		log.Debug("newQuery: ", newQuery)
		_, err = db.Exec(newQuery, newArgs...)
		if err != nil {
			if ok, _ := regexp.MatchString(`(?i)database is locked`, fmt.Sprintf("%s", err)); ok {
				log.Error("database is locked %s / %s / %s", newQuery, newArgs, utils.GetParent())
				time.Sleep(250 * time.Millisecond)
				continue
			} else {
				return fmt.Errorf("%s in query %s %s %s", err, newQuery, newArgs, utils.GetParent())
			}
		} else {
			break
		}
	}
	return nil
}

// ExecSQLGetAffect executes the query and returns amount of affected rows
func (db *DCDB) ExecSQLGetAffect(query string, args ...interface{}) (int64, error) {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	var res sql.Result
	var err error
	for {
		res, err = db.Exec(newQuery, newArgs...)
		if err != nil {
			if ok, _ := regexp.MatchString(`(?i)database is locked`, fmt.Sprintf("%s", err)); ok {
				log.Error("database is locked %s / %s / %s", newQuery, newArgs, utils.GetParent())
				time.Sleep(250 * time.Millisecond)
				continue
			} else {
				return 0, fmt.Errorf("%s in query %s %s %s", err, newQuery, newArgs, utils.GetParent())
			}
		} else {
			break
		}
	}
	affect, err := res.RowsAffected()
	lastID, err := res.LastInsertId()
	if db.ConfigIni["sql_log"] == "1" {
		log.Debug("SQL: %s / RowsAffected=%d / LastInsertId=%d / %s", newQuery, affect, lastID, newArgs)
	}
	return affect, nil
}

// ExecSQLGetLastInsertID inserts a row and returns the last id
func (db *DCDB) ExecSQLGetLastInsertID(query, table string, args ...interface{}) (string, error) {
	var v interface{}
	var lastID string
	var err error
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	colName, err := db.GetFirstColumnNamesPg(table)
	if err != nil {
		return "", fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
	}
	newQuery = newQuery + " RETURNING " + colName
	for {
		err := db.QueryRow(newQuery, newArgs...).Scan(&v)
		if err != nil {
			if ok, _ := regexp.MatchString(`(?i)database is locked`, fmt.Sprintf("%s", err)); ok {
				log.Error("database is locked %s / %s / %s", newQuery, newArgs, utils.GetParent())
				time.Sleep(250 * time.Millisecond)
				continue
			} else {
				return "", fmt.Errorf("%s in query %s %s %s", err, newQuery, newArgs, utils.GetParent())
			}
		} else {
			switch v.(type) {
			case int:
				lastID = utils.IntToStr(v.(int))
			case int64:
				lastID = utils.Int64ToStr(v.(int64))
			case float64:
				lastID = utils.Float64ToStr(v.(float64))
			case string:
				lastID = v.(string)
			case []byte:
				lastID = string(v.([]byte))
			}
			break
		}
	}

	if db.ConfigIni["sql_log"] == "1" {
		log.Debug("SQL: %s / LastInsertId=%d / %s", newQuery, lastID, newArgs)
	}
	return lastID, nil
}
