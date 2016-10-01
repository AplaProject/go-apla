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

package utils

import (
	//	"crypto"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/lib"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/op/go-logging"
)

var Mutex = &sync.Mutex{}
var log = logging.MustGetLogger("daemons")
var DB *DCDB

type DCDB struct {
	*sql.DB
	ConfigIni map[string]string
	//GoroutineName string
}

func ReplQ(q string) string {
	q1 := strings.Split(q, "?")
	result := ""
	for i := 0; i < len(q1); i++ {
		if i != len(q1)-1 {
			result += q1[i] + "$" + IntToStr(i+1)
		} else {
			result += q1[i]
		}
	}
	//log.Debug("%v", result)
	return result
}

func NewDbConnect(ConfigIni map[string]string) (*DCDB, error) {
	var db *sql.DB
	var err error
	switch ConfigIni["db_type"] {
	case "sqlite":

		log.Debug("sqlite connect")
		db, err = sql.Open("sqlite3", *Dir+"/litedb.db")
		log.Debug("%v", db)
		if err != nil {
			log.Debug("%v", err)
			return &DCDB{}, err
		}
		ddl := `
				PRAGMA synchronous = NORMAL;
				PRAGMA journal_mode = WAL;
				PRAGMA encoding = "UTF-8";
				`
		log.Debug("Exec ddl0")
		_, err = db.Exec(ddl)
		log.Debug("Exec ddl")
		if err != nil {
			log.Debug("%v", ErrInfo(err))
			db.Close()
			return &DCDB{}, err
		}
	case "postgresql":
		db, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=%s", ConfigIni["db_user"], ConfigIni["db_password"], ConfigIni["db_name"], ConfigIni["db_port"]))
		if err != nil {
			return &DCDB{}, err
		}
	case "mysql":
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", ConfigIni["db_user"], ConfigIni["db_password"], ConfigIni["db_name"]))
		if err != nil {
			return &DCDB{}, err
		}
	}
	log.Debug("return")
	return &DCDB{db, ConfigIni}, err
}

func (db *DCDB) GetConfigIni(name string) string {
	return db.ConfigIni[name]
}

func (db *DCDB) GetMainLockName() (string, error) {
	return db.Single("SELECT script_name FROM main_lock").String()
}

func (db *DCDB) GetFirstColumnName(table string) (string, error) {
	rows, err := db.Query(`SELECT * FROM "` + table + `" LIMIT 1`)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	if len(columns) > 0 {
		return columns[0], nil
	}
	return "", nil
}

func (db *DCDB) GetAllTables() ([]string, error) {
	var result []string
	var sql string
	switch db.ConfigIni["db_type"] {
	case "sqlite":
		sql = "SELECT name FROM sqlite_master WHERE type IN ('table','view') AND name NOT LIKE 'sqlite_%'"
	case "postgresql":
		sql = "SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND    table_schema NOT IN ('pg_catalog', 'information_schema')"
	case "mysql":
		sql = "SHOW TABLES"
	}
	result, err := db.GetList(sql).String()
	if err != nil {
		return result, err
	}
	return result, nil
}

func (db *DCDB) GetFirstColumnNamesPg(table string) (string, error) {
	var result []string
	result, err := db.GetList("SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name='" + table + "'").String()
	if err != nil {
		return "", err
	}
	return result[0], nil
}

type singleResult struct {
	result []byte
	err    error
}

type listResult struct {
	result []string
	err    error
}

type oneRow struct {
	result map[string]string
	err    error
}

func (r *listResult) Int64() ([]int64, error) {
	var result []int64
	if r.err != nil {
		return result, r.err
	}
	for _, v := range r.result {
		result = append(result, StrToInt64(v))
	}
	return result, nil
}

func (r *listResult) MapInt() (map[int]int, error) {
	result := make(map[int]int)
	if r.err != nil {
		return result, r.err
	}
	i := 0
	for _, v := range r.result {
		result[i] = StrToInt(v)
		i++
	}
	return result, nil
}

func (r *listResult) String() ([]string, error) {
	if r.err != nil {
		return r.result, r.err
	}
	return r.result, nil
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
		result[k] = StrToInt64(v)
	}
	return result, nil
}

func (r *oneRow) Float64() (map[string]float64, error) {
	result := make(map[string]float64)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = StrToFloat64(v)
	}
	return result, nil
}

func (r *oneRow) Int() (map[string]int, error) {
	result := make(map[string]int)
	if r.err != nil {
		return result, r.err
	}
	for k, v := range r.result {
		result[k] = StrToInt(v)
	}
	return result, nil
}

func (r *singleResult) Int64() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return BytesToInt64(r.result), nil
}
func (r *singleResult) Int() (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	return BytesToInt(r.result), nil
}

func (r *singleResult) Float64() (float64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return StrToFloat64(string(r.result)), nil
}

func (r *singleResult) String() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return string(r.result), nil
}

func (r *singleResult) Bytes() ([]byte, error) {
	if r.err != nil {
		return []byte(""), r.err
	}
	return r.result, nil
}

func (db *DCDB) Single(query string, args ...interface{}) *singleResult {

	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)

	var result []byte
	err := db.QueryRow(newQuery, newArgs...).Scan(&result)
	switch {
	case err == sql.ErrNoRows:
		return &singleResult{[]byte(""), nil}
	case err != nil:
		return &singleResult{[]byte(""), fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)}
	}
	if db.ConfigIni["sql_log"] == "1" {
		/*parent := ""
		for i:=2;;i++{
			name := ""
			if pc, _, _, ok := runtime.Caller(i); ok {
				name = filepath.Base(runtime.FuncForPC(pc).Name())
				file, line := runtime.FuncForPC(pc).FileLine(pc)
				if i > 5 || name == "runtime.goexit" {
					break
				} else {
					parent += fmt.Sprintf("%s:%d -> %s / ", filepath.Base(file), line, name, parent)
				}
			}
		}
		*/
		parent := GetParent()
		log.Debug("SQL: %s / %v / %v", newQuery, newArgs, parent)
	}
	return &singleResult{result, nil}
}

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

