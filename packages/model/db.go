package model

import (
	"database/sql"

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
	DBConn.SingularTable(true)
	return nil
}

func GormSet(db *gorm.DB) {
	DBConn = db
	DBConn.SingularTable(true)
}

// TODO: should be atomic ?
func GetCurrentDB() *gorm.DB {
	return DBConn
}
