package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"

	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	log "github.com/sirupsen/logrus"
)

var (
	DBConn         *gorm.DB
	RecordNotFound = gorm.ErrRecordNotFound
)

func GormInit(user string, pass string, dbName string) error {
	connect, err := gorm.Open("postgres",
		fmt.Sprintf("host=localhost user=%s dbname=%s sslmode=disable password=%s", user, dbName, pass))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("cant open connection to DB")
		return err
	}
	DBConn = connect
	return nil
}

func GormClose() error {
	if DBConn != nil {
		return DBConn.Close()
	}
	return nil
}

type DbTransaction struct {
	conn *gorm.DB
}

func StartTransaction() (*DbTransaction, error) {
	conn := DBConn.Begin()
	if conn.Error != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": conn.Error}).Error("cannot start transaction because of connection error")
		return nil, conn.Error
	}

	return &DbTransaction{
		conn: conn,
	}, nil
}

func (tr *DbTransaction) Rollback() {
	tr.conn.Rollback()
}

func (tr *DbTransaction) Commit() error {
	return tr.conn.Commit().Error
}

func GetDB(tr *DbTransaction) *gorm.DB {
	if tr != nil && tr.conn != nil {
		return tr.conn
	}
	return DBConn
}

func DropTables() error {
	return DBConn.Exec(`
	DO $$ DECLARE
	    r RECORD;
	BEGIN
	    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
		EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
	    END LOOP;
	END $$;
	`).Error
}

func GetRecordsCount(tableName string) (int64, error) {
	var count int64
	err := DBConn.Table(tableName).Count(&count).Error
	return count, err
}

func ExecSchemaEcosystem(id int, wallet int64, name string) error {
	schema, err := static.Asset("static/schema-ecosystem-v2.sql")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("getting schema from static asset")
		return err
	}
	err = DBConn.Exec(fmt.Sprintf(string(schema), id, wallet, name)).Error
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing ecosystem schema")
		return err
	}
	if id == 1 {
		schema, err = static.Asset("static/schema-firstecosystem-v2.sql")
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("getting schema for first ecosystem")
			return err
		}
		err = DBConn.Exec(fmt.Sprintf(string(schema), wallet)).Error
	}
	return err
}

func ExecSchema() error {
	schema, err := static.Asset("static/schema-v2.sql")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting static/schema-v2.sql asset")
		os.Remove(*utils.Dir + "/config.ini")
		return err
	}
	return DBConn.Exec(string(schema)).Error
}

func GetColumnsCount(tableName string) (int64, error) {
	var count int64
	err := DBConn.Table("information_schema.columns").
		Where("table_name=?", tableName).
		Select("column_name").
		Count(&count).Error
	return count, err
}

func GetTables() ([]string, error) {
	var result []string
	err := DBConn.Table("information_schema.tables").
		Where("table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')").
		Select("table_name").Scan(&result).Error
	return result, err
}

func Update(transaction *DbTransaction, tblname, set, where string) error {
	return GetDB(transaction).Exec(`UPDATE "` + strings.Trim(tblname, `"`) + `" SET ` + set + " " + where).Error
}

func Delete(tblname, where string) error {
	return DBConn.Exec(`DELETE FROM "` + tblname + `" ` + where).Error
}

func InsertReturningLastID(transaction *DbTransaction, table, columns, values string) (string, error) {
	var result string
	returning, err := GetFirstColumnName(table)
	if err != nil {
		return "", err
	}
	insertQuery := `INSERT INTO "` + table + `" (` + columns + `) VALUES (` + values + `) RETURNING ` + returning

	err = GetDB(transaction).Raw(insertQuery).Row().Scan(&result)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": insertQuery}).Error("scanning result for query")
		return "", err
	}
	return result, nil
}

func SequenceRestartWith(seqName string, id int64) error {
	return DBConn.Exec("ALTER SEQUENCE " + seqName + " RESTART WITH " + converter.Int64ToStr(id)).Error
}

func SequenceLastValue(seqName string) (int64, error) {
	var result int64
	if err := DBConn.Raw("SELECT last_value FROM " + seqName).Row().Scan(&result); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("scanning last_value into result")
		return 0, err
	}
	return result, nil
}