func (db *DCDB) GetList(query string, args ...interface{}) *listResult {
	var result []string
	all, err := db.GetAll(query, -1, args...)
	if err != nil {
		return &listResult{result, err}
	}
	for _, v := range all {
		for _, v2 := range v {
			result = append(result, v2)
		}
	}
	return &listResult{result, nil}
}

func GetParent() string {
	parent := ""
	for i := 2; ; i++ {
		name := ""
		if pc, _, num, ok := runtime.Caller(i); ok {
			name = filepath.Base(runtime.FuncForPC(pc).Name())
			file, line := runtime.FuncForPC(pc).FileLine(pc)
			if i > 5 || name == "runtime.goexit" {
				break
			} else {
				parent += fmt.Sprintf("%s:%d -> %s:%d / ", filepath.Base(file), line, name, num)
			}
		}
	}
	return parent
}

func (db *DCDB) GetAll(query string, countRows int, args ...interface{}) ([]map[string]string, error) {

	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)

	if db.ConfigIni["db_type"] == "postgresql" {
		query = ReplQ(query)
	}
	var result []map[string]string
	// Execute the query
	//fmt.Println("query", query)
	rows, err := db.Query(newQuery, newArgs...)
	if err != nil {
		return result, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
	}
	defer rows.Close()

	if db.ConfigIni["sql_log"] == "1" {
		/*parent := ""
		for i:=2;;i++{
			name := ""
			if pc, _, _, ok := runtime.Caller(i); ok {
				name = filepath.Base(runtime.FuncForPC(pc).Name())
				file, line := runtime.FuncForPC(pc).FileLine(pc)
				if i > 5 || name == "runtime.goexit" {
					break
				} else {
					parent += fmt.Sprintf("%s:%d -> %s / ", filepath.Base(file), line, name)
				}
			}
		}*/
		parent := GetParent()
		log.Debug("SQL: %s / %v / %v", newQuery, newArgs, parent)
	}
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return result, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
	}
	//fmt.Println("columns", columns)

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

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
			//fmt.Println(columns[i], ": ", value)
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
	//fmt.Println(result)
	return result, nil
}

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

func (db *DCDB) InsertInLogTx(binaryTx []byte, time int64) error {
	txMD5 := Md5(binaryTx)
	err := db.ExecSql("INSERT INTO log_transactions (hash, time) VALUES ([hex], ?)", txMD5, time)
	log.Debug("INSERT INTO log_transactions (hash, time) VALUES ([hex], %s)", txMD5)
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

func (db *DCDB) DelLogTx(binaryTx []byte) error {
	txMD5 := Md5(binaryTx)
	affected, err := db.ExecSqlGetAffect("DELETE FROM log_transactions WHERE hex(hash) = ?", txMD5)
	log.Debug("DELETE FROM log_transactions WHERE hex(hash) = %s / affected = %d", txMD5, affected)
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

func (db *DCDB) GetJSON(query string, args ...interface{}) (string, error) {

	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)

	rows, err := db.Query(newQuery, newArgs...)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}
	jsonData, err := json.Marshal(tableData)
	if err != nil {
		return "", err
	}
	fmt.Println(string(jsonData))
	return string(jsonData), nil
}

func (db *DCDB) QueryRows(query string, args ...interface{}) (*sql.Rows, error) {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	return db.Query(newQuery, newArgs...)
}
func (db *DCDB) ExecSqlGetLastInsertId(query, table string, args ...interface{}) (string, error) {
	var lastId_ interface{}
	var lastId string
	var err error
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	colName, err := db.GetFirstColumnNamesPg(table)
		if err != nil {
			return "", fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
		}
		newQuery = newQuery + " RETURNING " + colName
		for {
			err := db.QueryRow(newQuery, newArgs...).Scan(&lastId_)
			if err != nil {
				if ok, _ := regexp.MatchString(`(?i)database is locked`, fmt.Sprintf("%s", err)); ok {
					log.Error("database is locked %s / %s / %s", newQuery, newArgs, GetParent())
					time.Sleep(250 * time.Millisecond)
					continue
				} else {
					return "", fmt.Errorf("%s in query %s %s %s", err, newQuery, newArgs, GetParent())
				}
			} else {
				switch lastId_.(type) {
					case int:
					lastId = IntToStr(lastId_.(int))
					case int64:
					lastId = Int64ToStr(lastId_.(int64))
					case float64:
					lastId = Float64ToStr(lastId_.(float64))
					case string:
					lastId = lastId_.(string)
					case []byte:
					lastId = string(lastId_.([]byte))
				}
				break
			}
		}

	if db.ConfigIni["sql_log"] == "1" {
		log.Debug("SQL: %s / LastInsertId=%d / %s", newQuery, lastId, newArgs)
	}
	return lastId, nil
}

