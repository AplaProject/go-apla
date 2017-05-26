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
	"encoding/pem"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	//	b58 "github.com/jbenet/go-base58"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	//	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/op/go-logging"
	"github.com/shopspring/decimal"
)

// Mutex for locking DB
var Mutex = &sync.Mutex{}
var log = logging.MustGetLogger("daemons")

// DB is a database variable
var DB *DCDB

//var cacheFuel int64
//var fuelMutex = &sync.Mutex{}

// DCDB is a database structure
type DCDB struct {
	*sql.DB
	ConfigIni map[string]string
	//GoroutineName string
}

// ReplQ preprocesses a database query
func ReplQ(q string) string {
	var quote, skip bool
	ind := 1
	in := []rune(q)
	out := make([]rune, 0, len(in)+16)
	for i, ch := range in {
		if skip {
			skip = false
		} else if ch == '\'' {
			if quote {
				if i == len(in)-1 || in[i+1] != '\'' {
					quote = false
				} else {
					skip = true
				}
			} else {
				quote = true
			}
		}
		if ch != '?' || quote {
			out = append(out, ch)
		} else {
			out = append(out, []rune(fmt.Sprintf(`$%d`, ind))...)
			ind++
		}
	}
	return string(out)
}

// NewDbConnect creates a new database connection
func NewDbConnect(ConfigIni map[string]string) (*DCDB, error) {
	var db *sql.DB
	var err error
	if len(ConfigIni["db_user"]) == 0 || len(ConfigIni["db_password"]) == 0 || len(ConfigIni["db_name"]) == 0 {
		return &DCDB{}, err
	}
	db, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=%s", ConfigIni["db_user"], ConfigIni["db_password"], ConfigIni["db_name"], ConfigIni["db_port"]))
	if err != nil || db.Ping() != nil {
		return &DCDB{}, err
	}
	log.Debug("return")
	return &DCDB{db, ConfigIni}, err
}

// GetFirstColumnName returns the name of the first column in the table
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

// GetAllTables returns the list of the tables
func (db *DCDB) GetAllTables() ([]string, error) {
	var result []string
	sql := "SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND    table_schema NOT IN ('pg_catalog', 'information_schema')"
	result, err := db.GetList(sql).String()
	if err != nil {
		return result, err
	}
	return result, nil
}

// GetFirstColumnNamesPg returns the first name of the column with PostgreSQL request
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

// Int64 converts all string values to int64
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

// String return the slice of strings
func (r *listResult) String() ([]string, error) {
	if r.err != nil {
		return r.result, r.err
	}
	return r.result, nil
}

//
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

// Single returns the single result of the query
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

// GetList returns the result of the query as listResult variable
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

// GetParent возвращает информацию откуда произошел вызов функции
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

// GetAll returns the result of the query as slice of map[string]string
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