func GetSerialSequence(table, AiID string) (string, error) {
	var result string
	query := `SELECT pg_get_serial_sequence('` + table + `', '` + AiID + `')`
	err := DBConn.Raw(query).Row().Scan(&result)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("scanning last sequence into result")
		return "", err
	}
	return result, nil
}

func GetCurrentSeqID(id, tblname string) (int64, error) {
	var result int64
	query := "SELECT " + id + " FROM " + tblname + " ORDER BY " + id + " DESC LIMIT 1"
	err := DBConn.Raw(query).Row().Scan(&result)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("getting current sequence id")
		return 0, err
	}
	return result, nil
}

func GetRollbackID(tblname, where, ordering string) (int64, error) {
	var result int64
	query := `SELECT rb_id FROM "` + tblname + `" ` + where + " order by rb_id " + ordering
	err := DBConn.Raw(query).Row().Scan(&result)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("selecting rollback id from table")
		return 0, nil
	}
	return result, nil
}

func GetFirstColumnName(table string) (string, error) {
	query := `SELECT * FROM "` + table + `" LIMIT 1`
	rows, err := DBConn.Raw(query).Rows()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("selecting rollback id from table")
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("getting rows columns")
		return "", err
	}
	if len(columns) > 0 {
		return columns[0], nil
	}
	return "", nil
}

func GetQueryTotalCost(query string, args ...interface{}) (int64, error) {
	var planStr string
	err := DBConn.Raw(fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", query), args...).Row().Scan(&planStr)
	switch {
	case err == sql.ErrNoRows:
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("no rows while explaining query")
		return 0, errors.New("No rows")
	case err != nil:
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": query}).Error("error explaining query")
		return 0, err
	}
	var queryPlan []map[string]interface{}
	dec := json.NewDecoder(strings.NewReader(planStr))
	dec.UseNumber()
	if err := dec.Decode(&queryPlan); err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("decoding query plan from JSON")
		return 0, err
	}
	if len(queryPlan) == 0 {
		log.Error("Query plan is empty")
		return 0, errors.New("Query plan is empty")
	}
	firstNode := queryPlan[0]
	var plan interface{}
	var ok bool
	if plan, ok = firstNode["Plan"]; !ok {
		log.Error("No Plan key in result")
		return 0, errors.New("No Plan key in result")
	}
	var planMap map[string]interface{}
	if planMap, ok = plan.(map[string]interface{}); !ok {
		log.Error("Plan is not map[string]interface{}")
		return 0, errors.New("Plan is not map[string]interface{}")
	}
	if totalCost, ok := planMap["Total Cost"]; ok {
		if totalCostNum, ok := totalCost.(json.Number); ok {
			if totalCostF64, err := totalCostNum.Float64(); err != nil {
				log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("converting total cost num to float64")
				return 0, err
			} else {
				return int64(totalCostF64), nil
			}
		} else {
			log.Error("Total cost is not a number")
			return 0, errors.New("Total cost is not a number")
		}
	} else {
		log.Error("Plan map has no TotalCost")
		return 0, errors.New("PlanMap has no TotalCost")
	}
	return 0, nil
}