func FormatQueryArgs(q, dbType string, args ...interface{}) (string, []interface{}) {
	var newArgs []interface{}
	newQ := q
	if ok, _ := regexp.MatchString(`CREATE TABLE`, newQ); !ok {
		switch dbType {
		case "sqlite":
			//log.Debug(q)
			r, _ := regexp.Compile(`(\[hex\]|\?)`)
			indexArr := r.FindAllStringSubmatchIndex(q, -1)
			//log.Debug("indexArr %v", indexArr)
			for i := 0; i < len(indexArr); i++ {
				str := q[indexArr[i][0]:indexArr[i][1]]
				//log.Debug("i: %v, len: %v str: %v, q: %v", i, len(args), str, q)
				if str != "[hex]" {
					switch args[i].(type) {
					case []byte:
						newArgs = append(newArgs, string(args[i].([]byte)))
					default:
						newArgs = append(newArgs, args[i])
					}
				} else {
					switch args[i].(type) {
					case string:
						newQ = strings.Replace(newQ, "[hex]", "x'"+args[i].(string)+"'", 1)
					case []byte:
						newQ = strings.Replace(newQ, "[hex]", "x'"+string(args[i].([]byte))+"'", 1)
					}
				}
			}
			newQ = strings.Replace(newQ, "[hex]", "?", -1)
		//log.Debug("%v", "newQ", newQ)
		case "postgresql":
			newQ = strings.Replace(newQ, "[hex]", "decode(?,'HEX')", -1)
			newQ = strings.Replace(newQ, " authorization", ` "authorization"`, -1)
			newQ = strings.Replace(newQ, "user,", `"user",`, -1)
			newQ = ReplQ(newQ)
			newArgs = args
		case "mysql":
			newQ = strings.Replace(newQ, "[hex]", "UNHEX(?)", -1)
			newQ = strings.Replace(newQ, "lock,", "`lock`,", -1)
			newQ = strings.Replace(newQ, " lock ", " `lock` ", -1)
			newArgs = args
		}
	}
	if dbType == "postgresql" || dbType == "sqlite" {
		r, _ := regexp.Compile(`\s*([0-9]+_[\w]+)(?:\.|\s|\)|$)`)
		indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
		for i := len(indexArr) - 1; i >= 0; i-- {
			newQ = newQ[:indexArr[i][2]] + `"` + newQ[indexArr[i][2]:indexArr[i][3]] + `"` + newQ[indexArr[i][3]:]
		}
	}

	r, _ := regexp.Compile(`hex\(([\w]+)\)`)
	indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
	for i := len(indexArr) - 1; i >= 0; i-- {
		if dbType == "mysql" || dbType == "sqlite" {
			newQ = newQ[:indexArr[i][0]] + `LOWER(HEX(` + newQ[indexArr[i][2]:indexArr[i][3]] + `))` + newQ[indexArr[i][1]:]
		} else {
			newQ = newQ[:indexArr[i][0]] + `LOWER(encode(` + newQ[indexArr[i][2]:indexArr[i][3]] + `, 'hex'))` + newQ[indexArr[i][1]:]
		}
	}

	return newQ, newArgs
}

func (db *DCDB) CheckInstall(DaemonCh chan bool, AnswerDaemonCh chan string, GoroutineName string) bool {
	// Возможна ситуация, когда инсталяция еще не завершена. База данных может быть создана, а таблицы еще не занесены
	for {
		select {
		case <-DaemonCh:
			log.Debug("Restart from CheckInstall")
			AnswerDaemonCh <- GoroutineName
			return false
		default:
		}
		progress, err := db.Single("SELECT progress FROM install").String()
		if err != nil || progress != "complete" {
			// возможно попасть на тот момент, когда БД закрыта и идет скачивание готовой БД с сервера
			if ok, _ := regexp.MatchString(`database is closed`, fmt.Sprintf("%s", err)); ok {
				if DB != nil {
					db = DB
				}
			}
			//log.Debug("%v", `progress != "complete"`, db.GoroutineName)
			if err != nil {
				log.Error("%v", ErrInfo(err))
			}
			Sleep(1)
		} else {
			break
		}
	}
	return true
}

func (db *DCDB) GetQuotes() string {
	dq := `"`
	if db.ConfigIni["db_type"] == "mysql" {
		dq = ``
	}
	return dq
}

func (db *DCDB) ExecSql(query string, args ...interface{}) error {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	var res sql.Result
	var err error
	for {
		res, err = db.Exec(newQuery, newArgs...)
		if err != nil {
			if ok, _ := regexp.MatchString(`(?i)database is locked`, fmt.Sprintf("%s", err)); ok {
				log.Error("database is locked %s / %s / %s", newQuery, newArgs, GetParent())
				time.Sleep(250 * time.Millisecond)
				continue
			} else {
				return fmt.Errorf("%s in query %s %s %s", err, newQuery, newArgs, GetParent())
			}
		} else {
			break
		}
	}
	affect, err := res.RowsAffected()
	lastId, err := res.LastInsertId()
	if db.ConfigIni["sql_log"] == "1" {
		parent := GetParent()
		log.Debug("SQL: %v / RowsAffected=%d / LastInsertId=%d / %s / %s", newQuery, affect, lastId, newArgs, parent)
	}
	return nil
}

func (db *DCDB) ExecSqlGetAffect(query string, args ...interface{}) (int64, error) {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	var res sql.Result
	var err error
	for {
		res, err = db.Exec(newQuery, newArgs...)
		if err != nil {
			if ok, _ := regexp.MatchString(`(?i)database is locked`, fmt.Sprintf("%s", err)); ok {
				log.Error("database is locked %s / %s / %s", newQuery, newArgs, GetParent())
				time.Sleep(250 * time.Millisecond)
				continue
			} else {
				return 0, fmt.Errorf("%s in query %s %s %s", err, newQuery, newArgs, GetParent())
			}
		} else {
			break
		}
	}
	affect, err := res.RowsAffected()
	lastId, err := res.LastInsertId()
	if db.ConfigIni["sql_log"] == "1" {
		log.Debug("SQL: %s / RowsAffected=%d / LastInsertId=%d / %s", newQuery, affect, lastId, newArgs)
	}
	return affect, nil
}

