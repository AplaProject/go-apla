package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/migration"
	"github.com/GenesisKernel/go-genesis/packages/migration/vde"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	// Postgresql driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	// DBConn is orm connection
	DBConn *gorm.DB

	// ErrRecordNotFound is Not Found Record wrapper
	ErrRecordNotFound = gorm.ErrRecordNotFound

	// ErrDBConn database connection error
	ErrDBConn = errors.New("Database connection error")

	FirstEcosystemTables = map[string]bool{
		`keys`:       false,
		`menu`:       true,
		`pages`:      true,
		`blocks`:     true,
		`languages`:  true,
		`contracts`:  true,
		`tables`:     true,
		`parameters`: true,
		`history`:    true,
		`sections`:   true,
		`members`:    false,
		/*	`roles`:              true,
			`roles_participants`: true,
			`notifications`:      true,
			`applications`:       true,
			`binaries`:           true,
			`buffer_data`:        true,
			`app_params`:         true,*/
	}
)

func RealNameEcosystem(tableName string) (name string, ecosystem int64, unique bool) {
	var quote, ok bool
	name = tableName
	if name[0] == '"' {
		name = strings.Trim(name, `"`)
		quote = true
	}
	start := strings.IndexByte(name, '_')
	if start > 0 && start < len(name)-1 {
		if unique, ok = FirstEcosystemTables[name[start+1:]]; ok {
			if ecosystem = converter.StrToInt64(name[:start]); ecosystem > 0 {
				name = `1_` + name[start+1:]
			}
		}
	}
	if quote {
		name = `"` + name + `"`
	}
	return
}

func RealName(tableName string) string {
	tableName, _, _ = RealNameEcosystem(tableName)
	return tableName
}

func isFound(db *gorm.DB) (bool, error) {
	if db.RecordNotFound() {
		return false, nil
	}
	return true, db.Error
}

// GormInit is initializes Gorm connection
func GormInit(host string, port int, user string, pass string, dbName string) error {
	var err error
	DBConn, err = gorm.Open("postgres",
		fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s", host, port, user, dbName, pass))
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
		err := DBConn.Close()
		DBConn = nil
		if err != nil {
			return err
		}
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

// Connection returns connection of database
func (tr *DbTransaction) Connection() *gorm.DB {
	return tr.conn
}

// Savepoint creates PostgreSQL savepoint
func (tr *DbTransaction) Savepoint(idTx int) error {
	return tr.Connection().Exec(fmt.Sprintf("SAVEPOINT \"tx-%d\";", idTx)).Error
}

// RollbackSavepoint rollbacks PostgreSQL savepoint
func (tr *DbTransaction) RollbackSavepoint(idTx int) error {
	return tr.Connection().Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT \"tx-%d\";", idTx)).Error
}

// ReleaseSavepoint releases PostgreSQL savepoint
func (tr *DbTransaction) ReleaseSavepoint(idTx int) error {
	return tr.Connection().Exec(fmt.Sprintf("RELEASE SAVEPOINT \"tx-%d\";", idTx)).Error
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

// GetRecordsCountTx is counting all records of table in transaction
func GetRecordsCountTx(db *DbTransaction, tableName string) (int64, error) {
	var count int64
	realName, ecosysID, _ := RealNameEcosystem(tableName)
	dbQuery := GetDB(db).Table(realName)
	if ecosysID != 0 {
		dbQuery = dbQuery.Where(`ecosystem = ?`, ecosysID)
	}
	err := dbQuery.Count(&count).Error
	return count, err
}

// ExecSchemaEcosystem is executing ecosystem schema
func ExecSchemaEcosystem(db *DbTransaction, id int, wallet int64, name string, founder int64) error {
	if id == 1 {
		q := fmt.Sprintf(migration.GetCommonEcosystemScript())
		if err := GetDB(db).Exec(q).Error; err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing comma ecosystem schema")
			return err
		}
	}
	q := fmt.Sprintf(migration.GetEcosystemScript(), id, wallet, name, founder)
	if err := GetDB(db).Exec(q).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing ecosystem schema")
		return err
	}
	if id == 1 {
		q = fmt.Sprintf(migration.GetFirstEcosystemScript(), wallet)
		if err := GetDB(db).Exec(q).Error; err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing first ecosystem schema")
		}
	}
	return nil
}