func GetAllTables() ([]string, error) {
	var result []string
	sql := `SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')`
	rows, err := DBConn.Raw(sql).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tblname string
		if err := rows.Scan(&tblname); err != nil {
			return nil, err
		}
		result = append(result, tblname)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func GetColumnCount(tableName string) (int64, error) {
	var count int64
	err := DBConn.Raw("SELECT count(*) FROM information_schema.columns WHERE table_name=?", tableName).Row().Scan(&count)
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return count, nil
}

func SendTx(txType int64, adminWallet int64, data []byte) ([]byte, error) {
	hash, err := crypto.Hash(data)
	if err != nil {
		return nil, err
	}
	ts := &TransactionStatus{
		Hash:     hash,
		Time:     time.Now().Unix(),
		Type:     txType,
		WalletID: adminWallet,
	}
	err = ts.Create()
	if err != nil {
		return nil, err
	}
	qtx := &QueueTx{
		Hash: hash,
		Data: data,
	}
	err = qtx.Create()
	return hash, err
}

func GetLastBlockData() (map[string]int64, error) {
	result := make(map[string]int64)
	confirmation := &Confirmation{}
	err := confirmation.GetMaxGoodBlock()
	if err != nil {
		return result, utils.ErrInfo(err)
	}
	confirmedBlockID := confirmation.BlockID
	if confirmedBlockID == 0 {
		confirmedBlockID = 1
	}
	// obtain the time of the last affected block
	block := &Block{}
	err = block.GetBlock(confirmedBlockID)
	if err != nil || len(block.Data) == 0 {
		return result, utils.ErrInfo(err)
	}
	result["blockId"] = block.ID
	// the time of the last block
	result["lastBlockTime"] = block.Time
	return result, nil
}

func GetMyWalletID() (int64, error) {
	conf := &Config{}
	err := conf.GetConfig()
	if err != nil {
		return 0, err
	}
	walletID := conf.DltWalletID
	if walletID == 0 {
		walletID = converter.StringToAddress(*utils.WalletAddress)
	}
	return walletID, nil
}

func AlterTableAddColumn(transaction *DbTransaction, tableName, columnName, columnType string) error {
	return GetDB(transaction).Exec(`ALTER TABLE "` + tableName + `" ADD COLUMN ` + columnName + ` ` + columnType).Error
}

func AlterTableDropColumn(tableName, columnName string) error {
	return DBConn.Exec(`ALTER TABLE "` + tableName + `" DROP COLUMN ` + columnName).Error
}

func CreateIndex(transaction *DbTransaction, indexName, tableName, onColumn string) error {
	return GetDB(transaction).Exec(`CREATE INDEX "` + indexName + `_index" ON "` + tableName + `" (` + onColumn + `)`).Error
}

func IsTable(tblname string) bool {
	var name string
	DBConn.Table("information_schema.tables").
		Where("table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema') AND table_name=?`", tblname).
		Select("table_name").Row().Scan(&name)

	return name == tblname
}

func GetColumnDataTypeCharMaxLength(tableName, columnName string) (map[string]string, error) {
	/*	var dataType string
		var characterMaximumLength string
			rows, err := DBConn.
			Table("information_schema.columns").
			Where("table_name = ? AND column_name = ?", tableName, columnName).
			Select("data_type", "character_maximum_length").Rows()*/
	return GetOneRow(`select data_type,character_maximum_length from
			information_schema.columns where table_name = ? AND column_name = ?`,
		tableName, columnName).String()
	/*	if err != nil {
			return nil, err
		}
		for rows.Next() {
			rows.Scan(&dataType)
			rows.Scan(&characterMaximumLength)
			fmt.Println(`COLDATA`, dataType, characterMaximumLength)
		}
		rows.Close()
		result := make(map[string]string, 0)
		result["data_type"] = dataType
		result["character_maximum_length"] = characterMaximumLength
		return row, nil*/
}

func GetColumnType(tblname, column string) (itype string, err error) {
	coltype, err := GetColumnDataTypeCharMaxLength(tblname, column)
	if err != nil {
		return
	}
	if dataType, ok := coltype["data_type"]; ok {
		switch {
		case dataType == "character varying":
			itype = `varchar`
			/*		case coltype[`data_type`] == "bytea":
					itype = "bytea"*/
		case dataType == `bigint`:
			itype = "number"
		case strings.HasPrefix(dataType, `timestamp`):
			itype = "datetime"
		case strings.HasPrefix(dataType, `numeric`):
			itype = "money"
		case strings.HasPrefix(dataType, `double`):
			itype = "double"
		default:
			itype = dataType
		}
	}
	return
}

func GetSleepTime(myWalletID, myStateID, prevBlockStateID, prevBlockWalletID int64) (int64, error) {
	// take the list of all full_nodes

	node := &FullNode{}
	fullNodes, err := node.GetAll()
	if err != nil {
		return 0, err
	}
	fullNodesList := make([]map[string]string, 0, len(*fullNodes))
	for _, node := range *fullNodes {
		nodeMap := node.ToMap()
		fullNodesList = append(fullNodesList, nodeMap)
	}

	// determine full_node_id of the one, who had to generate a block (but could delegate this)
	err = node.Get(prevBlockWalletID)
	if err != nil {
		return 0, err
	}
	prevBlockFullNodeID := node.ID
	prevBlockFullNodePosition := func(fullNodesList []map[string]string, prevBlockFullNodeID int64) int {
		for i, fullNodes := range fullNodesList {
			id, err := strconv.ParseInt(fullNodes["id"], 10, 64)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": fullNodes["id"]}).Error("converting full nodes id from string to int")
			}
			if id == prevBlockFullNodeID {
				return i
			}
		}
		return -1
	}(fullNodesList, int64(prevBlockFullNodeID))

	// define our place (Including in the 'delegate')
	myPosition := func(fullNodesList []map[string]string, myWalletID, myStateID int64) int {
		for i, fullNodes := range fullNodesList {
			stateID, err := strconv.ParseInt(fullNodes["state_id"], 10, 64)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": fullNodes["state_id"]}).Error("converting full nodes state_id from string to int")
			}
			walletID, err := strconv.ParseInt(fullNodes["wallet_id"], 10, 64)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": fullNodes["wallet_id"]}).Error("converting full nodes wallet_id from string to int")
			}
			if stateID == myStateID || walletID == myWalletID {
				return i
			}
		}
		return -1
	}(fullNodesList, myWalletID, myStateID)

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

	return int64(sleepTime), nil
}

