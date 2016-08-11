package utils

import (
	"crypto"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/op/go-logging"
	"math"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
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
		db, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", ConfigIni["db_user"], ConfigIni["db_password"], ConfigIni["db_name"]))
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

func (db *DCDB) GetCfUrl() (string, error) {
	return db.Single("SELECT cf_url FROM config").String()
}
func (db *DCDB) GetMainLockName() (string, error) {
	return db.Single("SELECT script_name FROM main_lock").String()
}

/*func (db *DCDB) SendMail(message, subject, To string, mailData map[string]string, community bool, poolAdminUserId int64) error {

	if len(mailData["use_smtp"]) > 0 && len(mailData["smtp_server"]) > 0 {
		err := sendMail(message, subject, To, mailData)
		if err != nil {
			return ErrInfo(err)
		}*/
		/*} else if community {
		// в пуле пробуем послать с смтп-ешника админа пула
		prefix := Int64ToStr(poolAdminUserId) + "_"
		mailData, err := db.OneRow("SELECT * FROM " + prefix + "my_table").String()
		if err != nil {
			return ErrInfo(err)
		}
		err = sendMail(message, subject, To, mailData)
		if err != nil {
			return ErrInfo(err)
		}*/
/*	} else {
		return errors.New(`Incorrect mail data`)
	}
	return nil
}*/

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

type Variables struct {
	Int64   map[string]int64
	String  map[string]string
	Float64 map[string]float64
}

func (db *DCDB) GetAllVariables() (*Variables, error) {
	result := new(Variables)
	result.Int64 = make(map[string]int64)
	result.String = make(map[string]string)
	result.Float64 = make(map[string]float64)
	all, err := db.GetAll("SELECT * FROM variables", -1)
	//fmt.Println(all)
	if err != nil {
		return result, err
	}
	for _, v := range all {
		switch v["name"] {
		case "max_pool_users", "alert_error_time", "error_time", "promised_amount_points", "promised_amount_votes_0", "promised_amount_votes_1", "promised_amount_votes_period", "holidays_max", "limit_abuses", "limit_abuses_period", "limit_promised_amount", "limit_promised_amount_period", "limit_cash_requests_out", "limit_cash_requests_out_period", "limit_change_geolocation", "limit_change_geolocation_period", "limit_holidays", "limit_holidays_period", "limit_message_to_admin", "limit_message_to_admin_period", "limit_mining", "limit_mining_period", "limit_node_key", "limit_node_key_period", "limit_primary_key", "limit_primary_key_period", "limit_votes_miners", "limit_votes_miners_period", "limit_votes_complex", "limit_votes_complex_period", "limit_commission", "limit_commission_period", "limit_new_miner", "limit_new_miner_period", "limit_new_user", "limit_new_user_period", "max_block_size", "max_block_user_transactions", "max_day_votes", "max_tx_count", "max_tx_size", "max_user_transactions", "miners_keepers", "miner_points", "miner_votes_0", "miner_votes_1", "miner_votes_attempt", "miner_votes_period", "mining_votes_0", "mining_votes_1", "mining_votes_period", "min_miners_keepers", "node_voting", "node_voting_period", "rollback_blocks_1", "rollback_blocks_2", "limit_change_host", "limit_change_host_period", "min_miners_of_voting", "min_hold_time_promise_amount", "min_promised_amount", "points_update_time", "reduction_period", "new_pct_period", "new_max_promised_amount", "new_max_other_currencies", "cash_request_time", "limit_for_repaid_fix", "limit_for_repaid_fix_period", "miner_newbie_time", "system_commission":
			result.Int64[v["name"]] = StrToInt64(v["value"])
		case "points_factor":
			result.Float64[v["name"]] = StrToFloat64(v["value"])
		case "sleep":
			result.String[v["name"]] = v["value"]
		}
	}
	return result, err
}

/*
func (db *DCDB) SingleInt64(query string, args ...interface{}) (int64, error) {
	result, err := db.Single(query, args...)
	if err != nil {
		return 0, err
	}
	return StrToInt64(result), nil
}
*/

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

func (db *DCDB) GetCfAuthorInfo(userId, levelUp string) (map[string]string, error) {
	data, err := db.OneRow("SELECT name, avatar FROM users WHERE user_id  =  ?", userId).String()
	if err != nil {
		return nil, ErrInfo(err)
	}
	if len(data["avatar"]) == 0 {
		data["avatar"] = levelUp + "static/img/noavatar.png"
	}
	if len(data["name"]) == 0 {
		data["name"] = "Noname"
	}

	// сколько проектов создал
	created, err := db.Single("SELECT count(id) FROM cf_projects WHERE user_id  =  ?", userId).Int64()
	if err != nil {
		return nil, ErrInfo(err)
	}
	data["created"] = Int64ToStr(created)

	// сколько проектов профинансировал
	backed, err := db.Single("SELECT count(project_id) FROM cf_funding WHERE user_id  =  ? GROUP BY project_id", userId).Int64()
	if err != nil {
		return nil, ErrInfo(err)
	}
	data["backed"] = Int64ToStr(backed)

	return data, nil
}

func (db *DCDB) GetAllCfLng() (map[string]string, error) {
	return db.GetMap(`SELECT id, name FROM cf_lang ORDER BY name`, "id", "name")
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

func (db *DCDB) GetCountMiners() (int64, error) {
	return db.Single("SELECT count(miner_id) FROM miners WHERE active = 1").Int64()
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

func (db *DCDB) CountMinerAttempt(userId int64, vType string) (int64, error) {
	count, err := db.Single("SELECT count(user_id) FROM votes_miners WHERE user_id = ? AND type = ?", userId, vType).Int64()
	if err != nil {
		return 0, ErrInfo(err)
	}
	return count, nil
}

func (db *DCDB) GetCfProjectData(id, endTime, langId int64, amount float64, levelUp string) (map[string]string, error) {
	var err error
	result := make(map[string]string)

	// и картинка для обложки
	data, err := db.OneRow("SELECT blurb_img, lang_id FROM cf_projects_data WHERE project_id  =  ? AND lang_id  =  ? ORDER BY id ASC", id, langId).String()
	if err != nil {
		return result, ErrInfo(err)
	}
	result["blurb_img"] = data["blurb_img"]
	result["lang_id"] = data["lang_id"]
	if len(result["blurb_img"]) == 0 {
		result["blurb_img"] = levelUp + "img/cf_blurb_img.png"
	}
	// сколько собрано
	funding_amount, err := db.Single("SELECT sum(amount) FROM cf_funding WHERE project_id  =  ? AND del_block_id  =  0", id).Float64()
	if err != nil {
		return result, ErrInfo(err)
	}
	result["funding_amount"] = Float64ToStrPct(funding_amount)
	// % собрано
	log.Debug("%v", "funding_amount", funding_amount)
	log.Debug("%v", "amount", amount)
	if amount > 0 {
		result["pct"] = Float64ToStrPct(Round((funding_amount / amount * 100), 0))
	} else {
		result["pct"] = "0"
	}
	result["funding_amount"] = Float64ToStrPct(Round(funding_amount, 1))

	// дней до окончания
	days_ := int64(Round(float64(endTime-time.Now().Unix())/86400, 0))
	if days_ < 0 {
		result["days"] = "0"
	} else {
		result["days"] = Int64ToStr(days_)
	}
	return result, nil
}

func (db *DCDB) NodeAdminAccess(sessUserId, sessRestricted int64) (bool, error) {
	if sessRestricted != 0 || sessUserId <= 0 {
		log.Debug("%v", "NodeAdminAccess1")
		return false, nil
	}
	community, err := db.GetCommunityUsers()
	if err != nil {
		log.Debug("%v", "NodeAdminAccess2")
		return false, ErrInfo(err)
	}
	if len(community) > 0 {
		pool_admin_user_id, err := db.GetPoolAdminUserId()
		if err != nil {
			log.Debug("%v", "NodeAdminAccess3")
			return false, ErrInfo(err)
		}
		if sessUserId == pool_admin_user_id {
			return true, nil
		} else {
			log.Debug("%v", "NodeAdminAccess4")
			return false, nil
		}
	} else {
		log.Debug("%v", "NodeAdminAccess0")
		return true, nil
	}
}

func (db *DCDB) ExecSqlGetLastInsertId(query, returning string, args ...interface{}) (int64, error) {
	var lastId int64
	var res sql.Result
	var err error
	newQuery, newArgs := FormatQueryArgs(query, db.ConfigIni["db_type"], args...)
	if db.ConfigIni["db_type"] == "postgresql" {
		newQuery = newQuery + " RETURNING " + returning
		for {
			err := db.QueryRow(newQuery, newArgs...).Scan(&lastId)
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

		if db.ConfigIni["sql_log"] == "1" {
			log.Debug("SQL: %s / LastInsertId=%d / %s", newQuery, lastId, newArgs)
		}
		/*r, _ := regexp.Compile(`(?i)insert into (\w+)`)
		find := r.FindStringSubmatch(newQuery)
		err =  db.ExecSql("SELECT setval('"+find[1]+"_"+returning+"_seq', max("+returning+")) FROM   "+find[1]+";")
		if err != nil {
			return 0, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
		}*/

	} else {
		/*res, err := db.Exec(newQuery, newArgs...)
		if err != nil {
			return 0, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
		}*/
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
		if err != nil {
			return 0, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
		}
		lastId, err = res.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("%s in query %s %s", err, newQuery, newArgs)
		}
		if db.ConfigIni["sql_log"] == "1" {
			log.Debug("SQL: %s / RowsAffected=%d / LastInsertId=%d / %s", newQuery, affect, lastId, newArgs)
		}
	}
	return lastId, nil
}

type exampleSpots struct {
	Face    map[string][]interface{} `json:"face"`
	Profile map[string][]interface{} `json:"profile"`
}

func (db *DCDB) GetPoints(lng map[string]string) (map[string]string, error) {

	result := make(map[string]string)
	result["face"] = ""
	result["profile"] = ""

	exampleSpots_, err := db.Single("SELECT example_spots FROM spots_compatibility").String()
	if err != nil {
		return nil, ErrInfo(err)
	}
	exampleSpots := make(map[string]map[string][]interface{})
	err = json.Unmarshal([]byte(exampleSpots_), &exampleSpots)
	if err != nil {
		return nil, ErrInfo(err)
	}
	for pType, data := range exampleSpots {
		for i := 1; i <= len(data); i++ {
			arr := data[IntToStr(i)]
			id := IntToStr(i)
			result[pType] += fmt.Sprintf(`[%v, %v, '%v. %s'`, arr[0], arr[1], id, lng["points-"+pType+"-"+id])
			switch arr[2].(type) {
			case []interface{}:
				result[pType] += fmt.Sprintf(`, [%v, %v,`, StrToInt(arr[3].(string))-1, StrToInt(arr[4].(string))-1)
				for j := 0; j < len(arr[2].([]interface{})); j++ {
					result[pType] += fmt.Sprintf(`'%v'`, arr[2].([]interface{})[j])
					if j != len(arr[2].([]interface{}))-1 {
						result[pType] += ","
					}
				}
				result[pType] += "] ]"
			case string:
				result[pType] += "]"
			}
			result[pType] += ",\n"
		}
		result[pType] = result[pType][0 : len(result[pType])-2]
	}
	return result, nil
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

func (db *DCDB) GetMyNoticeData(sessCitizenId int64, sessWalletId int64, lang map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	result["account_status_text"] = lang["status_user"]

	// Инфа о последнем блоке
	blockData, err := db.GetLastBlockData()
	if err != nil {
		return result, ErrInfo(err)
	}
	result["cur_block_id"] = Int64ToStr(blockData["blockId"])
	t := time.Unix(blockData["lastBlockTime"], 0)
	result["time_last_block"] = t.Format("2006-01-02 15:04:05")
	result["time_last_block_int"] = Int64ToStr(blockData["lastBlockTime"])

	result["connections"], err = db.Single("SELECT count(*) FROM nodes_connection").String()
	if err != nil {
		return result, ErrInfo(err)
	}

	if time.Now().Unix()-blockData["lastBlockTime"] > 1800 {
		result["main_status"] = lang["downloading_blocks"]
		result["main_status_complete"] = "0"
	} else {
		result["main_status"] = lang["downloading_complete"]
		result["main_status_complete"] = "1"
	}

	return result, nil
}

func (db *DCDB) GetPoolAdminUserId() (int64, error) {
	result, err := db.Single("SELECT pool_admin_user_id FROM config").Int64()
	if err != nil {
		return 0, ErrInfo(err)
	}
	return result, nil
}

func (db *DCDB) CalcProfitGen(currencyId int64, amount float64, userId int64, lastUpdate, endTime int64, calcType string) (float64, error) {
	var err error
	pct, err := db.GetPct()
	if err != nil {
		return 0, err
	}
	var pointsStatus []map[int64]string
	var userHolidays [][]int64
	var maxPromisedAmounts map[int64][]map[int64]string
	var repaidAmount float64
	if calcType == "wallet" {
		pointsStatus = []map[int64]string{{0: "user"}}
	} else if calcType == "mining" { // обычная обещанная сумма
		pointsStatus, err = db.GetPointsStatus(userId, 0, nil)
		if err != nil {
			return 0, err
		}
		userHolidays, err = db.GetHolidays(userId)
		if err != nil {
			return 0, err
		}
		maxPromisedAmounts, err = db.GetMaxPromisedAmounts()
		if err != nil {
			return 0, err
		}
		repaidAmount, err = db.GetRepaidAmount(userId, currencyId)
		if err != nil {
			return 0, err
		}
	} else if calcType == "repaid" { // погашенная обещанная сумма
		pointsStatus, err = db.GetPointsStatus(userId, 0, nil)
		if err != nil {
			return 0, err
		}
	}
	var profit float64
	if (calcType == "mining" || calcType == "repaid" && db.CheckCashRequests(userId) == nil) || calcType == "wallet" {
		log.Debug("currencyId", currencyId, "amount", amount, "lastUpdate", lastUpdate, "endTime", endTime, "pct[currencyId]", pct[currencyId], "pointsStatus",pointsStatus, "userHolidays", userHolidays, "maxPromisedAmounts[currencyId]", maxPromisedAmounts[currencyId], "repaidAmount", repaidAmount)
		profit, err = CalcProfit(amount, lastUpdate, endTime, pct[currencyId], pointsStatus, userHolidays, maxPromisedAmounts[currencyId], currencyId, repaidAmount)
		if err != nil {
			return 0, err
		}
	}
	return profit, nil
}

func (db *DCDB) GetCurrencyListFullName() (map[int64]string, error) {
	var result_ map[string]string
	result := make(map[int64]string)
	result_, err := db.GetMap("SELECT id, full_name FROM currency ORDER BY full_name", "id", "full_name")
	if err != nil {
		return result, err
	}
	for k, v := range result_ {
		result[StrToInt64(k)] = v
	}
	return result, nil
}

func (db *DCDB) GetCurrencyList(cf bool) (map[int64]string, error) {

	var result_ map[string]string
	result := make(map[int64]string)
	result_, err := db.GetMap("SELECT id, name FROM currency ORDER BY name", "id", "name")
	if err != nil {
		return result, err
	}

	if cf {
		result0, err := db.GetMap("SELECT id, name FROM cf_currency ORDER BY name", "id", "name")
		if err != nil {
			return result, err
		}
		for id, name := range result0 {
			result_[id] = name
		}
	}
	for k, v := range result_ {
		result[StrToInt64(k)] = v
	}
	return result, nil
}

func (db *DCDB) SendTxChangePkey(userId int64) error {
	txTime := Time()
	myPrefix := ""
	community, err := db.GetCommunityUsers()
	if len(community) > 0 {
		myPrefix = Int64ToStr(userId) + "_"
	}
	PendingPublicKey, err := db.Single(`SELECT public_key FROM ` + myPrefix + `my_keys WHERE status='my_pending'`).String()
	bin_public_key_1 := PendingPublicKey
	binPublicKeyPack := EncodeLengthPlusData(bin_public_key_1)
	// генерируем тр-ию и шлем в DC-сеть
	forSign := fmt.Sprintf("%d,%d,%d,%s,%s,%s", TypeInt("ChangePrimaryKey"), txTime, userId, BinToHex(PendingPublicKey), "", "")
	currentPrivateKey, err := db.GetMyPrivateKey(myPrefix)
	if err != nil {
		return ErrInfo(err)
	}
	privateKey, err := MakePrivateKey(currentPrivateKey)
	if err != nil {
		return ErrInfo(err)
	}
	signature1, err := rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA1, HashSha1(forSign))
	if err != nil {
		return ErrInfo(err)
	}
	log.Debug("bin signature1: %s", signature1)
	sign := EncodeLengthPlusData(([]byte(signature1)))
	binSignatures := EncodeLengthPlusData([]byte(sign))

	data := DecToBin(TypeInt("ChangePrimaryKey"), 1)
	data = append(data, DecToBin(Time(), 4)...)
	data = append(data, EncodeLengthPlusData(Int64ToByte(userId))...)
	data = append(data, EncodeLengthPlusData(binPublicKeyPack)...)
	data = append(data, binSignatures...)
	err = db.InsertReplaceTxInQueue(data)
	if err != nil {
		return ErrInfo(err)
	}
	md5 := Md5(data)
	err = db.ExecSql(`INSERT INTO transactions_status (
				hash,
				time,
				type,
				user_id
			)
			VALUES (
				[hex],
				?,
				?,
				?
			)`, md5, time.Now().Unix(), TypeInt("ChangePrimaryKey"), userId)
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

// последние тр-ии от данного юзера
func (db *DCDB) GetLastTx(userId int64, types []int64, limit int64, timeFormat string) ([]map[string]string, error) {
	var result []map[string]string
	var sqltypes string
	if types != nil {
		sqltypes = ` AND transactions_status.type IN (`+strings.Join(SliceInt64ToString(types), ",")+`)`
	}
	rows, err := db.Query(db.FormatQuery(`
			SELECT  transactions_status.hash,
						 transactions_status.time,
						 transactions_status.type,
						 transactions_status.user_id,
						 transactions_status.block_id,
						 transactions_status.error as txerror,
						 queue_tx.hash as queue_tx,
						 transactions.hash as tx
			FROM transactions_status
			LEFT JOIN transactions ON transactions.hash = transactions_status.hash
			LEFT JOIN queue_tx ON queue_tx.hash = transactions_status.hash
			WHERE  transactions_status.user_id = ? ` + sqltypes + `
			ORDER BY time DESC
			LIMIT ?
			`), userId, limit)
	if err != nil {
		return result, ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var hash, txTime, txType, user_id, block_id, txerror, queue_tx, tx []byte
		err = rows.Scan(&hash, &txTime, &txType, &user_id, &block_id, &txerror, &queue_tx, &tx)
		if err != nil {
			return result, ErrInfo(err)
		}
		if len(tx) > 0 || len(queue_tx) > 0 {
			txerror = []byte("")
		}
		timeInt := StrToInt64(string(txTime))
		t := time.Unix(timeInt, 0)
		txTimeFormat := []byte(t.Format(timeFormat))
		result = append(result, map[string]string{"hash": string(hash), "time": string(txTimeFormat), "time_int": string(txTime), "type": string(txType), "user_id": string(user_id), "block_id": string(block_id), "error": string(txerror), "queue_tx": string(queue_tx), "tx": string(tx)})
	}
	return result, nil
}

func (db *DCDB) GetBalances(userId int64) ([]DCAmounts, error) {
	var result []DCAmounts
	rows, err := db.Query(db.FormatQuery("SELECT amount, currency_id, last_update FROM dlt_wallets WHERE user_id= ?"), userId)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var amount float64
		var currency_id, last_update int64
		err = rows.Scan(&amount, &currency_id, &last_update)
		if err != nil {
			return result, err
		}
		profit, err := db.CalcProfitGen(currency_id, amount, userId, last_update, time.Now().Unix(), "wallet")
		if err != nil {
			return result, err
		}
		amount += profit
		amount = Round(amount, 8)
		forexOrdersAmount, err := db.Single("SELECT sum(amount) FROM forex_orders WHERE user_id  =  ? AND sell_currency_id  =  ? AND del_block_id  =  0", userId, currency_id).Float64()
		if err != nil {
			return result, err
		}
		amount -= forexOrdersAmount
		pctSec, err := db.Single("SELECT user FROM pct WHERE currency_id  =  ? ORDER BY block_id DESC", currency_id).Float64()
		if err != nil {
			return result, err
		}
		pct := Round((math.Pow(1+pctSec, 3600*24*365)-1)*100, 2)
		result = append(result, DCAmounts{CurrencyId: (currency_id), Amount: amount, Pct: pct, PctSec: pctSec})
	}
	return result, err
}

func (db *DCDB) GetPointsStatus(userId, pointsUpdateTime int64, BlockData *BlockData) ([]map[int64]string, error) {

	// т.к. перед вызовом этой функции всегда идет обновление points_status, значит при данном запросе у нас
	// всегда будут свежие данные, т.е. крайний элемент массива всегда будет относиться к текущим 30-и дням
	var result []map[int64]string
	rows, err := db.Query(db.FormatQuery("SELECT time_start, status FROM points_status WHERE user_id= ? ORDER BY time_start ASC"), userId)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var time_start int64
		var status string
		err = rows.Scan(&time_start, &status)
		if err != nil {
			return result, err
		}
		result = append(result, map[int64]string{time_start: status})
	}

	// НО! При фронтальной проверке может получиться, что последний элемент miner и прошло более 30-и дней.
	// поэтому нужно добавлять последний элемент = user, если вызов происходит не в блоке
	if BlockData == nil && len(result) > 0 {
		for time_start, _ := range result[len(result)-1] {
			if time_start < time.Now().Unix()-pointsUpdateTime {
				result = append(result, map[int64]string{time_start + pointsUpdateTime: "user"})
			}
		}
	}
	// для майнеров, которые не получили ни одного балла, а уже шлют кому-то DC, или для всех юзеров
	if len(result) == 0 {
		result = append(result, map[int64]string{0: "user"})
	}
	return result, nil
}

func (db *DCDB) GetMyPublicKey(myPrefix string) ([]byte, error) {
	result, err := db.Single("SELECT public_key FROM " + myPrefix + "my_keys WHERE block_id = (SELECT max(block_id) FROM " + myPrefix + "my_keys)").Bytes()
	if err != nil {
		return []byte(""), ErrInfo(err)
	}
	return result, nil
}

func (db *DCDB) GetDataAuthorization(hash []byte) (string, error) {
	// получим данные для подписи
	log.Debug("hash %s", hash)
	data, err := db.Single(`SELECT data FROM authorization WHERE hex(hash) = ?`, hash).String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return data, nil
}

func (db *DCDB) GetAdminUserId() (int64, error) {
	result, err := db.Single("SELECT user_id FROM admin").Int64()
	if err != nil {
		return 0, ErrInfo(err)
	}
	return result, nil
}

func (db *DCDB) GetUserPublicKey(userId int64) (string, error) {
	result, err := db.Single("SELECT public_key_0 FROM users WHERE user_id = ?", userId).String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return result, nil
}

func (db *DCDB) GetMyPrivateKey(myPrefix string) (string, error) {
	key, err := db.Single("SELECT private_key FROM " + myPrefix + "my_keys WHERE block_id = (SELECT max(block_id) FROM " + myPrefix + "my_keys)").String()
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
	key, err := db.Single("SELECT public_key FROM " + myPrefix + "my_node_keys WHERE block_id = (SELECT max(block_id) FROM " + myPrefix + "my_node_keys)").String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return key, nil
}

func (db *DCDB) GetMaxPromisedAmount(currencyId int64) (float64, error) {
	result, err := db.Single("SELECT amount FROM max_promised_amounts WHERE currency_id = ? ORDER BY time DESC", currencyId).Float64()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (db *DCDB) GetMaxPromisedAmounts() (map[int64][]map[int64]string, error) {
	result := make(map[int64][]map[int64]string)
	rows, err := db.Query("SELECT currency_id, time, amount  FROM max_promised_amounts ORDER BY time ASC")
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id, time int64
		var amount string
		err = rows.Scan(&currency_id, &time, &amount)
		if err != nil {
			return result, err
		}
		result[currency_id] = append(result[currency_id], map[int64]string{time: amount})
	}
	return result, nil
}

func (db *DCDB) GetPrivateKey(myPrefix string) (string, error) {
	var key string
	key, err := db.Single("SELECT private_key FROM " + myPrefix + "my_keys WHERE block_id = (SELECT max(block_id) FROM " + myPrefix + "my_keys)").String()
	if err != nil {
		return "", ErrInfo(err)
	}
	return key, nil
}

func (db *DCDB) GetNodeConfig() (map[string]string, error) {
	return db.OneRow("SELECT * FROM config").String()
}

func (db *DCDB) Candidate_block() (*prevBlockType, int64, int64, int64, int64, [][][]int64, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()

	var minerId, userId, level, i, currentMinerId, currentUserId int64
	prevBlock := new(prevBlockType)
	var levelsRange [][][]int64
	// последний успешно записанный блок
	rows, err := db.Query(db.FormatQuery(`SELECT hex(hash), hex(head_hash), block_id, time, level FROM info_block`))
	if err != nil {
		return prevBlock, userId, minerId, currentUserId, level, levelsRange, ErrInfo(err)
	}
	defer rows.Close()
	if ok := rows.Next(); ok {
		err = rows.Scan(&prevBlock.Hash, &prevBlock.HeadHash, &prevBlock.BlockId, &prevBlock.Time, &prevBlock.Level)
		if err != nil {
			return prevBlock, userId, minerId, currentUserId, level, levelsRange, ErrInfo(err)
		}
	}
	log.Debug(db.FormatQuery(`SELECT hex(hash), hex(head_hash), block_id, time, level FROM info_block`))
	log.Debug("prevBlock: %v (%v)", prevBlock, GetParent())

	// общее кол-во майнеров
	maxMinerId, err := db.Single("SELECT max(miner_id) FROM miners").Int64()
	if err != nil {
		return prevBlock, userId, minerId, currentUserId, level, levelsRange, ErrInfo(err)
	}
	log.Debug("maxMinerId: %v (%v)", maxMinerId, GetParent())

	for currentUserId == 0 {
		// если майнера заморозили то у него исчезает miner_id, чтобы не попасть на такой пустой miner_id
		// нужно пербирать энтропию, пока не дойдем до существующего miner_id
		var entropy int64
		if i == 0 {
			entropy = GetEntropy(prevBlock.HeadHash)
			log.Debug("entropy: %v (%v)", entropy, GetParent())
		} else {
			time.Sleep(1000 * time.Millisecond)

			blockId := prevBlock.BlockId - i
			if blockId < 1 {
				break
			}

			newHeadHash, err := db.Single("SELECT hex(head_hash) FROM block_chain  WHERE id = ?", blockId).String()
			log.Debug("newHeadHash: %v (%v)", newHeadHash, GetParent())
			if err != nil {
				return prevBlock, userId, minerId, currentUserId, level, levelsRange, ErrInfo(err)
			}
			entropy = GetEntropy(newHeadHash)
		}
		currentMinerId = GetBlockGeneratorMinerId(maxMinerId, entropy)
		log.Debug("currentMinerId: %v (%v)", currentMinerId, GetParent())

		// получим ID юзера по его miner_id
		currentUserId, err = db.Single("SELECT user_id  FROM miners_data  WHERE miner_id = " + strconv.FormatInt(currentMinerId, 10)).Int64()
		if err != nil {
			return prevBlock, userId, minerId, currentUserId, level, levelsRange, ErrInfo(err)
		}
		i++
	}

	collective, err := db.GetMyUsersIds(true, true)
	if err != nil {
		return prevBlock, userId, minerId, currentUserId, level, levelsRange, ErrInfo(err)
	}
	log.Debug("collective: %v (%v)", collective, GetParent())

	// в сингл-моде будет только $my_miners_ids[0]
	myMinersIds, err := db.GetMyMinersIds(collective)
	if err != nil {
		return prevBlock, userId, minerId, currentUserId, level, levelsRange, ErrInfo(err)
	}
	log.Debug("myMinersIds: %v (%v)", myMinersIds, GetParent())

	// есть ли кто-то из нашего пула (или сингл-мода), кто находится на 0-м уровне
	if InSliceInt64(currentMinerId, myMinersIds) {
		level = 0
		levelsRange = append(levelsRange, [][]int64{{1, 1}})
		minerId = currentMinerId
	} else {
		levelsRange = GetBlockGeneratorMinerIdRange(currentMinerId, maxMinerId)
		log.Debug("levelsRange %v (%v)", levelsRange, GetParent())
		log.Debug("myMinersIds %v (%v)", myMinersIds, GetParent())
		if len(myMinersIds) > 0 {
			minerId, level = FindMinerIdLevel(myMinersIds, levelsRange)
		} else {
			level = -1 // у нас нет уровня, т.к. пуст $my_miners_ids, т.е. на сервере нет майнеров
			minerId = 0
		}
	}

	log.Debug("minerId: %v (%v)", minerId, GetParent())

	if minerId > 0 {
		userId, err = db.Single("SELECT user_id FROM miners_data WHERE miner_id = ?", minerId).Int64()
		if err != nil {
			return prevBlock, userId, minerId, currentUserId, level, levelsRange, ErrInfo(err)
		}
	} else {
		userId = 0
	}
	log.Debug("return (%v)", GetParent())
	return prevBlock, userId, minerId, currentUserId, level, levelsRange, nil
}

func (db *DCDB) GetSleepData() (map[string][]int64, error) {
	sleepDataMap := make(map[string][]int64)
	var sleepDataJson []byte
	sleepDataJson, err := db.Single("SELECT value FROM variables WHERE name = 'sleep'").Bytes()
	if err != nil {
		return sleepDataMap, ErrInfo(err)
	}
	if len(sleepDataJson) > 0 {
		err = json.Unmarshal(sleepDataJson, &sleepDataMap)
		if err != nil {
			return sleepDataMap, ErrInfo(err)
		}
	}
	log.Debug("sleepDataMap: %v", sleepDataMap)
	return sleepDataMap, nil
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

type DCAmounts struct {
	Tdc        float64
	Amount     float64
	PctSec     float64
	CurrencyId int64
	Pct        float64
}

type PromisedAmounts struct {
	Id                 int64
	Pct                float64
	PctSec             float64
	CurrencyId         int64
	Amount             float64
	MaxAmount          float64
	MaxOtherCurrencies int64
	StatusText         string
	Tdc                float64
	TdcAmount          float64
	Status             string
	InProcess          bool
}

func (db *DCDB) GetPromisedAmounts(userId, cash_request_time int64) (int64, []PromisedAmounts, map[int]DCAmounts, error) {
	log.Debug("%v", "cash_request_time", cash_request_time)
	var actualizationPromisedAmounts int64
	var promisedAmountListAccepted []PromisedAmounts
	promisedAmountListGen := make(map[int]DCAmounts)
	rows, err := db.Query(db.FormatQuery("SELECT id, currency_id, status, tdc_amount, amount, del_block_id, tdc_amount_update FROM promised_amount WHERE user_id = ?"), userId)
	if err != nil {
		return 0, nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, currency_id, del_block_id, tdc_amount_update int64
		var tdc_amount, amount float64
		var status string
		err = rows.Scan(&id, &currency_id, &status, &tdc_amount, &amount, &del_block_id, &tdc_amount_update)
		if err != nil {
			return 0, nil, nil, err
		}
		log.Debug("%v", "GetPromisedAmounts: ", currency_id, status, tdc_amount, amount, del_block_id, tdc_amount_update)
		// есть ли просроченные запросы
		cashRequestPending, err := db.Single("SELECT status FROM cash_requests WHERE to_user_id = ? AND del_block_id = 0 AND for_repaid_del_block_id = 0 AND time < ? AND status = 'pending'", userId, time.Now().Unix()-cash_request_time).String()
		if err != nil {
			return 0, nil, nil, err
		}
		if len(cashRequestPending) > 0 && currency_id > 1 && status == "mining" {
			status = "for_repaid"
			// и заодно проверим, можно ли делать актуализацию обещанных сумм
			actualizationPromisedAmounts, err = db.Single("SELECT id FROM promised_amount WHERE status = 'mining' AND user_id = ? AND currency_id > 1 AND del_block_id = 0 AND del_mining_block_id = 0 AND (cash_request_out_time > 0 AND cash_request_out_time < ?)", userId, time.Now().Unix()-cash_request_time).Int64()
			if err != nil {
				return 0, nil, nil, err
			}
		}
		tdc := tdc_amount
		if del_block_id > 0 {
			continue
		}
		log.Debug("%v", "tdc", tdc)
		if status == "mining" {
			profit, err := db.CalcProfitGen(currency_id, amount+tdc_amount, userId, tdc_amount_update, time.Now().Unix(), "mining")
			log.Debug("%v", "profit", profit)
			if err != nil {
				return 0, nil, nil, err
			}
			tdc += profit
			log.Debug("%v", "tdc", tdc)
			tdc = Round(tdc, 9)
			log.Debug("%v", "tdc", tdc)
		} else if status == "repaid" {
			profit, err := db.CalcProfitGen(currency_id, tdc_amount, userId, tdc_amount_update, time.Now().Unix(), "repaid")
			if err != nil {
				return 0, nil, nil, err
			}
			tdc += profit
			tdc = Round(tdc, 9)
		} else {
			tdc = tdc_amount
		}

		status_text := status
		maxAmount, err := db.Single("SELECT amount FROM max_promised_amounts WHERE currency_id  =  ? ORDER BY block_id DESC", currency_id).Float64()
		if err != nil {
			return 0, nil, nil, err
		}
		maxOtherCurrencies, err := db.Single("SELECT max_other_currencies FROM currency WHERE id  =  ?", currency_id).Int64()
		if err != nil {
			return 0, nil, nil, err
		}
		// для WOC amount не учитывается. Вместо него берется max_promised_amount
		if currency_id == 1 {
			amount = maxAmount
		}
		// обещанная не может быть больше max_promised_amounts
		if amount >= maxAmount {
			amount = maxAmount
		}
		if status == "repaid" {
			amount = 0
		}
		// последний статус юзера
		pctStatus, err := db.Single("SELECT status FROM points_status WHERE user_id  =  ? ORDER BY time_start DESC", userId).String()
		if err != nil {
			return 0, nil, nil, err
		}
		if len(pctStatus) == 0 {
			pctStatus = "user"
		}
		pct, err := db.Single(db.FormatQuery("SELECT "+pctStatus+" FROM pct WHERE currency_id  =  ? ORDER BY block_id DESC"), currency_id).Float64()
		if err != nil {
			return 0, nil, nil, err
		}
		log.Debug("%v", "pct", pct, "currency_id", currency_id, "pctStatus", pctStatus)
		pct_sec := pct
		pct = Round((math.Pow(1+pct, 3600*24*365)-1)*100, 2)
		// тут accepted значит просто попало в блок
		promisedAmountListAccepted = append(promisedAmountListAccepted, PromisedAmounts{Id: id, Pct: pct, PctSec: pct_sec, CurrencyId: currency_id, Amount: amount, MaxAmount: maxAmount, MaxOtherCurrencies: maxOtherCurrencies, StatusText: status_text, Tdc: tdc, TdcAmount: tdc_amount, Status: status, InProcess: false})
		// для вывода на главную общей инфы
		promisedAmountListGen[int(currency_id)] = DCAmounts{Tdc: tdc, Amount: amount, PctSec: pct_sec, CurrencyId: (currency_id)}
	}
	return actualizationPromisedAmounts, promisedAmountListAccepted, promisedAmountListGen, nil
}

func (db *DCDB) GetMinerId(userId int64) (int64, error) {
	return db.Single("SELECT miner_id FROM miners_data WHERE user_id  =  ?", userId).Int64()
}

func (db *DCDB) GetMyMinersIds(collective []int64) ([]int64, error) {
	log.Debug("user_id IN %v", strings.Join(SliceInt64ToString(collective), ","))
	return db.GetList("SELECT miner_id FROM miners_data WHERE user_id IN (" + strings.Join(SliceInt64ToString(collective), ",") + ") AND miner_id > 0").Int64()
}

func (db *DCDB) GetConfirmedBlockId() (int64, error) {

		result, err := db.Single("SELECT max(block_id) FROM confirmations WHERE good >= ?", consts.MIN_CONFIRMED_NODES).Int64()
		if err != nil {
			return 0, err
		}
		//log.Debug("%v", "result int64",StrToInt64(result))
		return result, nil

}

func (db *DCDB) GetCommunityUsers() ([]int64, error) {
	var users []int64
	rows, err := db.Query("SELECT user_id FROM community")
	if err != nil {
		return users, ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var userId int64
		err = rows.Scan(&userId)
		if err != nil {
			return users, ErrInfo(err)
		}
		users = append(users, userId)
	}
	return users, err
}

func (db *DCDB) GetMyUserId(myPrefix string) (int64, error) {
	userId, err := db.Single("SELECT user_id FROM " + myPrefix + "my_table").Int64()
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (db *DCDB) GetMyCBIDAndWalletId() (int64, int64, error) {
	myCBID, err := db.GetMyCBID();
	if err != nil {
		return 0, 0, err
	}
	myWalletId, err := db.GetMyWalletId();
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
	delegate, err := db.OneRow("SELECT delegate_wallet_id, delegate_cb_id FROM central_banks WHERE cb_id = ?", myCBID).Int64()
	if err != nil {
		return false, err
	}
	// Если мы - ЦБ и у нас указан delegate, т.е. мы делегировали полномочия по поддержанию ноды другому юзеру или ЦБ, то выходим.
	if delegate["delegate_wallet_id"] > 0 || delegate["delegate_cb_id"] > 0 {
		return true, nil
	}
	return false, nil
}

func (db *DCDB) GetMyWalletId() (int64, error) {
	return db.Single("SELECT dlt_wallet_id FROM config").Int64()
}

func (db *DCDB) GetMyCBID() (int64, error) {
	return db.Single("SELECT cb_id FROM config").Int64()
}

func (db *DCDB) GetMyUsersIds(checkCommission, checkNodeKey bool) ([]int64, error) {
	var usersIds []int64
	usersIds, err := db.GetCommunityUsers()
	if err != nil {
		return usersIds, err
	}
	if len(usersIds) == 0 { // сингл-мод
		rows, err := db.Query("SELECT user_id FROM my_table")
		if err != nil {
			return usersIds, err
		}
		defer rows.Close()
		if ok := rows.Next(); ok {
			var x int64
			err = rows.Scan(&x)
			if err != nil {
				return usersIds, err
			}
			usersIds = append(usersIds, x)
		}
	} else {
		// нельзя допустить, чтобы блок подписал майнер, у которого комиссия больше той, что разрешана в пуле,
		// т.к. это приведет к попаднию в блок некорректной тр-ии, что приведет к сбою пула
		if checkCommission {
			// комиссия на пуле
			commissionJson, err := db.Single("SELECT commission FROM config").Bytes()
			if err != nil {
				return usersIds, err
			}
			if err != nil {
				return usersIds, err
			}
			var commissionPoolMap map[string][]float64
			err = json.Unmarshal(commissionJson, &commissionPoolMap)
			if err != nil {
				return usersIds, err
			}
			rows2, err := db.Query("SELECT user_id, commission FROM commission WHERE user_id IN (" + strings.Join(SliceInt64ToString(usersIds), ",") + ")")
			if err != nil {
				return usersIds, err
			}
			defer rows2.Close()
			for rows2.Next() {
				var uid int64
				var commJson []byte
				err = rows2.Scan(&uid, &commJson)
				if err != nil {
					return usersIds, err
				}
				if len(commJson) > 0 {
					var commissionUserMap map[string][]float64
					err := json.Unmarshal(commJson, &commissionUserMap)
					if err != nil {
						return usersIds, err
					}
					for currencyId, Commissions := range commissionUserMap {
						if len(commissionPoolMap[currencyId]) > 0 {
							if Commissions[0] > commissionPoolMap[currencyId][0] || Commissions[1] > commissionPoolMap[currencyId][1] {
								log.Debug("DelUserIdFromArray %v > %v || %v > %v / %v", Commissions[0], commissionPoolMap[currencyId][0], Commissions[1], commissionPoolMap[currencyId][1], uid)
								DelUserIdFromArray(&usersIds, uid)
							}
						}
					}
				}
			}
		}
		// нельзя чтобы блок сгенерировал майнер, чьего нодовского приватного ключа нет у нас,
		// т.к. это приведет к ступору в candidateBlockIsReady в проверке подписи
		if checkNodeKey {

			rows, err := db.Query("SELECT user_id, node_public_key FROM miners_data WHERE user_id IN (" + strings.Join(SliceInt64ToString(usersIds), ",") + ")")
			if err != nil {
				return usersIds, err
			}
			defer rows.Close()
			for rows.Next() {
				var uid, nodePublicKey string
				err = rows.Scan(&uid, &nodePublicKey)
				if err != nil {
					return usersIds, err
				}

				publicKey, err := db.GetMyNodePublicKey(uid + "_")
				if err != nil {
					return usersIds, err
				}
				if publicKey != nodePublicKey {
					log.Debug("publicKey != nodePublicKey (%d)", uid)
					DelUserIdFromArray(&usersIds, StrToInt64(uid))
					//log.Debug("DelUserIdFromArray publicKey != nodePublicKey (%x != %x) %v / %v", publicKey, nodePublicKey, uid, usersIds)
				}
			}
		}
	}
	return usersIds, nil
}

func (db *DCDB) GetBlockId() (int64, error) {
	return db.Single("SELECT block_id FROM info_block").Int64()
}

func (db *DCDB) GetMyBlockId() (int64, error) {
	return db.Single("SELECT my_block_id FROM config").Int64()
}

// наличие cash_requests с pending означает, что у юзера все обещанные суммы в for_repaid. Возможно, временно, если это свежий запрос и юзер еще не успел послать cash_requests_in
func (db *DCDB) CheckCashRequests(userId int64) error {
	cashRequestStatus, err := db.Single("SELECT status FROM cash_requests WHERE to_user_id  =  ? AND del_block_id  =  0 AND for_repaid_del_block_id  =  0 AND status  =  'pending'", userId).String()
	if err != nil {
		return err
	}
	if len(cashRequestStatus) > 0 {
		log.Debug("%v", "cashRequestStatus")
		return fmt.Errorf("cashRequestStatus")
	}
	return nil
}

func (db *DCDB) CheckUser(userId int64) error {
	user_id, err := db.Single("SELECT user_id FROM users WHERE user_id = ?", userId).Int64()
	if err != nil {
		return err
	}
	if user_id > 0 {
		return nil
	} else {
		return fmt.Errorf("user_id is null")
	}
}

func (db *DCDB) GetPct() (map[int64][]map[int64]map[string]float64, error) {
	result := make(map[int64][]map[int64]map[string]float64)
	var q string
	if db.ConfigIni["db_type"] == "postgresql" {
		q = `SELECT currency_id, time, "user", miner FROM pct ORDER BY time ASC`
	} else {
		q = `SELECT currency_id, time, user, miner FROM pct ORDER BY time ASC`
	}
	rows, err := db.Query(q)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id, time int64
		var user, miner float64
		err = rows.Scan(&currency_id, &time, &user, &miner)
		if err != nil {
			return result, err
		}
		result[currency_id] = append(result[currency_id], map[int64]map[string]float64{time: {"miner": miner, "user": user}})
	}
	return result, nil
}

func (db *DCDB) CheckCurrencyId(id int64) (int64, error) {
	return db.Single("SELECT id FROM currency WHERE id = ?", id).Int64()
}

/*
func(db *DCDB) GetRepaidAmount(userId, currencyId int64) (float64, error) {
	return db.Single("SELECT amount FROM promised_amount WHERE status = 'repaid' AND currency_id = ? AND user_id = ?", currencyId, userId).Float64()
}
*/
func (db *DCDB) GetHolidays(userId int64) ([][]int64, error) {
	var result [][]int64
	sql := "SELECT start_time, end_time FROM holidays WHERE user_id = ? AND del = 0"
	rows, err := db.Query(db.FormatQuery(sql), userId)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var start_time, end_time int64
		err = rows.Scan(&start_time, &end_time)
		if err != nil {
			return result, err
		}
		result = append(result, []int64{start_time, end_time})
	}
	return result, nil
}

func (db *DCDB) GetRepaidAmount(currencyId, userId int64) (float64, error) {
	amount, err := db.Single("SELECT amount FROM promised_amount WHERE status = 'repaid' AND currency_id = ? AND user_id = ? AND del_block_id = 0 AND del_mining_block_id = 0", currencyId, userId).Float64()
	if err != nil {
		return 0, err
	}
	return amount, nil
}

func (db *DCDB) CheckCurrency(currency_id int64) (bool, error) {
	id, err := db.Single("SELECT id FROM currency WHERE id = ?", currency_id).Int()
	if err != nil {
		return false, err
	}
	if id == 0 {
		return false, nil
	} else {
		return true, nil
	}
}
func (db *DCDB) CheckCurrencyCF(currency_id int64) (bool, error) {
	id, err := db.Single("SELECT id FROM cf_currency WHERE id = ?", currency_id).Int()
	if err != nil {
		return false, err
	}
	if id == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (db *DCDB) GetWalletIdByPublicKey(publicKey []byte) (int64, error) {
	walletId, err := db.Single(`SELECT wallet_id FROM dlt_wallets WHERE lower(hex(address)) = ?`, string(HashSha1Hex(publicKey))).Int64()
	if err != nil {
		return 0, ErrInfo(err)
	}
	return walletId, nil
}

func (db *DCDB) GetCitizenIdByPublicKey(publicKey []byte) (int64, error) {
	walletId, err := db.Single(`SELECT citizen_id FROM citizens WHERE hex(public_key_0) = ?`, string(publicKey)).Int64()
	if err != nil {
		return 0, ErrInfo(err)
	}
	return walletId, nil
}

func (db *DCDB) InsertIntoMyKey(prefix string, publicKey []byte, curBlockId string) error {
	err := db.ExecSql(`INSERT INTO `+prefix+`my_keys (public_key, status, block_id) VALUES ([hex],'approved', ?)`, publicKey, curBlockId)
	if err != nil {
		return err
	}
	return nil
}

func (db *DCDB) GetPaymentSystems() (map[string]string, error) {
	return db.GetMap(`SELECT id, name FROM payment_systems ORDER BY name`, "id", "name")
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

func (db *DCDB) GetcandidateBlockId() (int64, error) {
	rows, err := db.Query("SELECT block_id FROM candidateBlock")
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if ok := rows.Next(); ok {
		var block_id int64
		err = rows.Scan(&block_id)
		if err != nil {
			return 0, err
		}
		return block_id, nil
	}
	return 0, nil
}

func (db *DCDB) GetMyPrefix(userId int64) (string, error) {
	collective, err := db.GetCommunityUsers()
	if err != nil {
		return "", ErrInfo(err)
	}
	if len(collective) == 0 {
		return "", nil
	} else {
		/*myUserId, err := db.GetPoolAdminUserId()
		if err != nil || myUserId == 0  {
			if err != nil {
				return "", ErrInfo(err)
			} else {
				return "", fmt.Errorf("myUserId==0")
			}
		}*/
		return Int64ToStr(userId) + "_", nil
	}
}


func (db *DCDB) GetNodePublicKey(userId int64) ([]byte, error) {
	result, err := db.Single("SELECT node_public_key FROM miners_data WHERE user_id = ?", userId).Bytes()
	if err != nil {
		return []byte(""), err
	}
	return result, nil
}
func (db *DCDB) GetNodePublicKeyWalletOrCB(wallet_id, cb_id int64) ([]byte, error) {
	var result []byte
	var err error
	if wallet_id > 0 {
		result, err = db.Single("SELECT node_public_key FROM dlt_wallets WHERE wallet_id = ?", wallet_id).Bytes()
		if err != nil {
			return []byte(""), err
		}
	} else {
		result, err = db.Single("SELECT node_public_key FROM central_banks WHERE cb_id = ?", cb_id).Bytes()
		if err != nil {
			return []byte(""), err
		}
	}
	return result, nil
}

func (db *DCDB) GetCountCurrencies() (int64, error) {
	result, err := db.Single("SELECT count(id) FROM currency").Int64()
	if err != nil {
		return 0, err
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

func (db *DCDB) DeleteQueueBlock(head_hash_hex, hash_hex string) error {
	return db.ExecSql("DELETE FROM queue_blocks WHERE hex(head_hash) = ? AND hex(hash) = ?", head_hash_hex, hash_hex)
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
	return db.ExecSql("DELETE FROM main_lock WHERE script_name=?", name)
}

func (db *DCDB) GetIsReadySleep(level int64, data []int64) int64 {
	//SleepData := db.GetSleepData();
	return GetIsReadySleep0(level, data)
}

func (db *DCDB) GetGenSleep(prevBlock *prevBlockType, level int64) (int64, error) {

	sleepData, err := db.GetSleepData()
	if err != nil {
		return 0, err
	}

	// узнаем время, которые было затрачено в ожидании is_ready предыдущим блоком
	isReadySleep := db.GetIsReadySleep(prevBlock.Level, sleepData["is_ready"])
	//fmt.Println("isReadySleep", isReadySleep)

	// сколько сек должен ждать нод, перед тем, как начать генерить блок, если нашел себя в одном из уровней.
	generatorSleep := GetGeneratorSleep(level, sleepData["generator"])
	//fmt.Println("generatorSleep", generatorSleep)

	// сумма is_ready всех предыдущих уровней, которые не успели сгенерить блок
	isReadySleep2 := GetIsReadySleepSum(level, sleepData["is_ready"])
	//fmt.Println("isReadySleep2", isReadySleep2)

	// узнаем, сколько нам нужно спать
	sleep := isReadySleep + generatorSleep + isReadySleep2
	return sleep, nil
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
					err = db.QueryRow(db.FormatQuery("SELECT log_id FROM " + table)).Scan(&exists)
					if err != nil {
						if ok, _ := regexp.MatchString(`no rows`, fmt.Sprintf("%s", err)); ok {
							column = "log_id"
						} else {
							return "", ErrInfo(err)
						}
					}
					column = "log_id"
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
				err := db.QueryRow("SELECT column_name FROM information_schema.columns WHERE table_name=$1 and column_name=$2", table, "log_id").Scan(&exists)
				if err != nil {
					return "", err
				}
				if len(exists) == 0 {
					return "", fmt.Errorf("no id, log_id")
				}
				column = "log_id"
			}
		case "mysql":
			exists = ""
			err := db.QueryRow("SELECT TABLE_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name=? and column_name=?", db.ConfigIni["db_name"], table, "id").Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				return "", err
			}
			if len(exists) == 0 {
				err := db.QueryRow("SELECT TABLE_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name=? and column_name=?", db.ConfigIni["db_name"], table, "log_id").Scan(&exists)
				if err != nil {
					return "", err
				}
				if len(exists) == 0 {
					return "", fmt.Errorf("no id, log_id")
				}
				column = "log_id"
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
		BlockData.HeadHash = BinToHex([]byte(data["head_hash"]))
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
							UNION
							SELECT hash
							FROM transactions_candidate_block
							WHERE type = ?
										  `+addSql+`
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
							UNION
							SELECT hash
							FROM transactions_candidate_block
							WHERE type IN (`+whereType+`)
										 `+addSql+` `+addSql1+` AND
										 citizen_id = ?
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

func GetTxTypeAndUserId(binaryBlock []byte) (int64, int64, int64, int64) {
	var thirdVar int64
	txType := BinToDecBytesShift(&binaryBlock, 1)
	BytesShift(&binaryBlock, 4) // уберем время
	walletId := BytesToInt64(BytesShift(&binaryBlock, DecodeLength(&binaryBlock)))
	citizenId := BytesToInt64(BytesShift(&binaryBlock, DecodeLength(&binaryBlock)))
	// thirdVar - нужен тогда, когда нужно недопустить попадание в блок несовместимых тр-ий.
	// Например, удаление крауд-фандинг проекта и инвестирование в него средств.
	if InSliceInt64(txType, TypesToIds([]string{"CfSendDc", "DelCfProject"})) {
		thirdVar = BytesToInt64(BytesShift(&binaryBlock, DecodeLength(&binaryBlock)))
	}
	log.Debug("txType, userId, thirdVar %v, %v, %v, %v", txType, walletId, citizenId, thirdVar)
	return txType, walletId, citizenId, thirdVar
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

	collective, err := db.GetCommunityUsers()
	if err != nil {
		return nil, nil, nil, err
	}
	if len(collective) > 0 {
		if !InSliceInt64(myUserId, collective) {
			return nil, nil, nil, ErrInfo(fmt.Sprintf("!InSliceInt64(myUserId, collective) %d %v", myUserId, collective))
		}
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

func (db *DCDB) GetBinSign(forSign string, myUserId int64) ([]byte, error) {


	nodePrivateKey, err := db.GetNodePrivateKey()
	if err != nil {
		return nil, ErrInfo(err)
	}
	log.Debug("nodePrivateKey = %s", nodePrivateKey)
	// подписываем нашим нод-ключем данные транзакции
	privateKey, err := MakePrivateKey(nodePrivateKey)
	if err != nil {
		return nil, ErrInfo(err)
	}
	return rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA1, HashSha1(forSign))
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

func (db *DCDB) CheckChatMessage(message string, sender, receiver, lang, room, status, signTime int64, signature []byte) error {

	if sender <= 0 || sender > 16777215 {
		return ErrInfoFmt("incorrect sender")
	}
	if receiver < 0 || receiver > 16777215 {
		return ErrInfoFmt("incorrect receiver")
	}
	if lang <= 0 || lang > 255 {
		return ErrInfoFmt("incorrect lang")
	}
	if room < 0 || room > 16777215 {
		return ErrInfoFmt("incorrect room")
	}
	if status != 0 && status != 1 {
		return ErrInfoFmt("incorrect status")
	}
	if signTime < 0 || signTime > 4294967295 {
		return ErrInfoFmt("incorrect room")
	}
	// chatEncrypted == 1
	if len(message) == 0 || ((status == 0 && len(message) > 1024) || (status == 1 && len(message) > 5120)) {
		return ErrInfoFmt("incorrect message")
	}
	if len(signature) < 128 || len(signature) > 5120 {
		return ErrInfoFmt("incorrect signature")
	}

	if receiver > 0 {
		// проверим, есть ли такой юзер и заодно получим public_key
		publicKey, err := db.Single("SELECT public_key_0 FROM users WHERE user_id = ?", receiver).Bytes()
		if err != nil {
			return ErrInfo(err)
		}
		if len(publicKey) == 0 {
			return ErrInfoFmt("incorrect receiver")
		}
	}

	publicKey, err := db.Single("SELECT public_key_0 FROM users WHERE user_id = ?", sender).Bytes()
	if err != nil {
		return ErrInfo(err)
	}
	if len(publicKey) == 0 {
		return ErrInfoFmt("incorrect sender. null public_key_0")
	}

	// проверяем подпись
	forSign := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v", lang, room, receiver, sender, status, message, signTime)
	CheckSignResult, err := CheckSign([][]byte{publicKey}, forSign, HexToBin(signature), true)
	if err != nil {
		return ErrInfo(err)
	}
	if !CheckSignResult {
		return ErrInfoFmt("incorrect signature %s", forSign)
	}

	// нельзя за сутки слать более X сообщений
	count, err := db.Single(`SELECT count(hash) FROM chat WHERE sender = ? AND time > ?`, sender, Time()-86400).Int64()
	if count > 100 {
		return ErrInfoFmt(">100 messages per 24h from %d", sender)
	}

	// нет ли бана от админа
	ban, err := db.Single(`SELECT time_start+sec FROM chat_ban WHERE user_id = ? AND time_start+sec > ?`, sender, Time()).Int64()
	if ban > 0 {
		return ErrInfoFmt("ban %d. remaing %d seconds", sender, ban-Time())
	}

	return nil
}

func (db *DCDB) GetPromisedAmountCounter(userId int64) ( float64, float64, error) {
	paRestricted, err := db.OneRow("SELECT * FROM promised_amount_restricted WHERE user_id = ?", userId).String()
	if err != nil {
		return 0, 0, err
	}
	if _, ok := paRestricted[`user_id`]; !ok {
		return 0, 0, nil
	}
	
	amount := StrToFloat64(paRestricted["amount"])
	// Временная проверка для старого формата таблицы promised_amount_restricted. 
	if _, ok := paRestricted["start_time"]; ok && StrToInt64(paRestricted["last_update"]) == 0 {
		paRestricted["last_update"] = paRestricted["start_time"]
	}
	profit, err := db.CalcProfitGen(StrToInt64(paRestricted["currency_id"]), amount, userId, StrToInt64(paRestricted["last_update"]), Time(), "wallet")
	if err != nil {
		return 0, 0, err
	}
	profit += amount
	
	pct, err := db.Single(db.FormatQuery("SELECT user FROM pct WHERE currency_id  =  ? ORDER BY block_id DESC"), StrToInt64(paRestricted["currency_id"])).Float64()
	if err != nil {
		return 0, 0, err
	}
	return profit, pct, nil
}

func (db *DCDB) GetNotificationsCount(userId int64) (int64, error) {
	return db.Single("SELECT count(id) FROM notifications WHERE user_id=? AND isread=1", userId ).Int64()
}