// для юнит-тестов. снимок всех данных в БД
func (db *DCDB) HashTableData(table, where, orderBy string) (string, error) {
	/*var columns string;
	rows, err := db.Query("select column_name from information_schema.columns where table_name= $1", table)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return "", err
		}
		columns+=name+"+"
	}
	columns = columns[:len(columns)-1]

	if len(columns) > 0 {
		if len(orderBy) > 0 {
			orderBy = " ORDER BY "+orderBy;
		}
	}*/
	if len(orderBy) > 0 {
		orderBy = " ORDER BY " + orderBy
	}

	// это у всех разное, а значит и хэши будут разные, а это будет вызывать путаницу
	var logOff bool
	if db.ConfigIni["sql_log"] == "1" {
		db.ConfigIni["sql_log"] = "0"
		logOff = true
	}

	var err error
	var hash string
	switch db.ConfigIni["db_type"] {
	case "sqlite":
		//q = "SELECT md5(CAST((array_agg(t.* " + orderBy + ")) AS text)) FROM \"" + table + "\" t " + where
	case "postgresql":
		q := "SELECT md5(CAST((array_agg(t.* " + orderBy + ")) AS text)) FROM \"" + table + "\" t " + where
		hash, err = db.Single(q).String()
		if err != nil {
			return "", ErrInfo(err, q)
		}
	case "mysql":
		err := db.ExecSql("SET @@group_concat_max_len = 4294967295")
		if err != nil {
			return "", ErrInfo(err)
		}
		columns, err := db.Single("SELECT GROUP_CONCAT( column_name SEPARATOR '`,`' ) FROM information_schema.columns WHERE table_schema = ? AND table_name = ?", db.ConfigIni["db_name"], table).String()
		if err != nil {
			return "", ErrInfo(err)
		}
		columns = strings.Replace(columns, ",`status_backup`", "", -1)
		columns = strings.Replace(columns, "`status_backup`,", "", -1)
		columns = strings.Replace(columns, ",`cash_request_in_block_id`", "", -1)
		columns = strings.Replace(columns, "`cash_request_in_block_id`,", "", -1)
		q := "SELECT MD5(GROUP_CONCAT( CONCAT_WS( '#', `" + columns + "`)  " + orderBy + " )) FROM `" + table + "` " + where
		log.Debug("%v", q)
		hash, err = db.Single(q).String()
		if err != nil {
			return "", ErrInfo(err, q)
		}
	}
	//fmt.Println(q)

	/*if strings.Count(table, "my_table")>0 {
		columns = strings.Replace(columns,",notification","",-1)
		columns = strings.Replace(columns,"notification,","",-1)
		q="SELECT md5(CAST((array_agg("+columns+" "+orderBy+")) AS text)) FROM \""+table+"\" "+where
	}*/
	/*if strings.Count(columns, "cron_checked_time")>0 {
		columns = strings.Replace(columns, ",cron_checked_time", "", -1)
		columns = strings.Replace(columns, "cron_checked_time,", "", -1)
		q="SELECT md5(CAST((array_agg("+columns+" "+orderBy+")) AS text)) FROM \""+table+"\" "+where
	}*/

	if logOff {
		db.ConfigIni["sql_log"] = "1"
	}
	return hash, nil
}

func (db *DCDB) GetLastBlockData() (map[string]int64, error) {
	result := make(map[string]int64)
	confirmedBlockId, err := db.GetConfirmedBlockId()
	if err != nil {
		return result, ErrInfo(err)
	}
	if confirmedBlockId == 0 {
		confirmedBlockId = 1
	}
	log.Debug("%v", "confirmedBlockId", confirmedBlockId)
	// получим время из последнего подвержденного блока
	lastBlockBin, err := db.Single("SELECT data FROM block_chain WHERE id = ?", confirmedBlockId).Bytes()
	if err != nil || len(lastBlockBin) == 0 {
		return result, ErrInfo(err)
	}
	// ID блока
	result["blockId"] = int64(BinToDec(lastBlockBin[1:5]))
	// Время последнего блока
	result["lastBlockTime"] = int64(BinToDec(lastBlockBin[5:9]))
	return result, nil
}

func (db *DCDB) GetMyPublicKey(myPrefix string) ([]byte, error) {
	result, err := db.Single("SELECT public_key FROM my_keys WHERE block_id = (SELECT max(block_id) FROM my_keys)").Bytes()
	if err != nil {
		return []byte(""), ErrInfo(err)
	}
	return result, nil
}

func (db *DCDB) GetMyPrivateKey(myPrefix string) (string, error) {
	key, err := db.Single("SELECT private_key FROM my_keys WHERE block_id = (SELECT max(block_id) FROM my_keys)").String()
	if err != nil {
		return "", ErrInfo(err)
	}
	key = strings.Replace(key, "-----BEGIN RSA PRIVATE KEY-----", "-----BEGIN RSA PRIVATE KEY-----\n", -1)
	key = strings.Replace(key, "-----END RSA PRIVATE KEY-----", "\n-----END RSA PRIVATE KEY-----", -1)
	return key, nil
}

func (db *DCDB) GetNodePrivateKey() (string, error) {
	var key string
	key, err := db.Single("SELECT private_key FROM my_node_keys WHERE block_id = (SELECT max(block_id) FROM my_node_keys)").String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return key, nil
}

func (db *DCDB) GetMyNodePublicKey(myPrefix string) (string, error) {
	var key string
	key, err := db.Single("SELECT public_key FROM my_node_keys WHERE block_id = (SELECT max(block_id) FROM my_node_keys)").String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return key, nil
}

func (db *DCDB) GetPrivateKey(myPrefix string) (string, error) {
	var key string
	key, err := db.Single("SELECT private_key FROM my_keys WHERE block_id = (SELECT max(block_id) FROM my_keys)").String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return key, nil
}

func (db *DCDB) GetNodeConfig() (map[string]string, error) {
	return db.OneRow("SELECT * FROM config").String()
}

func (db *DCDB) FormatQuery(q string) string {

	newQ := q
	if ok, _ := regexp.MatchString(`CREATE TABLE`, newQ); !ok {
		switch db.ConfigIni["db_type"] {
		case "sqlite":
			newQ = strings.Replace(newQ, "[hex]", "?", -1)
			newQ = strings.Replace(newQ, "user,", "`user`,", -1)
			newQ = strings.Replace(newQ, ", user ", ", `user` ", -1)
		case "postgresql":
			newQ = strings.Replace(newQ, "[hex]", "decode(?,'HEX')", -1)
			newQ = strings.Replace(newQ, " authorization", ` "authorization"`, -1)
			newQ = strings.Replace(newQ, "user,", `"user",`, -1)
			newQ = strings.Replace(newQ, ", user ", `, "user" `, -1)
			newQ = ReplQ(newQ)
		case "mysql":
			newQ = strings.Replace(newQ, "[hex]", "UNHEX(?)", -1)
		}
	}

	if db.ConfigIni["db_type"] == "postgresql" || db.ConfigIni["db_type"] == "sqlite" {
		r, _ := regexp.Compile(`\s*([0-9]+_[\w]+)(?:\.|\s|\)|$)`)
		indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
		for i := len(indexArr) - 1; i >= 0; i-- {
			newQ = newQ[:indexArr[i][2]] + `"` + newQ[indexArr[i][2]:indexArr[i][3]] + `"` + newQ[indexArr[i][3]:]
		}
	}

	r, _ := regexp.Compile(`hex\(([\w]+)\)`)
	indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
	for i := len(indexArr) - 1; i >= 0; i-- {
		if db.ConfigIni["db_type"] == "mysql" || db.ConfigIni["db_type"] == "sqlite" {
			newQ = newQ[:indexArr[i][0]] + `LOWER(HEX(` + newQ[indexArr[i][2]:indexArr[i][3]] + `))` + newQ[indexArr[i][1]:]
		} else {
			newQ = newQ[:indexArr[i][0]] + `LOWER(encode(` + newQ[indexArr[i][2]:indexArr[i][3]] + `, 'hex'))` + newQ[indexArr[i][1]:]
		}
	}

	log.Debug("%v", newQ)
	return newQ
}