func GetNameList(tableName string, count int) ([]map[string]string, error) {
	var names []string
	err := DBConn.Table(tableName).Order("name").Limit(count).Pluck("name", &names).Error
	if err != nil {
		return nil, err
	}
	result := make([]map[string]string, 0)
	for _, name := range names {
		line := make(map[string]string)
		line["name"] = name
		result = append(result, line)
	}
	return result, nil
}

func DropTable(transaction *DbTransaction, tableName string) error {
	return GetDB(transaction).DropTable(tableName).Error
}

func GetConditionsAndValue(tableName, name string) (map[string]string, error) {
	type proxy struct {
		Conditions string
		Value      string
	}
	var temp proxy
	err := DBConn.Table(tableName).Where("name = ?", name).Select("conditions", "value").Find(&temp).Error
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting conditions value")
		return nil, err
	}
	result := make(map[string]string)
	result["conditions"] = temp.Conditions
	result["value"] = temp.Value
	return result, nil
}

// Because of import cycle utils and config
func IsNodeState(state int64, host string) bool {
	if strings.HasPrefix(host, `localhost`) {
		return true
	}
	if val, ok := config.ConfigIni[`node_state_id`]; ok {
		if val == `*` {
			return true
		}
		for _, id := range strings.Split(val, `,`) {
			stateID, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("converting node state id from string to int")
			}
			if stateID == state {
				return true
			}
		}
	}
	return false
}

func NumIndexes(tblname string) (int, error) {
	var indexes int64
	err := DBConn.Raw(fmt.Sprintf(`select count( i.relname) from pg_class t, pg_class i, pg_index ix, pg_attribute a
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
         and t.relkind = 'r'  and t.relname = '%s'`, tblname)).Row().Scan(&indexes)
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(indexes - 1), nil
}

func IsIndex(tblname, column string) (bool, error) {
	row, err := GetOneRow(`select t.relname as table_name, i.relname as index_name, a.attname as column_name
	 from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
		 and t.relkind = 'r'  and t.relname = ?  and a.attname = ?`, tblname, column).String()
	return len(row) > 0 && row[`column_name`] == column, err
}

func GetTableData(tableName string, limit int) ([]map[string]string, error) {
	// TODO fix with EGAAS-240
	if tableName == "dlt_wallets" {
		return GetAll(`SELECT * FROM "`+tableName+`" order by wallet_id`, limit)
	}
	return GetAll(`SELECT * FROM "`+tableName+`" order by id`, limit)
}

