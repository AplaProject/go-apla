package model

import "github.com/jinzhu/gorm"

type App struct {
	tableName string
	Name      string `gorm:"primary_key;not null;size:100"`
	Done      int32  `gorm:"not null"`
	Blocks    string `gorm:"not null"`
}

func (a *App) SetTablePrefix(tablePrefix string) {
	a.tableName = tablePrefix + "_apps"
}

func (a *App) TableName() string {
	return a.tableName
}

func (a *App) Get(name string) error {
	return handleError(DBConn.Where("name = ?", name).First(a).Error)
}

func (a *App) GetAll() ([]App, error) {
	var apps []App
	err := DBConn.Table(a.tableName).Find(&apps).Error
	return apps, err
}

func (a *App) IsExists(name string) (bool, error) {
	query := DBConn.Where("name = ?", name).First(a)
	if query.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	return !query.RecordNotFound(), query.Error
}

func (a *App) Save() error {
	return DBConn.Save(a).Error
}

func (a *App) Create() error {
	return DBConn.Create(a).Error
}

func CreateStateAppsTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_apps" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"done" integer NOT NULL DEFAULT '0',
				"blocks" text  NOT NULL DEFAULT ''
				);
				ALTER TABLE ONLY "` + stateID + `_apps" ADD CONSTRAINT "` + stateID + `_apps_pkey" PRIMARY KEY (name);
			`).Error
}
