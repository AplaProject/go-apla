package sql

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/EGaaS/go-egaas-mvp/packages/config"

	_ "github.com/lib/pq"
	"github.com/op/go-logging"
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
}

// NewDbConnect creates a new database connection
func NewDbConnect() (*DCDB, error) {
	var db *sql.DB
	var err error
	if len(config.ConfigIni["db_user"]) == 0 || len(config.ConfigIni["db_password"]) == 0 || len(config.ConfigIni["db_name"]) == 0 {
		return &DCDB{}, err
	}
	db, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=%s", config.ConfigIni["db_user"], config.ConfigIni["db_password"], config.ConfigIni["db_name"], config.ConfigIni["db_port"]))
	if err != nil || db.Ping() != nil {
		return &DCDB{}, err
	}
	log.Debug("return")
	return &DCDB{db}, err
}