func InsertIntoMigration(version string, timeApplied int64) error {
	id, err := GetNextID(`migration_history`)
	if err != nil {
		return err
	}
	return DBConn.Exec(`INSERT INTO migration_history (id, version, date_applied) VALUES (?, ?, ?)`,
		id, version, timeApplied).Error
}

func GetMap(query string, name, value string, args ...interface{}) (map[string]string, error) {
	result := make(map[string]string)
	all, err := GetAll(query, -1, args...)
	if err != nil {
		return result, err
	}
	for _, v := range all {
		result[v[name]] = v[value]
	}
	return result, err
}

// ListResult is a structure for the list result
type ListResult struct {
	result []string
	err    error
}

// String return the slice of strings
func (r *ListResult) String() ([]string, error) {
	if r.err != nil {
		return r.result, r.err
	}
	return r.result, nil
}

// GetList returns the result of the query as ListResult variable
func GetList(query string, args ...interface{}) *ListResult {
	var result []string
	all, err := GetAll(query, -1, args...)
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

func handleError(err error) error {
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	return err
}

func GetNextID(table string) (int64, error) {
	var id int64
	rows, err := DBConn.Raw(`select id from "` + table + `" order by id desc limit 1`).Rows()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting next id from table")
		return 0, err
	}
	rows.Next()
	rows.Scan(&id)
	rows.Close()
	return id + 1, err
}

// Now returns the current time of postgresql
func Now(vars *map[string]string, pars ...string) string {
	var (
		cut             int
		query, interval string
	)
	if len(pars) > 1 && len(pars[1]) > 0 {
		interval = converter.SanitizeNumber(pars[1])
		if interval[0] != '-' && interval[0] != '+' {
			interval = `+` + interval
		}
		interval = fmt.Sprintf(` %s interval '%s'`, interval[:1], strings.TrimSpace(interval[1:]))
	}
	if pars[0] == `` {
		query = `select round(extract(epoch from now()` + interval + `))::integer`
		cut = 10
	} else {
		query = `select now()` + interval
		format := converter.Sanitize(pars[0], `+-: /.`)
		switch format {
		case `datetime`:
			cut = 19
		default:
			if strings.Index(format, `HH`) >= 0 && strings.Index(format, `HH24`) < 0 {
				format = strings.Replace(format, `HH`, `HH24`, -1)
			}
			query = fmt.Sprintf(`select to_char(now()%s, '%s')`, interval, format)
		}
	}
	ret, err := Single(query).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing query to get now value")
		return err.Error()
	}
	if cut > 0 {
		ret = strings.Replace(ret[:cut], `T`, ` `, -1)
	}
	return ret
}

func GetRowVars(vars *map[string]string, pars ...string) string {
	if len(pars) != 4 && len(pars) != 3 {
		return ``
	}
	where := ``
	if len(pars) == 4 {
		where = ` where ` + converter.EscapeName(pars[2]) + `='` + converter.Escape(pars[3]) + `'`
	} else if len(pars) == 3 {
		where = ` where ` + converter.Escape(pars[2])
	}
	value, err := GetOneRow(`select * from ` + converter.EscapeName(pars[1]) + where).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting one row")
		return err.Error()
	}
	for key, val := range value {
		if val == `NULL` {
			val = ``
		}
		(*vars)[pars[0]+`_`+key] = converter.StripTags(val)
	}
	return ``
}

func GetOne(vars *map[string]string, pars ...string) string {
	if len(pars) < 2 {
		return ``
	}
	where := ``
	if len(pars) == 4 {
		where = ` where ` + converter.EscapeName(pars[2]) + `='` + converter.Escape(pars[3]) + `'`
	} else if len(pars) == 3 {
		where = ` where ` + converter.Escape(pars[2])
	}
	value, err := Single(`select ` + converter.Escape(pars[0]) + ` from ` + converter.EscapeName(pars[1]) + where).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting single")
		return err.Error()
	}
	if value == `NULL` {
		value = ``
	}
	return strings.Replace(converter.StripTags(value), "\n", "\n<br>", -1)
}