func (db *DCDB) GetConfirmedBlockId() (int64, error) {

	result, err := db.Single("SELECT max(block_id) FROM confirmations WHERE good >= ?", consts.MIN_CONFIRMED_NODES).Int64()
	if err != nil {
		return 0, err
	}
	//log.Debug("%v", "result int64",StrToInt64(result))
	return result, nil

}

func (db *DCDB) GetMyCBIDAndWalletId() (int64, int64, error) {
	myCBID, err := db.GetMyCBID()
	if err != nil {
		return 0, 0, err
	}
	myWalletId, err := db.GetMyWalletId()
	if err != nil {
		return 0, 0, err
	}
	return myCBID, myWalletId, nil
}

func (db *DCDB) GetHosts() ([]string, error) {
	q := ""
	if db.ConfigIni["db_type"] == "postgresql" {
		q = "SELECT DISTINCT ON (host) host FROM full_nodes"
	} else {
		q = "SELECT host FROM full_nodes GROUP BY host"
	}
	hosts, err := db.GetList(q).String()
	if err != nil {
		return nil, err
	}
	return hosts, nil
}

func (db *DCDB) CheckDelegateCB(myCBID int64) (bool, error) {
	delegate, err := db.OneRow("SELECT delegate_wallet_id, delegate_state_id FROM central_banks WHERE state_id = ?", myCBID).Int64()
	if err != nil {
		return false, err
	}
	// Если мы - ЦБ и у нас указан delegate, т.е. мы делегировали полномочия по поддержанию ноды другому юзеру или ЦБ, то выходим.
	if delegate["delegate_wallet_id"] > 0 || delegate["delegate_state_id"] > 0 {
		return true, nil
	}
	return false, nil
}

func (db *DCDB) GetMyWalletId() (int64, error) {
	return db.Single("SELECT dlt_wallet_id FROM config").Int64()
}

func (db *DCDB) GetMyCBID() (int64, error) {
	return db.Single("SELECT state_id FROM config").Int64()
}

func (db *DCDB) GetBlockId() (int64, error) {
	return db.Single("SELECT block_id FROM info_block").Int64()
}

func (db *DCDB) GetMyBlockId() (int64, error) {
	return db.Single("SELECT my_block_id FROM config").Int64()
}

func (db *DCDB) GetWalletIdByPublicKey(publicKey []byte) (int64, error) {
	log.Debug("string(HashSha1Hex(publicKey) %s", string(HashSha1Hex(publicKey)))
	log.Debug("publicKey %s", publicKey)
	key, _ := hex.DecodeString(string(publicKey))
	walletId, err := db.Single(`SELECT wallet_id FROM dlt_wallets WHERE lower(hex(address)) = ?`,
		string( /*HashSha1Hex*/ hex.EncodeToString(lib.Address(key)))).Int64()
	if err != nil {
		return 0, ErrInfo(err)
	}
	return walletId, nil
}

func (db *DCDB) GetCitizenIdByPublicKey(publicKey []byte) (int64, error) {
	walletId, err := db.Single(`SELECT citizen_id FROM ea_citizens WHERE hex(public_key_0) = ?`, string(publicKey)).Int64()
	if err != nil {
		return 0, ErrInfo(err)
	}
	return walletId, nil
}

func (db *DCDB) GetInfoBlock() (map[string]string, error) {
	var result map[string]string
	result, err := db.OneRow("SELECT * FROM info_block").String()
	if err != nil {
		return result, ErrInfo(err)
	}
	if len(result) == 0 {
		return result, fmt.Errorf("empty info_block")
	}
	return result, nil
}

func (db *DCDB) GetNodePublicKey(userId int64) ([]byte, error) {
	result, err := db.Single("SELECT node_public_key FROM miners_data WHERE user_id = ?", userId).Bytes()
	if err != nil {
		return []byte(""), err
	}
	return result, nil
}
func (db *DCDB) GetNodePublicKeyWalletOrCB(wallet_id, state_id int64) ([]byte, error) {
	var result []byte
	var err error
	if wallet_id > 0 {
		log.Debug("wallet_id %v state_id %v", wallet_id, state_id)
		result, err = db.Single("SELECT node_public_key FROM dlt_wallets WHERE wallet_id = ?", wallet_id).Bytes()
		if err != nil {
			return []byte(""), err
		}
	} else {
		result, err = db.Single("SELECT node_public_key FROM central_banks WHERE state_id = ?", state_id).Bytes()
		if err != nil {
			return []byte(""), err
		}
	}
	return result, nil
}

func (db *DCDB) GetPublicKeyWalletOrCitizen(wallet_id, citizen_id int64) ([]byte, error) {
	var result []byte
	var err error
	if wallet_id > 0 {
		result, err = db.Single("SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?", wallet_id).Bytes()
		if err != nil {
			return []byte(""), err
		}
	} else {
		result, err = db.Single("SELECT public_key_0 FROM ea_citizens WHERE citizen_is = ?", citizen_id).Bytes()
		if err != nil {
			return []byte(""), err
		}
	}
	return result, nil
}

func (db *DCDB) UpdMainLock() error {
	return db.ExecSql("UPDATE main_lock SET lock_time = ?", time.Now().Unix())
}

func (db *DCDB) CheckDaemonsRestart() bool {
	return false
}

