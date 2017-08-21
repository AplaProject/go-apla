package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	logging "github.com/op/go-logging"

	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	DBConn         *gorm.DB
	log            = logging.MustGetLogger("model")
	RecordNotFound = gorm.ErrRecordNotFound
)

func GormInit(user string, pass string, dbName string) error {
	var err error
	DBConn, err = gorm.Open("postgres",
		fmt.Sprintf("host=localhost user=%s dbname=%s sslmode=disable password=%s", user, dbName, pass))
	if err != nil {
		return err
	}

	//DBConn.LogMode(true)
	return nil
}

func GormClose() error {
	if DBConn != nil {
		return DBConn.Close()
	}
	return nil
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

func ExecSchema() error {
	schema, err := static.Asset("static/schema.sql")
	if err != nil {
		os.Remove(*utils.Dir + "/config.ini")
		return err
	}
	return DBConn.Exec(string(schema)).Error
}

func GetColumnsCount(tableName string) (int64, error) {
	var count int64
	err := DBConn.Table("information_schema.columns").
		Where("table_name=?", tableName).
		Select("count(column_name)").
		Scan(&count).Error
	return count, err
}

func GetTables() ([]string, error) {
	var result []string
	err := DBConn.Table("information_schema.tables").
		Where("table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')").
		Select("table_name").Scan(&result).Error
	return result, err
}

func Update(tblname, set, where string) error {
	return DBConn.Exec("UPDATE " + tblname + " SET " + set + " " + where).Error
}

func Delete(tblname, where string) error {
	return DBConn.Exec("DELETE FROM " + tblname + " " + where).Error
}

func InsertReturningLastID(table, columns, values string) (int64, error) {
	var result int64
	returning, err := GetFirstColumnName(table)
	if err != nil {
		return 0, err
	}
	insertQuery := `INSERT INTO "` + table + `" (` + columns + `) VALUES (` + values + `) RETURNING ` + returning
	err = DBConn.Raw(insertQuery).Row().Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func SequenceRestartWith(seqName string, id int64) error {
	return DBConn.Exec("ALTER SEQUENCE " + seqName + " RESTART WITH " + converter.Int64ToStr(id)).Error
}

func GetSerialSequence(table, AiID string) (string, error) {
	var result string
	query := `SELECT pg_get_serial_sequence('` + table + `', '` + AiID + `')`
	err := DBConn.Raw(query).Row().Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func GetCurrentSeqID(id, tblname string) (int64, error) {
	var result int64
	query := "SELECT " + id + " FROM " + tblname + " ORDER BY " + id + " DESC LIMIT 1"
	err := DBConn.Raw(query).Row().Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func GetRollbackID(tblname, where, ordering string) (int64, error) {
	var result int64
	query := "SELECT rb_id FROM " + tblname + " " + where + " order by rb_id " + ordering
	err := DBConn.Raw(query).Row().Scan(&result)
	if err != nil {
		log.Errorf("can't get rollback_id: %s for query %s", err, query)
		// TODO
		return 0, nil
	}
	return result, nil
}

func GetFirstColumnName(table string) (string, error) {
	rows, err := DBConn.Raw(`SELECT * FROM "` + table + `" LIMIT 1`).Rows()
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

func GetQueryTotalCost(query string, args ...interface{}) (int64, error) {
	var planStr string
	err := DBConn.Raw(fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", query), args...).Row().Scan(&planStr)
	switch {
	case err == sql.ErrNoRows:
		return 0, errors.New("No rows")
	case err != nil:
		return 0, err
	}
	var queryPlan []map[string]interface{}
	dec := json.NewDecoder(strings.NewReader(planStr))
	dec.UseNumber()
	if err := dec.Decode(&queryPlan); err != nil {
		return 0, err
	}
	if len(queryPlan) == 0 {
		return 0, errors.New("Query plan is empty")
	}
	firstNode := queryPlan[0]
	var plan interface{}
	var ok bool
	if plan, ok = firstNode["Plan"]; !ok {
		return 0, errors.New("No Plan key in result")
	}
	var planMap map[string]interface{}
	if planMap, ok = plan.(map[string]interface{}); !ok {
		return 0, errors.New("Plan is not map[string]interface{}")
	}
	if totalCost, ok := planMap["Total Cost"]; ok {
		if totalCostNum, ok := totalCost.(json.Number); ok {
			if totalCostF64, err := totalCostNum.Float64(); err != nil {
				return 0, err
			} else {
				return int64(totalCostF64), nil
			}
		} else {
			return 0, errors.New("Total cost is not a number")
		}
	} else {
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

func GetAiID(table string) (string, error) {
	exists := ""
	column := "id"
	if table == "users" {
		column = "user_id"
	} else if table == "miners" {
		column = "miner_id"
	} else {
		exists = ""
		err := DBConn.Raw("SELECT column_name FROM information_schema.columns WHERE table_name=? and column_name=?", table, "id").Row().Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return "", err
		}
		if len(exists) == 0 {
			err := DBConn.Raw("SELECT column_name FROM information_schema.columns WHERE table_name=? and column_name=?", table, "rb_id").Row().Scan(&exists)
			if err != nil {
				return "", err
			}
			if len(exists) == 0 {
				return "", fmt.Errorf("no id, rb_id")
			}
			column = "rb_id"
		}
	}
	return column, nil
}

func SetAI(table string, AI int64) error {
	AiID, err := GetAiID(table)
	if err != nil {
		return err
	}
	pgGetSerialSequence, err := GetSerialSequence(table, AiID)
	if err != nil {
		return err
	}
	err = SequenceRestartWith(pgGetSerialSequence, AI)
	if err != nil {
		return err
	}
	return nil
}

func SendTx(txType int64, adminWallet int64, data []byte) (hash []byte, err error) {
	hash, err = crypto.Hash(data)
	if err != nil {
		return nil, err
	}

	ts := &TransactionStatus{
		Hash:      hash,
		Time:      time.Now().Unix(),
		Type:      txType,
		WalletID:  adminWallet,
		CitizenID: adminWallet}
	err = ts.Create()
	if err != nil {
		return nil, err
	}
	qtx := &QueueTx{Hash: hash,
		Data: data}
	err = qtx.Create()
	return
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

func AlterTableAddColumn(tableName, columnName, columnType string) error {
	return DBConn.Exec(`ALTER TABLE "` + tableName + `" ADD COLUMN ` + columnName + ` ` + columnType).Error
}

func AlterTableDropColumn(tableName, columnName string) error {
	return DBConn.Exec(`ALTER TABLE '` + tableName + `' DROP COLUMN ` + columnName).Error
}

func CreateIndex(indexName, tableName, onColumn string) error {
	return DBConn.Exec(`CREATE INDEX "` + indexName + `_index" ON "` + tableName + `" (` + onColumn + `)`).Error
}

func IsTable(tblname string) bool {
	var name string
	DBConn.Table("information_schema.tables").
		Where("table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema') AND table_name=?`", tblname).
		Select("table_name").Row().Scan(&name)

	return name == tblname
}

func GetColumnDataTypeCharMaxLength(tableName, columnName string) (map[string]string, error) {
	type Proxy struct {
		DataType               string
		CharacterMaximumLength string
	}
	var temp Proxy
	err := DBConn.
		Table("information_schema.columns").
		Where("table_name = ? AND column_name = ?, tableName, columnName").
		Select("data_type", "character_maximum_length").
		Find(&temp).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, 0)
	result["data_type"] = temp.DataType
	result["character_maximum_length"] = temp.CharacterMaximumLength
	return result, nil
}

func GetColumnType(tblname, column string) (itype string) {
	coltype, _ := GetColumnDataTypeCharMaxLength(tblname, column)
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
			if converter.StrToInt64(fullNodes["id"]) == prevBlockFullNodeID {
				return i
			}
		}
		return -1
	}(fullNodesList, int64(prevBlockFullNodeID))

	// define our place (Including in the 'delegate')
	myPosition := func(fullNodesList []map[string]string, myWalletID, myStateID int64) int {
		for i, fullNodes := range fullNodesList {
			if converter.StrToInt64(fullNodes["state_id"]) == myStateID || converter.StrToInt64(fullNodes["wallet_id"]) == myWalletID ||
				converter.StrToInt64(fullNodes["final_delegate_state_id"]) == myWalletID || converter.StrToInt64(fullNodes["final_delegate_wallet_id"]) == myWalletID {
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

func DropTable(tableName string) error {
	return DBConn.DropTable(tableName).Error
}

func GetConditionsAndValue(tableName, name string) (map[string]string, error) {
	type proxy struct {
		Conditions string
		Value      string
	}
	var temp proxy
	err := DBConn.Table(tableName).Where("name = ?", name).Select("conditions", "value").Find(&temp).Error
	if err != nil {
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
			if converter.StrToInt64(id) == state {
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
	query := DBConn.Raw(fmt.Sprintf(`select t.relname as table_name, i.relname as index_name, a.attname as column_name
	 from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
		 and t.relkind = 'r'  and t.relname = '%s'  and a.attname = '%s'`, tblname, column))
	return query.RowsAffected > 0, query.Error
}

func GetTableData(tableName string, limit int) ([]map[string]string, error) {
	return GetAll(`SELECT * FROM "`+tableName+`" order by id`, limit)
}

func InsertIntoMigration(version string, timeApplied int64) error {
	return DBConn.Exec(`INSERT INTO migration_history (version, date_applied) VALUES (?, ?)`, version, timeApplied).Error
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
