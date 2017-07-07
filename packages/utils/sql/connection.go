package sql

import (
	"fmt"
	"sync"

	"database/sql"

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
	ConfigIni map[string]string
	//GoroutineName string
}

// NewDbConnect creates a new database connection
func NewDbConnect(ConfigIni map[string]string) (*DCDB, error) {
	var db *sql.DB
	var err error
	if len(ConfigIni["db_user"]) == 0 || len(ConfigIni["db_password"]) == 0 || len(ConfigIni["db_name"]) == 0 {
		return &DCDB{}, err
	}
	db, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable port=%s", ConfigIni["db_user"], ConfigIni["db_password"], ConfigIni["db_name"], ConfigIni["db_port"]))
	if err != nil || db.Ping() != nil {
		return &DCDB{}, err
	}
	log.Debug("return")
	return &DCDB{db, ConfigIni}, err
}

func GetCurrentDB() *DCDB {
	// TODO:  should be atomic.Pointer ??
	return DB
}