func (db *DCDB) DbLock(DaemonCh chan bool, AnswerDaemonCh chan string, goRoutineName string) (error, bool) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()

	log.Debug("DbLock")
	var ok bool
	for {
		select {
		case <-DaemonCh:
			log.Debug("Restart from DbLock")
			AnswerDaemonCh <- goRoutineName
			return ErrInfo("Restart from DbLock"), true
		default:
		}

		Mutex.Lock()

		log.Debug("DbLock Mutex.Lock()")

		exists, err := db.OneRow("SELECT lock_time, script_name FROM main_lock").String()
		if err != nil {
			Mutex.Unlock()
			return ErrInfo(err), false
		}
		if len(exists["script_name"]) == 0 {
			err = db.ExecSql(`INSERT INTO main_lock(lock_time, script_name, info) VALUES(?, ?, ?)`, time.Now().Unix(), goRoutineName, Caller(2))
			if err != nil {
				Mutex.Unlock()
				return ErrInfo(err), false
			}
			ok = true
		} else {
			t := StrToInt64(exists["lock_time"])
			if Time()-t > 600 {
				log.Error("%d %s %d", t, exists["script_name"], Time()-t)
				if Mobile() {
					db.ExecSql(`DELETE FROM main_lock`)
				}
			}
		}
		Mutex.Unlock()
		if !ok {
			time.Sleep(time.Duration(RandInt(300, 400)) * time.Millisecond)
		} else {
			break
		}
	}
	return nil, false
}

func (db *DCDB) DbLockGate(name string) error {
	var ok bool
	for {
		Mutex.Lock()
		exists, err := db.OneRow("SELECT lock_time, script_name FROM main_lock").String()
		if err != nil {
			Mutex.Unlock()
			return ErrInfo(err)
		}
		if len(exists["script_name"]) == 0 {
			err = db.ExecSql(`INSERT INTO main_lock(lock_time, script_name, info) VALUES(?, ?, ?)`, time.Now().Unix(), name, Caller(1))
			if err != nil {
				Mutex.Unlock()
				return ErrInfo(err)
			}
			ok = true
		}
		Mutex.Unlock()
		if !ok {
			time.Sleep(time.Duration(RandInt(300, 400)) * time.Millisecond)
		} else {
			break
		}
	}
	return nil
}

func (db *DCDB) DeleteQueueBlock(hash_hex string) error {
	return db.ExecSql("DELETE FROM queue_blocks WHERE hex(hash) = ?", hash_hex)
}

func (db *DCDB) SetAI(table string, AI int64) error {

	AiId, err := db.GetAiId(table)
	if err != nil {
		return ErrInfo(err)
	}

	if db.ConfigIni["db_type"] == "postgresql" {
		pg_get_serial_sequence, err := db.Single("SELECT pg_get_serial_sequence('" + table + "', '" + AiId + "')").String()
		if err != nil {
			return ErrInfo(err)
		}
		err = db.ExecSql("ALTER SEQUENCE " + pg_get_serial_sequence + " RESTART WITH " + Int64ToStr(AI))
		if err != nil {
			return ErrInfo(err)
		}
	} else if db.ConfigIni["db_type"] == "mysql" {
		err := db.ExecSql("ALTER TABLE " + table + " AUTO_INCREMENT = " + Int64ToStr(AI))
		if err != nil {
			return ErrInfo(err)
		}
	} else if db.ConfigIni["db_type"] == "sqlite" {
		err := db.ExecSql("UPDATE SQLITE_SEQUENCE SET seq = ? WHERE name = ?", AI, table)
		if err != nil {
			return ErrInfo(err)
		}
	}
	return nil
}

func (db *DCDB) PrintSleep(err_ interface{}, sleep time.Duration) {
	var err error
	switch err_.(type) {
	case string:
		err = errors.New(err_.(string))
	case error:
		err = err_.(error)
	}
	log.Error("%v (%v)", err, GetParent())
	Sleep(sleep)
}

func (db *DCDB) PrintSleepInfo(err_ interface{}, sleep time.Duration) {
	var err error
	switch err_.(type) {
	case string:
		err = errors.New(err_.(string))
	case error:
		err = err_.(error)
	}
	log.Info("%v (%v)", err, GetParent())
	Sleep(sleep)
}

func (db *DCDB) DbUnlock(goRoutineName string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()
	log.Debug("DbUnlock %v %v", Caller(2), goRoutineName)
	affect, err := db.ExecSqlGetAffect("DELETE FROM main_lock WHERE script_name = ?", goRoutineName)
	log.Debug("main_lock affect: %d, goRoutineName: %s", affect, goRoutineName)
	if err != nil {
		log.Error("%s", ErrInfo(err))
		return ErrInfo(err)
	}
	return nil
}

func (db *DCDB) DbUnlockGate(name string) error {
	log.Debug("DbUnlockGate %v %v", Caller(2), name)
	return db.ExecSql("DELETE FROM main_lock WHERE script_name = ?", name)
}

func (db *DCDB) UpdDaemonTime(name string) {

}

func (db *DCDB) GetAiId(table string) (string, error) {
	exists := ""
	column := "id"
	if table == "users" {
		column = "user_id"
	} else if table == "miners" {
		column = "miner_id"
	} else {
		switch db.ConfigIni["db_type"] {
		case "sqlite":
			err := db.QueryRow(db.FormatQuery("SELECT id FROM " + table)).Scan(&exists)
			if err != nil {
				if fmt.Sprintf("%x", err) == fmt.Sprintf("%x", fmt.Errorf("no such column: id")) {
					err = db.QueryRow(db.FormatQuery("SELECT rb_id FROM " + table)).Scan(&exists)
					if err != nil {
						if ok, _ := regexp.MatchString(`no rows`, fmt.Sprintf("%s", err)); ok {
							column = "rb_id"
						} else {
							return "", ErrInfo(err)
						}
					}
					column = "rb_id"
				} else {
					if ok, _ := regexp.MatchString(`no rows`, fmt.Sprintf("%s", err)); ok {
						column = "id"
					} else {
						return "", ErrInfo(err)
					}
				}
			}
		case "postgresql":
			exists = ""
			err := db.QueryRow("SELECT column_name FROM information_schema.columns WHERE table_name=$1 and column_name=$2", table, "id").Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				return "", err
			}
			if len(exists) == 0 {
				err := db.QueryRow("SELECT column_name FROM information_schema.columns WHERE table_name=$1 and column_name=$2", table, "rb_id").Scan(&exists)
				if err != nil {
					return "", err
				}
				if len(exists) == 0 {
					return "", fmt.Errorf("no id, rb_id")
				}
				column = "rb_id"
			}
		case "mysql":
			exists = ""
			err := db.QueryRow("SELECT TABLE_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name=? and column_name=?", db.ConfigIni["db_name"], table, "id").Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				return "", err
			}
			if len(exists) == 0 {
				err := db.QueryRow("SELECT TABLE_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name=? and column_name=?", db.ConfigIni["db_name"], table, "rb_id").Scan(&exists)
				if err != nil {
					return "", err
				}
				if len(exists) == 0 {
					return "", fmt.Errorf("no id, rb_id")
				}
				column = "rb_id"
			}
		}
	}
	return column, nil
}