// ExecSchemaLocalData is executing schema with local data
func ExecSchemaLocalData(id int, wallet int64) error {
	return DBConn.Exec(fmt.Sprintf(vde.GetVDEScript(), id, wallet)).Error
}

// ExecSchema is executing schema
func ExecSchema() error {
	return migration.Migrate(&MigrationHistory{})
}

// Update is updating table rows
func Update(transaction *DbTransaction, tblname, set, where string) error {
	return GetDB(transaction).Exec(`UPDATE "` + strings.Trim(tblname, `"`) + `" SET ` + set + " " + where).Error
}

// Delete is deleting table rows
func Delete(transaction *DbTransaction, tblname, where string) error {
	return GetDB(transaction).Exec(`DELETE FROM "` + tblname + `" ` + where).Error
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
	return GetDB(transaction).Exec(`ALTER TABLE "` + RealName(tableName) + `" ADD COLUMN "` + columnName + `" ` + columnType).Error
}

// AlterTableDropColumn is dropping column from table
func AlterTableDropColumn(tableName, columnName string) error {
	return DBConn.Exec(`ALTER TABLE "` + RealName(tableName) + `" DROP COLUMN "` + columnName + `"`).Error
}

// CreateIndex is creating index on table column
func CreateIndex(transaction *DbTransaction, indexName, tableName, onColumn string) error {
	return GetDB(transaction).Exec(`CREATE INDEX "` + indexName + `_index" ON "` + RealName(tableName) + `" (` + onColumn + `)`).Error
}

// GetColumnDataTypeCharMaxLength is returns max length of table column
func GetColumnDataTypeCharMaxLength(tableName, columnName string) (map[string]string, error) {
	return GetOneRow(`select data_type,character_maximum_length from
			 information_schema.columns where table_name = ? AND column_name = ?`,
		RealName(tableName), columnName).String()
}

// GetAllColumnTypes returns column types for table
func GetAllColumnTypes(tblname string) ([]map[string]string, error) {
	return GetAll(`SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = ?
		ORDER BY ordinal_position ASC`, -1, RealName(tblname))
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
		case dataType == `jsonb`:
			itype = "json"
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
	rows, err := GetDB(transaction).Raw(`select id from "` + RealName(table) + `" order by id desc limit 1`).Rows()
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

// GetColumnByID returns the value of the column from the table by id
func GetColumnByID(table, column, id string) (result string, err error) {
	err = DBConn.Table(table).Select(column).Where(`id=?`, id).Row().Scan(&result)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting column by id")
	}
	return
}

// InitDB drop all tables and exec db schema
func InitDB(cfg conf.DBConfig) error {

	err := GormInit(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	if err != nil || DBConn == nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("initializing DB")
		return ErrDBConn
	}
	if err = DropTables(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping all tables")
		return err
	}
	if err = ExecSchema(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing db schema")
		return err
	}

	install := &Install{Progress: ProgressComplete}
	if err = install.Create(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating install")
		return err
	}

	if conf.Config.IsSupportingVDE() {
		if err := ExecSchemaLocalData(consts.DefaultVDE, conf.Config.KeyID); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating VDE schema")
			return err
		}
	}

	return nil
}

// DropDatabase kill all process and drop database
func DropDatabase(name string) error {
	query := `SELECT
	pg_terminate_backend (pg_stat_activity.pid)
   FROM
	pg_stat_activity
   WHERE
	pg_stat_activity.datname = ?`

	if err := DBConn.Exec(query, name).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "dbname": name}).Error("on kill db process")
		return err
	}

	if err := DBConn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "dbname": name}).Error("on drop db")
		return err
	}

	return nil
}
