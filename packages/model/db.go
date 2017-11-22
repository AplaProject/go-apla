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
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"

	"github.com/AplaProject/go-apla/packages/static"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

var (
	// DBConn is orm connection
	DBConn *gorm.DB

	// ErrRecordNotFound is Not Found Record wrapper
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

func isFound(db *gorm.DB) (bool, error) {
	if db.RecordNotFound() {
		return false, nil
	}
	return true, db.Error
}

// GormInit is initializes Gorm connection
func GormInit(host string, port string, user string, pass string, dbName string) error {
	var err error
	DBConn, err = gorm.Open("postgres",
		fmt.Sprintf("host=localhost user=%s dbname=%s sslmode=disable password=%s", user, dbName, pass))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("cant open connection to DB")
		DBConn = nil
		return err
	}
	return nil
}

// GormClose is closing Gorm connection
func GormClose() error {
	if DBConn != nil {
		return DBConn.Close()
	}
	return nil
}

// DbTransaction is gorm.DB wrapper
type DbTransaction struct {
	conn *gorm.DB
}

// StartTransaction is beginning transaction
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

// Rollback is transaction rollback
func (tr *DbTransaction) Rollback() {
	tr.conn.Rollback()
}

// Commit is transaction commit
func (tr *DbTransaction) Commit() error {
	return tr.conn.Commit().Error
}

// GetDB is returning gorm.DB
func GetDB(tr *DbTransaction) *gorm.DB {
	if tr != nil && tr.conn != nil {
		return tr.conn
	}
	return DBConn
}

// DropTables is dropping all of the tables
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

// GetRecordsCount is counting all records of table
func GetRecordsCount(tableName string) (int64, error) {
	var count int64
	err := DBConn.Table(tableName).Count(&count).Error
	return count, err
}

// ExecSchemaEcosystem is executing ecosystem schema
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

// ExecSchemaLocalData is executing schema with local data
func ExecSchemaLocalData(id int, wallet int64) error {
	schema, err := static.Asset("static/schema-vde.sql")
	if err != nil {
		return err
	}
	return DBConn.Exec(fmt.Sprintf(string(schema), id, wallet)).Error
}

// ExecSchema is executing schema
func ExecSchema() error {
	schema, err := static.Asset("static/schema-v2.sql")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting static/schema-v2.sql asset")
		os.Remove(*utils.Dir + "/config.ini")
		return err
	}
	return DBConn.Exec(string(schema)).Error
}

// Update is updating table rows
func Update(transaction *DbTransaction, tblname, set, where string) error {
	return GetDB(transaction).Exec(`UPDATE "` + strings.Trim(tblname, `"`) + `" SET ` + set + " " + where).Error
}

// Delete is deleting table rows
func Delete(tblname, where string) error {
	return DBConn.Exec(`DELETE FROM "` + tblname + `" ` + where).Error
}

// GetFirstColumnName is returning name of first column
func GetFirstColumnName(table string) (string, error) {
	rows, err := DBConn.Raw(`SELECT * FROM "` + table + `" LIMIT 1`).Rows()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting rollback id from table")
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rows columns")
		return "", err
	}
	if len(columns) > 0 {
		return columns[0], nil
	}
	return "", nil
}

// GetQueryTotalCost is counting query execution time
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

	planMap, ok := plan.(map[string]interface{})
	if !ok {
		log.Error("Plan is not map[string]interface{}")
		return 0, errors.New("Plan is not map[string]interface{}")
	}

	totalCost, ok := planMap["Total Cost"]
	if !ok {
		return 0, errors.New("PlanMap has no TotalCost")
	}

	totalCostNum, ok := totalCost.(json.Number)
	if !ok {
		log.Error("PlanMap has no TotalCost")
		return 0, errors.New("Total cost is not a number")
	}

	totalCostF64, err := totalCostNum.Float64()
	if err != nil {
		log.Error("Total cost is not a number")
		return 0, err
	}
	return int64(totalCostF64), nil
}

// GetAllTables returning a slice of table names
func GetAllTables() ([]string, error) {
	var result []string
	sql := `SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema')`
	rows, err := DBConn.Raw(sql).Rows()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": sql}).Error("executing raw query")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tblname string
		if err := rows.Scan(&tblname); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("scanning table name from row")
			return nil, err
		}
		result = append(result, tblname)
	}
	if err := rows.Err(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("retrieving rows from table")
		return nil, err
	}
	return result, nil
}

// GetColumnCount is counting rows in table
func GetColumnCount(tableName string) (int64, error) {
	var count int64
	err := DBConn.Raw("SELECT count(*) FROM information_schema.columns WHERE table_name=?", tableName).Row().Scan(&count)
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing raw query")
		return 0, err
	}
	return count, nil
}

// SendTx is creates transaction
func SendTx(txType int64, adminWallet int64, data []byte) ([]byte, error) {
	hash, err := crypto.Hash(data)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("hashing data")
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
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("transaction status create")
		return nil, err
	}
	qtx := &QueueTx{
		Hash: hash,
		Data: data,
	}
	err = qtx.Create()
	return hash, err
}

// AlterTableAddColumn is adding column to table
func AlterTableAddColumn(transaction *DbTransaction, tableName, columnName, columnType string) error {
	return GetDB(transaction).Exec(`ALTER TABLE "` + tableName + `" ADD COLUMN ` + columnName + ` ` + columnType).Error
}

// AlterTableDropColumn is dropping column from table
func AlterTableDropColumn(tableName, columnName string) error {
	return DBConn.Exec(`ALTER TABLE "` + tableName + `" DROP COLUMN ` + columnName).Error
}

// CreateIndex is creating index on table column
func CreateIndex(transaction *DbTransaction, indexName, tableName, onColumn string) error {
	return GetDB(transaction).Exec(`CREATE INDEX "` + indexName + `_index" ON "` + tableName + `" (` + onColumn + `)`).Error
}

// GetColumnDataTypeCharMaxLength is returns max length of table column
func GetColumnDataTypeCharMaxLength(tableName, columnName string) (map[string]string, error) {
	return GetOneRow(`select data_type,character_maximum_length from
			 information_schema.columns where table_name = ? AND column_name = ?`,
		tableName, columnName).String()
}

// GetColumnType is returns type of column
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

// DropTable is dropping table
func DropTable(transaction *DbTransaction, tableName string) error {
	return GetDB(transaction).DropTable(tableName).Error
}

// IsNodeState :Because of import cycle utils and config
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

// NumIndexes is counting table indexes
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

// IsIndex returns is table column is an index
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

// GetNextID returns next ID of table
func GetNextID(transaction *DbTransaction, table string) (int64, error) {
	var id int64
	rows, err := GetDB(transaction).Raw(`select id from "` + table + `" order by id desc limit 1`).Rows()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting next id from table")
		return 0, err
	}
	rows.Next()
	rows.Scan(&id)
	rows.Close()
	return id + 1, err
}

// IsTable returns is table exists
func IsTable(tblname string) bool {
	var name string
	DBConn.Table("information_schema.tables").
		Where("table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema') AND table_name=?", tblname).
		Select("table_name").Row().Scan(&name)

	return name == tblname
}

// GetRollbackID returns rollback id
func GetRollbackID(transaction *DbTransaction, tblname, where, ordering string) (int64, error) {
	var result int64
	err := GetDB(transaction).Raw(`SELECT rb_id FROM "` + tblname + `" ` + where + " order by rb_id " + ordering).Row().Scan(&result)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error(fmt.Errorf("GetRollbackID from table %s where %s order by rb_id %s", tblname, where, ordering))
		return 0, err
	}
	return result, nil
}