func (db *DCDB) NodesBan(info string) error {

	return nil
}

func (db *DCDB) GetBlockDataFromBlockChain(blockId int64) (*BlockData, error) {
	BlockData := new(BlockData)
	data, err := db.OneRow("SELECT * FROM block_chain WHERE id = ?", blockId).String()
	if err != nil {
		return BlockData, ErrInfo(err)
	}
	log.Debug("data: %x\n", data["data"])
	if len(data["data"]) > 0 {
		binaryData := []byte(data["data"])
		BytesShift(&binaryData, 1) // не нужно. 0 - блок, >0 - тр-ии
		BlockData = ParseBlockHeader(&binaryData)
		BlockData.Hash = BinToHex([]byte(data["hash"]))
	}
	return BlockData, nil
}

func (db *DCDB) ClearIncompatibleTxSql(whereType interface{}, walletId int64, citizenId int64, waitError *string) {
	var whereTypeID int64
	switch whereType.(type) {
	case string:
		whereTypeID = TypeInt(whereType.(string))
	case int64:
		whereTypeID = whereType.(int64)
	}
	addSql := ""
	if walletId > 0 {
		addSql = "AND wallet_id = " + Int64ToStr(walletId)
	}
	if citizenId > 0 {
		addSql = "AND citizen_id = " + Int64ToStr(citizenId)
	}
	num, err := db.Single(`
					SELECT count(*)
					FROM (
				            SELECT hash
				            FROM transactions
				            WHERE type = ?
				                          `+addSql+` AND
				                         verified=1 AND
				                         used = 0
					)  AS x
					`, whereTypeID, whereTypeID).Int64()
	if err != nil {
		*waitError = fmt.Sprintf("%v", ErrInfo(err))
	}
	if num > 0 {
		*waitError = "wait_error"
	}
}

func (db *DCDB) ClearIncompatibleTxSqlSet(typesArr []string, walletId_ interface{}, citizenId_ interface{}, waitError *string, thirdVar_ interface{}) error {

	var walletId int64
	switch walletId_.(type) {
	case string:
		walletId = StrToInt64(walletId_.(string))
	case int64:
		walletId = walletId_.(int64)
	}

	var citizenId int64
	switch citizenId_.(type) {
	case string:
		citizenId = StrToInt64(citizenId_.(string))
	case int64:
		citizenId = citizenId_.(int64)
	}

	var thirdVar string
	switch thirdVar_.(type) {
	case string:
		thirdVar = thirdVar_.(string)
	case int64:
		thirdVar = Int64ToStr(thirdVar_.(int64))
	}

	var whereType string
	for _, txType := range typesArr {
		whereType += Int64ToStr(TypeInt(txType)) + ","
	}
	whereType = whereType[:len(whereType)-1]

	addSql := ""
	if walletId > 0 {
		addSql = "AND wallet_id = " + Int64ToStr(walletId)
	}
	if citizenId > 0 {
		addSql = "AND citizen_id = " + Int64ToStr(citizenId)
	}

	addSql1 := ""
	if len(thirdVar) > 0 {
		addSql1 = "AND citizen_id = " + thirdVar
	}

	num, err := db.Single(`
					SELECT count(*)
					FROM (
				            SELECT hash
				            FROM transactions
				            WHERE type IN (`+whereType+`)
				                          `+addSql+` `+addSql1+` AND
				                         verified=1 AND
				                         used = 0
					)  AS x
					`, citizenId).Int64()
	if err != nil {
		*waitError = fmt.Sprintf("%v", ErrInfo(err))
	}
	if num > 0 {
		*waitError = "wait_error"
	}
	return nil
}

func GetTxTypeAndUserId(binaryBlock []byte) (txType int64, walletId int64, citizenId int64) {
	tmp := binaryBlock[:]
	txType = BinToDecBytesShift(&binaryBlock, 1)
	if consts.IsStruct(int(txType)) {
		var txHead consts.TxHeader
		lib.BinUnmarshal(&tmp, &txHead)
		walletId = txHead.WalletId
		citizenId = txHead.CitizenId
	} else {
		BytesShift(&binaryBlock, 4) // уберем время
		walletId = BytesToInt64(BytesShift(&binaryBlock, DecodeLength(&binaryBlock)))
		citizenId = BytesToInt64(BytesShift(&binaryBlock, DecodeLength(&binaryBlock)))
		// thirdVar - нужен тогда, когда нужно недопустить попадание в блок несовместимых тр-ий.
		// Например, удаление крауд-фандинг проекта и инвестирование в него средств.
	}
	return
}

