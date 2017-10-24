package model

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

func (a *App) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(a))
}

func (a *App) GetAll() ([]App, error) {
	apps := make([]App, 0)
	err := DBConn.Table(a.tableName).Find(&apps).Error
	return apps, err
}

func (a *App) Save() error {
	return DBConn.Save(a).Error
}

func (a *App) Create() error {
	return DBConn.Create(a).Error
}

func CreateStateAppsTable(transaction *DbTransaction, stateID string) error {
	return GetDB(transaction).Exec(`CREATE TABLE "` + stateID + `_apps" (
				"name" varchar(100)  NOT NULL DEFAULT '',
				"done" integer NOT NULL DEFAULT '0',
				"blocks" text  NOT NULL DEFAULT ''
				);
				ALTER TABLE ONLY "` + stateID + `_apps" ADD CONSTRAINT "` + stateID + `_apps_pkey" PRIMARY KEY (name);
			`).Error
}
