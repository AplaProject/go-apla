package model

import (
	"fmt"

	"github.com/AplaProject/go-apla/packages/conf"

	"github.com/AplaProject/go-apla/packages/consts"

	log "github.com/sirupsen/logrus"
)

// Install is model
type Install struct {
	Progress string `gorm:"not null;size:10"`
}

// TableName returns name of table
func (i *Install) TableName() string {
	return "install"
}

// Get is retrieving model from database
func (i *Install) Get() error {
	return DBConn.Find(i).Error
}

// Create is creating record of model
func (i *Install) Create() error {
	return DBConn.Create(i).Error
}

// InitDB drop all tables and exec db schema
func InitDB(cfg conf.DBConfig) error {

	err := GormInit(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
	if err != nil || DBConn == nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("initializing DB")
		err = fmt.Errorf(`E_DBNIL`)
		return err
	}
	if err = DropTables(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("dropping all tables")
		return err
	}
	if err = ExecSchema(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing db schema")
		return err
	}

	install := &Install{Progress: "complete"}
	if err = install.Create(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating install")
		return err
	}

	return nil
}