func (db *DCDB) DecryptData(binaryTx *[]byte) ([]byte, []byte, []byte, error) {

	if len(*binaryTx) == 0 {
		return nil, nil, nil, ErrInfo("len(binaryTx) == 0")
	}

	// вначале пишется user_id, чтобы в режиме пула можно было понять, кому шлется и чей ключ использовать
	myUserId := BinToDecBytesShift(&*binaryTx, 5)
	log.Debug("myUserId: %d", myUserId)

	// изымем зашифрванный ключ, а всё, что останется в $binary_tx - сами зашифрованные хэши тр-ий/блоков
	encryptedKey := BytesShift(&*binaryTx, DecodeLength(&*binaryTx))
	log.Debug("encryptedKey: %x", encryptedKey)
	log.Debug("encryptedKey: %s", encryptedKey)

	// далее идет 16 байт IV
	iv := BytesShift(&*binaryTx, 16)
	log.Debug("iv: %s", iv)
	log.Debug("iv: %x", iv)

	if len(encryptedKey) == 0 {
		return nil, nil, nil, ErrInfo("len(encryptedKey) == 0")
	}

	if len(*binaryTx) == 0 {
		return nil, nil, nil, ErrInfo("len(*binaryTx) == 0")
	}

	nodePrivateKey, err := db.GetNodePrivateKey()
	if len(nodePrivateKey) == 0 {
		return nil, nil, nil, ErrInfo("len(nodePrivateKey) == 0")
	}

	block, _ := pem.Decode([]byte(nodePrivateKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, nil, nil, ErrInfo("No valid PEM data found")
	}

	private_key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, nil, ErrInfo(err)
	}

	decKey, err := rsa.DecryptPKCS1v15(crand.Reader, private_key, encryptedKey)
	if err != nil {
		return nil, nil, nil, ErrInfo(err)
	}
	log.Debug("decrypted Key: %s", decKey)
	if len(decKey) == 0 {
		return nil, nil, nil, ErrInfo("len(decKey)")
	}

	log.Debug("binaryTx %x", *binaryTx)
	log.Debug("iv %s", iv)
	decrypted, err := DecryptCFB(iv, *binaryTx, decKey)
	if err != nil {
		return nil, nil, nil, ErrInfo(err)
	}

	return decKey, iv, decrypted, nil
}

func (db *DCDB) FindInFullNodes(myCBID, myWalletId int64) (int64, error) {
	return db.Single("SELECT id FROM full_nodes WHERE final_delegate_state_id = ? OR final_delegate_wallet_id = ? OR state_id = ? OR wallet_id = ?", myCBID, myWalletId, myCBID, myWalletId).Int64()
}

func (db *DCDB) GetBinSign(forSign string) ([]byte, error) {
	nodePrivateKey, err := db.GetNodePrivateKey()
	if err != nil {
		return nil, ErrInfo(err)
	}
	/*	log.Debug("nodePrivateKey = %s", nodePrivateKey)
		// подписываем нашим нод-ключем данные транзакции
		privateKey, err := MakePrivateKey(nodePrivateKey)
		if err != nil {
			return nil, ErrInfo(err)
		}
		return rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA1, HashSha1(forSign))*/
	return SignECDSA(nodePrivateKey, forSign)
}

func (db *DCDB) InsertReplaceTxInQueue(data []byte) error {

	err := db.ExecSql("DELETE FROM queue_tx  WHERE hex(hash) = ?", Md5(data))
	if err != nil {
		return ErrInfo(err)
	}
	err = db.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", Md5(data), BinToHex(data))
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

func (db *DCDB) GetSleepTime(myWalletId, myCBID, prevBlockCBID, prevBlockWalletId int64) (int64, error) {
	// возьмем список всех full_nodes
	fullNodesList, err := db.GetAll("SELECT id, wallet_id, state_id as state_id FROM full_nodes", -1)
	if err != nil {
		return int64(0), ErrInfo(err)
	}
	log.Debug("fullNodesList %s", fullNodesList)

	// определим full_node_id того, кто должен был генерить блок (но мог это делегировать)
	prevBlockFullNodeId, err := db.Single("SELECT id FROM full_nodes WHERE state_id = ? OR wallet_id = ?", prevBlockCBID, prevBlockWalletId).Int64()
	if err != nil {
		return int64(0), ErrInfo(err)
	}
	log.Debug("prevBlockFullNodeId %d", prevBlockFullNodeId)

	log.Debug("%v %v", fullNodesList, prevBlockFullNodeId)

	prevBlockFullNodePosition := func(fullNodesList []map[string]string, prevBlockFullNodeId int64) int {
		for i, full_nodes := range fullNodesList {
			if StrToInt64(full_nodes["id"]) == prevBlockFullNodeId {
				return i
			}
		}
		return -1
	}(fullNodesList, prevBlockFullNodeId)
	log.Debug("prevBlockFullNodePosition %d", prevBlockFullNodePosition)

	// определим свое место (в том числе в delegate)
	myPosition := func(fullNodesList []map[string]string, myWalletId, myCBID int64) int {
		log.Debug("%v %v", fullNodesList, myWalletId)
		for i, full_nodes := range fullNodesList {
			if StrToInt64(full_nodes["state_id"]) == myCBID || StrToInt64(full_nodes["wallet_id"]) == myWalletId || StrToInt64(full_nodes["final_delegate_state_id"]) == myWalletId || StrToInt64(full_nodes["final_delegate_wallet_id"]) == myWalletId {
				return i
			}
		}
		return -1
	}(fullNodesList, myWalletId, myCBID)
	log.Debug("myPosition %d", myPosition)

	sleepTime := 0
	if myPosition == prevBlockFullNodePosition {
		sleepTime = ((len(fullNodesList) + myPosition) - int(prevBlockFullNodePosition)) * consts.GAPS_BETWEEN_BLOCKS
	}

	if myPosition > prevBlockFullNodePosition {
		sleepTime = (myPosition - int(prevBlockFullNodePosition)) * consts.GAPS_BETWEEN_BLOCKS
	}

	if myPosition < prevBlockFullNodePosition {
		sleepTime = (len(fullNodesList) - prevBlockFullNodePosition) * consts.GAPS_BETWEEN_BLOCKS
	}
	log.Debug("sleepTime %v / myPosition %v / prevBlockFullNodePosition %v / consts.GAPS_BETWEEN_BLOCKS %v", sleepTime, myPosition, prevBlockFullNodePosition, consts.GAPS_BETWEEN_BLOCKS)

	return int64(sleepTime), nil
}

func (db *DCDB) GetStateName(stateId int64) (string, error) {
	stateName, err := db.Single(`SELECT name FROM system_states WHERE id = ?`, stateId).String()
	if err != nil {
		return ``, err
	}
	return stateName, nil
}