// InsertInLogTx insert md5 hash and time into log_transaction
func (db *DCDB) InsertInLogTx(binaryTx []byte, time int64) error {
	txMD5 := Md5(binaryTx)
	err := db.ExecSQL("INSERT INTO log_transactions (hash, time) VALUES ([hex], ?)", txMD5, time)
	log.Debug("INSERT INTO log_transactions (hash, time) VALUES ([hex], %s)", txMD5)
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

// DelLogTx deletes a row with the specified md5 hash in log_transaction
func (db *DCDB) DelLogTx(binaryTx []byte) error {
	txMD5 := Md5(binaryTx)
	affected, err := db.ExecSQLGetAffect("DELETE FROM log_transactions WHERE hex(hash) = ?", txMD5)
	log.Debug("DELETE FROM log_transactions WHERE hex(hash) = %s / affected = %d", txMD5, affected)
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

// QueryRows returns the result of the query
func (db *DCDB) QueryRows(query string, args ...interface{}) (*sql.Rows, error) {
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	return db.Query(newQuery, newArgs...)
}

// ExecSQLGetLastInsertID insert a row and returns the last id
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
				log.Error("database is locked %s / %s / %s", newQuery, newArgs, GetParent())
				time.Sleep(250 * time.Millisecond)
				continue
			} else {
				return "", fmt.Errorf("%s in query %s %s %s", err, newQuery, newArgs, GetParent())
			}
		} else {
			switch v.(type) {
			case int:
				lastID = IntToStr(v.(int))
			case int64:
				lastID = Int64ToStr(v.(int64))
			case float64:
				lastID = Float64ToStr(v.(float64))
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

// FormatQueryArgs formats the query
func FormatQueryArgs(q, dbType string, args ...interface{}) (string, []interface{}) {
	var newArgs []interface{}

	newQ := q
	if ok, _ := regexp.MatchString(`CREATE TABLE`, newQ); !ok {
		newQ = strings.Replace(newQ, "[hex]", "decode(?,'HEX')", -1)
		newQ = strings.Replace(newQ, " authorization", ` "authorization"`, -1)
		newQ = strings.Replace(newQ, "user,", `"user",`, -1)
		newQ = ReplQ(newQ)
		newArgs = args
	}

	/*r, _ := regexp.Compile(`\s*([0-9]+_[\w]+)(?:\.|\s|\)|$)`)
	indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
	for i := len(indexArr) - 1; i >= 0; i-- {
		newQ = newQ[:indexArr[i][2]] + `"` + newQ[indexArr[i][2]:indexArr[i][3]] + `"` + newQ[indexArr[i][3]:]
	}*/

	r, _ := regexp.Compile(`hex\(([\w]+)\)`)
	indexArr := r.FindAllStringSubmatchIndex(newQ, -1)
	for i := len(indexArr) - 1; i >= 0; i-- {
		newQ = newQ[:indexArr[i][0]] + `LOWER(encode(` + newQ[indexArr[i][2]:indexArr[i][3]] + `, 'hex'))` + newQ[indexArr[i][1]:]
	}

	return newQ, newArgs
}

// CheckInstall waits for the end of the installation
func (db *DCDB) CheckInstall(DaemonCh chan bool, AnswerDaemonCh chan string, GoroutineName string) bool {
	// Возможна ситуация, когда инсталяция еще не завершена. База данных может быть создана, а таблицы еще не занесены
	// there could be the situation when installation is not over yet. Database could be created but tables are not inserted yet
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
			// the moment could happen when the database is closed and there is a download of the completed database from the server
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
	/*affect, err := res.RowsAffected()
	lastID, err := res.LastInsertId()
	if db.ConfigIni["sql_log"] == "1" {
		parent := GetParent()
		log.Debug("SQL: %v / RowsAffected=%d / LastInsertId=%d / %s / %s", newQuery, affect, lastID, newArgs, parent)
	}*/
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
	lastID, err := res.LastInsertId()
	if db.ConfigIni["sql_log"] == "1" {
		log.Debug("SQL: %s / RowsAffected=%d / LastInsertId=%d / %s", newQuery, affect, lastID, newArgs)
	}
	return affect, nil
}

/* для юнит-тестов. снимок всех данных в БД
// for unit tests. snapshot of all data in database
func (db *DCDB) HashTableData(table, where, orderBy string) (string, error) {
	/var columns string;
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
	}/
	if len(orderBy) > 0 {
		orderBy = " ORDER BY " + orderBy
	}

	// это у всех разное, а значит и хэши будут разные, а это будет вызывать путаницу
	// this is different for every one and it means hashes will be different and this will cause confusion
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
		err := db.ExecSQL("SET @@group_concat_max_len = 4294967295")
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

	/if strings.Count(table, "my_table")>0 {
		columns = strings.Replace(columns,",notification","",-1)
		columns = strings.Replace(columns,"notification,","",-1)
		q="SELECT md5(CAST((array_agg("+columns+" "+orderBy+")) AS text)) FROM \""+table+"\" "+where
	}
	if strings.Count(columns, "cron_checked_time")>0 {
		columns = strings.Replace(columns, ",cron_checked_time", "", -1)
		columns = strings.Replace(columns, "cron_checked_time,", "", -1)
		q="SELECT md5(CAST((array_agg("+columns+" "+orderBy+")) AS text)) FROM \""+table+"\" "+where
	}/

	if logOff {
		db.ConfigIni["sql_log"] = "1"
	}
	return hash, nil
}*/

// GetLastBlockData returns the data of the latest block
func (db *DCDB) GetLastBlockData() (map[string]int64, error) {
	result := make(map[string]int64)
	confirmedBlockID, err := db.GetConfirmedBlockID()
	if err != nil {
		return result, ErrInfo(err)
	}
	if confirmedBlockID == 0 {
		confirmedBlockID = 1
	}
	log.Debug("%v", "confirmedBlockId", confirmedBlockID)
	// получим время из последнего подвержденного блока
	// obtain the time of the last affected block
	lastBlockBin, err := db.Single("SELECT data FROM block_chain WHERE id = ?", confirmedBlockID).Bytes()
	if err != nil || len(lastBlockBin) == 0 {
		return result, ErrInfo(err)
	}
	// ID блока
	result["blockId"] = int64(BinToDec(lastBlockBin[1:5]))
	// Время последнего блока
	// the time of the last block
	result["lastBlockTime"] = int64(BinToDec(lastBlockBin[5:9]))
	return result, nil
}

// GetNodePrivateKey returns the private key from my_nodes_key
func (db *DCDB) GetNodePrivateKey() (string, error) {
	var key string
	key, err := db.Single("SELECT private_key FROM my_node_keys WHERE block_id = (SELECT max(block_id) FROM my_node_keys)").String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return key, nil
}

/*
func (db *DCDB) GetPrivateKey(myPrefix string) (string, error) {
	var key string
	key, err := db.Single("SELECT private_key FROM my_keys WHERE block_id = (SELECT max(block_id) FROM my_keys)").String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return key, nil
}*/

// GetNodeConfig returns config parameters
func (db *DCDB) GetNodeConfig() (map[string]string, error) {
	return db.OneRow("SELECT * FROM config").String()
}

// FormatQuery formats the query
func (db *DCDB) FormatQuery(q string) string {

	newQ := q
	if ok, _ := regexp.MatchString(`CREATE TABLE`, newQ); !ok {
		switch db.ConfigIni["db_type"] {
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

// GetConfirmedBlockID returns the maximal block id from confirmations
func (db *DCDB) GetConfirmedBlockID() (int64, error) {

	result, err := db.Single("SELECT max(block_id) FROM confirmations WHERE good >= ?", consts.MIN_CONFIRMED_NODES).Int64()
	if err != nil {
		return 0, err
	}
	//log.Debug("%v", "result int64",StrToInt64(result))
	return result, nil

}

// GetMyStateIDAndWalletID returns state id and wallet id from config
func (db *DCDB) GetMyStateIDAndWalletID() (int64, int64, error) {
	myStateID, err := db.GetMyStateID()
	if err != nil {
		return 0, 0, err
	}
	myWalletID, err := db.GetMyWalletID()
	if err != nil {
		return 0, 0, err
	}
	return myStateID, myWalletID, nil
}

// GetHosts returns the list of hosts
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

// CheckDelegateCB checks if the state is delegated
func (db *DCDB) CheckDelegateCB(myStateID int64) (bool, error) {
	delegate, err := db.OneRow("SELECT delegate_wallet_id, delegate_state_id FROM system_recognized_states WHERE state_id = ?", myStateID).Int64()
	if err != nil {
		return false, err
	}
	// Если мы - государство и у нас указан delegate, т.е. мы делегировали полномочия по поддержанию ноды другому юзеру или государству, то выходим.
	// If we are the state and we have the delegate specified (we delegated the authority to maintain the node to another user or state, then we leave).
	if delegate["delegate_wallet_id"] > 0 || delegate["delegate_state_id"] > 0 {
		return true, nil
	}
	return false, nil
}

// GetMyWalletID returns wallet id from config
func (db *DCDB) GetMyWalletID() (int64, error) {
	walletID, err := db.Single("SELECT dlt_wallet_id FROM config").Int64()
	if err != nil {
		return 0, err
	}
	if walletID == 0 {
		//		walletId, err = db.Single("SELECT wallet_id FROM dlt_wallets WHERE address = ?", *WalletAddress).Int64()
		walletID = lib.StringToAddress(*WalletAddress)
	}
	return walletID, nil
}

// GetMyStateID returns state id from config
func (db *DCDB) GetMyStateID() (int64, error) {
	return db.Single("SELECT state_id FROM config").Int64()
}

// GetBlockID returns teh latest block id from info_block
func (db *DCDB) GetBlockID() (int64, error) {
	return db.Single("SELECT block_id FROM info_block").Int64()
}

// GetWalletIDByPublicKey convert public key to wallet id
func (db *DCDB) GetWalletIDByPublicKey(publicKey []byte) (int64, error) {
	/*	log.Debug("string(HashSha1Hex(publicKey) %s", string(HashSha1Hex(publicKey)))
		log.Debug("publicKey %s", publicKey)
		log.Debug("key %s", key)
		log.Debug("b58 %s", b58.Encode(lib.Address(key)))
		walletId, err := db.Single(`SELECT wallet_id FROM dlt_wallets WHERE address = ?`,
			string(b58.Encode(lib.Address(key)))).Int64()
		if err != nil {
			return 0, ErrInfo(err)
		}
		return walletId, nil*/
	key, _ := hex.DecodeString(string(publicKey))
	return int64(lib.Address(key)), nil
}

/*func (db *DCDB) GetCitizenIdByPublicKey(publicKey []byte) (int64, error) {
	walletId, err := db.Single(`SELECT citizen_id FROM ea_citizens WHERE hex(public_key_0) = ?`, string(publicKey)).Int64()
	if err != nil {
		return 0, ErrInfo(err)
	}
	return walletId, nil
}*/

// GetInfoBlock return the information about the latest block
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

// GetNodePublicKey returns the node public key of the wallet id
func (db *DCDB) GetNodePublicKey(waletID int64) ([]byte, error) {
	result, err := db.Single("SELECT node_public_key FROM dlt_wallets WHERE wallet_id = ?", waletID).Bytes()
	if err != nil {
		return []byte(""), err
	}
	return result, nil
}

// GetNodePublicKeyWalletOrCB returns node public key of wallet id or state id
func (db *DCDB) GetNodePublicKeyWalletOrCB(walletID, stateID int64) ([]byte, error) {
	var result []byte
	var err error
	if walletID != 0 {
		log.Debug("wallet_id %v state_id %v", walletID, stateID)
		result, err = db.Single("SELECT node_public_key FROM dlt_wallets WHERE wallet_id = ?", walletID).Bytes()
		if err != nil {
			return []byte(""), err
		}
	} else {
		result, err = db.Single("SELECT node_public_key FROM system_recognized_states WHERE state_id = ?", stateID).Bytes()
		if err != nil {
			return []byte(""), err
		}
	}
	return result, nil
}

// GetPublicKeyWalletOrCitizen returns public key of the wallet id or citizen id
func (db *DCDB) GetPublicKeyWalletOrCitizen(walletID, citizenID int64) ([]byte, error) {
	var result []byte
	var err error
	if walletID != 0 {
		result, err = db.Single("SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?", walletID).Bytes()
		if err != nil {
			return []byte(""), err
		}
	} else {
		result, err = db.Single("SELECT public_key_0 FROM ea_citizens WHERE citizen_is = ?", citizenID).Bytes()
		if err != nil {
			return []byte(""), err
		}
	}
	return result, nil
}

// UpdMainLock updates the lock time
func (db *DCDB) UpdMainLock() error {
	return db.ExecSQL("UPDATE main_lock SET lock_time = ?", time.Now().Unix())
}

// CheckDaemonsRestart is reserved
func (db *DCDB) CheckDaemonsRestart() bool {
	return false
}

// DbLock locks deamons
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
			err = db.ExecSQL(`INSERT INTO main_lock(lock_time, script_name, info) VALUES(?, ?, ?)`, time.Now().Unix(), goRoutineName, Caller(2))
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
					db.ExecSQL(`DELETE FROM main_lock`)
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

// DeleteQueueBlock deletes a row from queue_blocks with the specified hash
func (db *DCDB) DeleteQueueBlock(hashHex string) error {
	return db.ExecSQL("DELETE FROM queue_blocks WHERE hex(hash) = ?", hashHex)
}

// SetAI sets serial sequence for the table
func (db *DCDB) SetAI(table string, AI int64) error {

	AiID, err := db.GetAiID(table)
	if err != nil {
		return ErrInfo(err)
	}

	if db.ConfigIni["db_type"] == "postgresql" {
		pgGetSerialSequence, err := db.Single("SELECT pg_get_serial_sequence('" + table + "', '" + AiID + "')").String()
		if err != nil {
			return ErrInfo(err)
		}
		err = db.ExecSQL("ALTER SEQUENCE " + pgGetSerialSequence + " RESTART WITH " + Int64ToStr(AI))
		if err != nil {
			return ErrInfo(err)
		}
	}
	return nil
}

// PrintSleep writes the error to log and make a pause
func (db *DCDB) PrintSleep(v interface{}, sleep time.Duration) {
	var err error
	switch v.(type) {
	case string:
		err = errors.New(v.(string))
	case error:
		err = v.(error)
	}
	log.Error("%v (%v)", err, GetParent())
	Sleep(sleep)
}

// DbUnlock unlocks database
func (db *DCDB) DbUnlock(goRoutineName string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()
	log.Debug("DbUnlock %v %v", Caller(2), goRoutineName)
	affect, err := db.ExecSQLGetAffect("DELETE FROM main_lock WHERE script_name = ?", goRoutineName)
	log.Debug("main_lock affect: %d, goRoutineName: %s", affect, goRoutineName)
	if err != nil {
		log.Error("%s", ErrInfo(err))
		return ErrInfo(err)
	}
	return nil
}

// UpdDaemonTime is reserved
func (db *DCDB) UpdDaemonTime(name string) {

}

// GetAiID returns auto increment column
func (db *DCDB) GetAiID(table string) (string, error) {
	exists := ""
	column := "id"
	if table == "users" {
		column = "user_id"
	} else if table == "miners" {
		column = "miner_id"
	} else {
		switch db.ConfigIni["db_type"] {
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
		}
	}
	return column, nil
}

// NodesBan is reserved
func (db *DCDB) NodesBan(info string) error {

	return nil
}

// GetBlockDataFromBlockChain returns the block information from the blockchain
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

// GetTxTypeAndUserId returns tx type, wallet and citizen id from the block data
func GetTxTypeAndUserId(binaryBlock []byte) (txType int64, walletID int64, citizenID int64) {
	tmp := binaryBlock[:]
	txType = BinToDecBytesShift(&binaryBlock, 1)
	if consts.IsStruct(int(txType)) {
		var txHead consts.TxHeader
		lib.BinUnmarshal(&tmp, &txHead)
		walletID = txHead.WalletID
		citizenID = txHead.CitizenID
	} else if txType > 127 {
		header := consts.TXHeader{}
		err := lib.BinUnmarshal(&tmp, &header)
		if err == nil {
			if header.StateID > 0 {
				citizenID = int64(header.WalletID)
			} else {
				walletID = int64(header.WalletID)
			}
		}
	} else {
		BytesShift(&binaryBlock, 4) // уберем время
		walletID = BytesToInt64(BytesShift(&binaryBlock, DecodeLength(&binaryBlock)))
		citizenID = BytesToInt64(BytesShift(&binaryBlock, DecodeLength(&binaryBlock)))
	}
	return
}

// DecryptData decrypts tx data
func (db *DCDB) DecryptData(binaryTx *[]byte) ([]byte, []byte, []byte, error) {

	if len(*binaryTx) == 0 {
		return nil, nil, nil, ErrInfo("len(binaryTx) == 0")
	}

	// вначале пишется user_id, чтобы в режиме пула можно было понять, кому шлется и чей ключ использовать
	// at the beginning the user ID is written to know in the pool mode to whom it is sent and what key to use
	myUserId := BinToDecBytesShift(&*binaryTx, 5)
	log.Debug("myUserId: %d", myUserId)

	// изымем зашифрванный ключ, а всё, что останется в $binary_tx - сами зашифрованные хэши тр-ий/блоков
	// remove the encrypted key, and all that stay in $binary_tx will be encrypted keys of the transactions/blocks
	encryptedKey := BytesShift(&*binaryTx, DecodeLength(&*binaryTx))
	log.Debug("encryptedKey: %x", encryptedKey)
	log.Debug("encryptedKey: %s", encryptedKey)

	// далее идет 16 байт IV
	// 16 bytes IV go further
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

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, nil, ErrInfo(err)
	}

	decKey, err := rsa.DecryptPKCS1v15(crand.Reader, privateKey, encryptedKey)
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

func (db *DCDB) FindInFullNodes(myStateID, myWalletId int64) (int64, error) {
	return db.Single("SELECT id FROM full_nodes WHERE final_delegate_state_id = ? OR final_delegate_wallet_id = ? OR state_id = ? OR wallet_id = ?", myStateID, myWalletId, myStateID, myWalletId).Int64()
}

func (db *DCDB) GetBinSign(forSign string) ([]byte, error) {
	nodePrivateKey, err := db.GetNodePrivateKey()
	if err != nil {
		return nil, ErrInfo(err)
	}
	/*	log.Debug("nodePrivateKey = %s", nodePrivateKey)
				// подписываем нашим нод-ключем данные транзакции
		// sign the data of transaction by our node-key
				privateKey, err := MakePrivateKey(nodePrivateKey)
				if err != nil {
					return nil, ErrInfo(err)
				}
				return rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA1, HashSha1(forSign))*/
	return lib.SignECDSA(nodePrivateKey, forSign)
}

func (db *DCDB) InsertReplaceTxInQueue(data []byte) error {

	log.Debug("DELETE FROM queue_tx WHERE hex(hash) = %s", Md5(data))
	err := db.ExecSQL("DELETE FROM queue_tx WHERE hex(hash) = ?", Md5(data))
	if err != nil {
		return ErrInfo(err)
	}
	log.Debug("INSERT INTO queue_tx (hash, data) VALUES (%s, %s)", Md5(data), BinToHex(data))
	err = db.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", Md5(data), BinToHex(data))
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

func (db *DCDB) GetSleepTime(myWalletId, myStateID, prevBlockStateID, prevBlockWalletId int64) (int64, error) {
	// возьмем список всех full_nodes
	// take the list of all full_nodes
	fullNodesList, err := db.GetAll("SELECT id, wallet_id, state_id as state_id FROM full_nodes", -1)
	if err != nil {
		return int64(0), ErrInfo(err)
	}
	log.Debug("fullNodesList %s", fullNodesList)

	// определим full_node_id того, кто должен был генерить блок (но мог это делегировать)
	// determine full_node_id of the one, who had to generate a block (but could delegate this)
	prevBlockFullNodeId, err := db.Single("SELECT id FROM full_nodes WHERE state_id = ? OR wallet_id = ?", prevBlockStateID, prevBlockWalletId).Int64()
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
	myPosition := func(fullNodesList []map[string]string, myWalletId, myStateID int64) int {
		log.Debug("%v %v", fullNodesList, myWalletId)
		for i, full_nodes := range fullNodesList {
			if StrToInt64(full_nodes["state_id"]) == myStateID || StrToInt64(full_nodes["wallet_id"]) == myWalletId || StrToInt64(full_nodes["final_delegate_state_id"]) == myWalletId || StrToInt64(full_nodes["final_delegate_wallet_id"]) == myWalletId {
				return i
			}
		}
		return -1
	}(fullNodesList, myWalletId, myStateID)
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
	var err error
	stateId_, err := db.Single(`SELECT id FROM system_states WHERE id = ?`, stateId).String()
	if err != nil {
		return ``, err
	}
	stateName := ""
	if stateId_ != "0" {
		stateName, err = db.Single(`SELECT value FROM "` + stateId_ + `_state_parameters" WHERE name = 'state_name'`).String()
		if err != nil {
			return ``, err
		}
	}
	return stateName, nil
}

func (db *DCDB) CheckStateName(stateId int64) (bool, error) {
	stateId, err := db.Single(`SELECT id FROM system_states WHERE id = ?`, stateId).Int64()
	if err != nil {
		return false, err
	}
	if stateId > 0 {
		return true, nil
	}
	return false, fmt.Errorf("null stateId")
}

func (db *DCDB) GetFuel() decimal.Decimal {
	// fuel = qEGS/F
	/*	fuelMutex.Lock()
		defer fuelMutex.Unlock()
		if cacheFuel <= 0 {*/
	fuel, _ := db.Single(`SELECT value FROM system_parameters WHERE name = ?`, "fuel_rate").String()
	//}
	cacheFuel, _ := decimal.NewFromString(fuel)
	return cacheFuel
}

func (db *DCDB) UpdateFuel() {
	/*	fuelMutex.Lock()
		cacheFuel, _ = db.Single(`SELECT value FROM system_parameters WHERE name = ?`, "fuel_rate").Int64()
		fuelMutex.Unlock()*/
}

func (db *DCDB) IsIndex(tblname, column string) (bool, error) {
	/* Short version if index name = tablename_columnname
		indexes, err := db.GetAll(`SELECT 1 FROM  pg_class c JOIN  pg_namespace n ON n.oid = c.relnamespace
	                 WHERE  c.relname = ?  AND  n.nspname = 'public'`, 1, tblname + `_` + column)*/
	//	Full version
	indexes, err := db.GetAll(`select t.relname as table_name, i.relname as index_name, a.attname as column_name 
	 from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
         and t.relkind = 'r'  and t.relname = ?  and a.attname = ?`, 1, tblname, column)
	if err != nil {
		return false, err
	}
	return len(indexes) > 0, nil
}

func (db *DCDB) NumIndexes(tblname string) (int, error) {
	indexes, err := db.Single(`select count( i.relname) from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
         and t.relkind = 'r'  and t.relname = ?`, tblname).Int64()
	if err != nil {
		return 0, err
	}
	return int(indexes - 1), nil
}

func (db *DCDB) IsCustomTable(table string) (isCustom bool, err error) {
	if (table[0] >= '0' && table[0] <= '9') || strings.HasPrefix(table, `global_`) {
		if off := strings.IndexByte(table, '_'); off > 0 {
			prefix := table[:off]
			if name, err := db.Single(`select name from "`+prefix+`_tables" where name = ?`, table).String(); err == nil {
				isCustom = name == table
			}
		}
	}
	return
}

func GetColumnType(tblname, column string) (itype string) {
	coltype, _ := DB.OneRow(`select data_type,character_maximum_length from information_schema.columns
where table_name = ? and column_name = ?`, tblname, column).String()
	if len(coltype) > 0 {
		switch {
		case coltype[`data_type`] == "character varying":
			itype = `text`
		case coltype[`data_type`] == "bytea":
			itype = "varchar"
		case coltype[`data_type`] == `bigint`:
			itype = "numbers"
		case strings.HasPrefix(coltype[`data_type`], `timestamp`):
			itype = "date_time"
		case strings.HasPrefix(coltype[`data_type`], `numeric`):
			itype = "money"
		case strings.HasPrefix(coltype[`data_type`], `double`):
			itype = "double"
		}
	}
	return
}

func (db *DCDB) IsState(country string) (int64, error) {
	data, err := db.GetList(`SELECT id FROM system_states`).Int64()
	if err != nil {
		return 0, err
	}
	for _, id := range data {
		state_name, err := db.Single(fmt.Sprintf(`SELECT value FROM "%d_state_parameters" WHERE name = 'state_name'`, id)).String()
		if err != nil {
			return 0, err
		}
		if strings.ToLower(state_name) == strings.ToLower(country) {
			return id, nil
		}
	}
	return 0, nil
}

func (db *DCDB) IsNodeState(state int64, host string) bool {
	if strings.HasPrefix(host, `localhost`) {
		return true
	}
	if val, ok := db.ConfigIni[`node_state_id`]; ok {
		if val == `*` {
			return true
		}
		for _, id := range strings.Split(val, `,`) {
			if StrToInt64(id) == state {
				return true
			}
		}
	}
	return false
}

func (db *DCDB) IsTable(tblname string) bool {
	name, _ := db.Single(`SELECT table_name FROM information_schema.tables 
         WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')
     	AND table_name=?`, tblname).String()
	return name == tblname
}

func (db *DCDB) SendTx(txType int64, adminWallet int64, data []byte) (err error) {
	md5 := Md5(data)
	err = db.ExecSQL(`INSERT INTO transactions_status (
			hash, time,	type, wallet_id, citizen_id	) VALUES (
			[hex], ?, ?, ?, ? )`, md5, time.Now().Unix(), txType, adminWallet, adminWallet)
	if err != nil {
		return err
	}
	err = db.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, hex.EncodeToString(data))
	if err != nil {
		return err
	}
	return
}
