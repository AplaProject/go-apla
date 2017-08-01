package model

import (
	"fmt"
	"os"

	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	DBConn *gorm.DB
)

func GormInit(user string, pass string, host string, dbName string) error {
	var err error
	DBConn, err = gorm.Open("postgres", fmt.Sprintf(
		"user=%s password=pass host=localhost dbname=%s sslmode=disable "), user, pass, dbName)
	if err != nil {
		return err
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
		Select("table_name").Scan(result).Error
	return result, err
}
