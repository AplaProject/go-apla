package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"os"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"

	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	DBConn *gorm.DB
)

func GormInit(db *sql.DB) error {
	var err error
	DBConn, err = gorm.Open("postgres", db)
	if err != nil {
		return err
	}
	//DBConn.SingularTable(true)
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
	err := DBConn.Table(tableName).Count(count).Error
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
		Scan(count).Error
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
		return 0, err
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

func NumIndexes(tblname string) (int, error) {
	indexes, err := Single(`select count( i.relname) from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
         and t.relkind = 'r'  and t.relname = ?`, tblname).Int64()
	if err != nil {
		return 0, err
	}
	return int(indexes - 1), nil
}

func IsIndex(tblname, column string) (bool, error) {
	indexes, err := GetAll(`select t.relname as table_name, i.relname as index_name, a.attname as column_name 
	 from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
         and t.relkind = 'r'  and t.relname = ?  and a.attname = ?`, 1, tblname, column)
	if err != nil {
		return false, err
	}
	return len(indexes) > 0, nil
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

func GetFuel() decimal.Decimal {
	fuel, _ := Single(`SELECT value FROM system_parameters WHERE name = ?`, "fuel_rate").String()
	cacheFuel, _ := decimal.NewFromString(fuel)
	return cacheFuel
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
	err := DBConn.Raw("SELECT count(*) FROM information_schema.columns WHERE table_name ?", tableName).Scan(&count).Error
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

func GetSleepTime(myWalletID, myStateID, prevBlockStateID, prevBlockWalletID int64) (int64, error) {
	// возьмем список всех full_nodes
	// take the list of all full_nodes
	fullNodesList, err := GetAll("SELECT id, wallet_id, state_id as state_id FROM full_nodes", -1)
	if err != nil {
		return int64(0), err
	}

	// определим full_node_id того, кто должен был генерить блок (но мог это делегировать)
	// determine full_node_id of the one, who had to generate a block (but could delegate this)
	prevBlockFullNodeID, err := Single("SELECT id FROM full_nodes WHERE state_id = ? OR wallet_id = ?", prevBlockStateID, prevBlockWalletID).Int64()
	if err != nil {
		return int64(0), err
	}
	prevBlockFullNodePosition := func(fullNodesList []map[string]string, prevBlockFullNodeID int64) int {
		for i, fullNodes := range fullNodesList {
			if converter.StrToInt64(fullNodes["id"]) == prevBlockFullNodeID {
				return i
			}
		}
		return -1
	}(fullNodesList, prevBlockFullNodeID)

	// определим свое место (в том числе в delegate)
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

func AlterTableAddColumn(tableName, columnName, columnType string) error {
	return DBConn.Exec(`ALTER TABLE "` + tableName + `" ADD COLUMN ` + columnName + ` ` + columnType).Error
}

func AlterTableDropColumn(tableName, columnName string) error {
	return DBConn.Exec(`ALTER TABLE "` + tableName + `" DROP COLUMN ` + columnName).Error
}

func CreateIndex(indexName, tableName, onColumn string) error {
	return DBConn.Exec(`CREATE INDEX "` + indexName + `_index" ON "` + tableName + `" (` + onColumn + `)`).Error
}

func GormSet(db *gorm.DB) {
	DBConn = db
	DBConn.SingularTable(true)
}

// TODO: should be atomic ?
func GetCurrentDB() *gorm.DB {
	return DBConn
}
