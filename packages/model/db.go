package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	logging "github.com/op/go-logging"

	"github.com/AplaProject/go-apla/packages/static"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	DBConn         *gorm.DB
	log            = logging.MustGetLogger("model")
	RecordNotFound = gorm.ErrRecordNotFound
)

func isFound(db *gorm.DB) (bool, error) {
	if db.RecordNotFound() {
		return false, nil
	}
	return true, db.Error
}

func GormInit(user string, pass string, dbName string) error {
	var err error
	DBConn, err = gorm.Open("postgres",
		fmt.Sprintf("host=localhost user=%s dbname=%s sslmode=disable password=%s", user, dbName, pass))
	if err != nil {
		DBConn = nil
		return err
	}
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
		return err
	}
	err = DBConn.Exec(fmt.Sprintf(string(schema), id, wallet, name)).Error
	if err != nil {
		return err
	}
	if id == 1 {
		schema, err = static.Asset("static/schema-firstecosystem-v2.sql")
		if err != nil {
			return err
		}
		err = DBConn.Exec(fmt.Sprintf(string(schema), wallet)).Error
	}
	return err
}

func ExecSchema() error {
	schema, err := static.Asset("static/schema-v2.sql")
	if err != nil {
		os.Remove(*utils.Dir + "/config.ini")
		return err
	}
	return DBConn.Exec(string(schema)).Error
}

func Update(transaction *DbTransaction, tblname, set, where string) error {
	return GetDB(transaction).Exec(`UPDATE "` + strings.Trim(tblname, `"`) + `" SET ` + set + " " + where).Error
}

func Delete(tblname, where string) error {
	return DBConn.Exec(`DELETE FROM "` + tblname + `" ` + where).Error
}

func GetRollbackID(tblname, where, ordering string) (int64, error) {
	var result int64
	query := `SELECT rb_id FROM "` + tblname + `" ` + where + " order by rb_id " + ordering
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

	planMap, ok := plan.(map[string]interface{})
	if !ok {
		return 0, errors.New("Plan is not map[string]interface{}")
	}

	totalCost, ok := planMap["Total Cost"]
	if !ok {
		return 0, errors.New("PlanMap has no TotalCost")
	}

	totalCostNum, ok := totalCost.(json.Number)
	if !ok {
		return 0, errors.New("Total cost is not a number")
	}

	totalCostF64, err := totalCostNum.Float64()
	if err != nil {
		return 0, err
	}
	return int64(totalCostF64), nil
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

func AlterTableAddColumn(transaction *DbTransaction, tableName, columnName, columnType string) error {
	return GetDB(transaction).Exec(`ALTER TABLE "` + tableName + `" ADD COLUMN ` + columnName + ` ` + columnType).Error
}

func AlterTableDropColumn(tableName, columnName string) error {
	return DBConn.Exec(`ALTER TABLE "` + tableName + `" DROP COLUMN ` + columnName).Error
}

func CreateIndex(transaction *DbTransaction, indexName, tableName, onColumn string) error {
	return GetDB(transaction).Exec(`CREATE INDEX "` + indexName + `_index" ON "` + tableName + `" (` + onColumn + `)`).Error
}

func GetColumnDataTypeCharMaxLength(tableName, columnName string) (map[string]string, error) {
	return GetOneRow(`select data_type,character_maximum_length from
			 information_schema.columns where table_name = ? AND column_name = ?`,
		tableName, columnName).String()
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



func DropTable(transaction *DbTransaction, tableName string) error {
	return GetDB(transaction).DropTable(tableName).Error
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
	row, err := GetOneRow(`select t.relname as table_name, i.relname as index_name, a.attname as column_name
	 from pg_class t, pg_class i, pg_index ix, pg_attribute a 
	 where t.oid = ix.indrelid and i.oid = ix.indexrelid and a.attrelid = t.oid and a.attnum = ANY(ix.indkey)
		 and t.relkind = 'r'  and t.relname = ?  and a.attname = ?`, tblname, column).String()
	return len(row) > 0 && row[`column_name`] == column, err
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

func GetNextID(transaction *DbTransaction, table string) (int64, error) {
	var id int64
	rows, err := GetDB(transaction).Raw(`select id from "` + table + `" order by id desc limit 1`).Rows()
	if err != nil {
		return 0, err
	}
	rows.Next()
	rows.Scan(&id)
	rows.Close()
	return id + 1, err
}
