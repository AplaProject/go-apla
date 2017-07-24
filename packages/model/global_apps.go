package model

type Apps struct {
	tableName string
	Name      string `gorm:"private_key;not null;size:100"`
	Done      int32  `gorm:"not null"`
	Blocks    string `gorm:"not null"`
}

func (a *Apps) SetTableName(tablePrefix string) {
	a.tableName = tablePrefix + "_apps"
}

func (a *Apps) TableName() string {
	return a.tableName
}

func (a *Apps) Get(name string) error {
	return DBConn.Where("name = ?", name).First(a).Error
}

func (a *Apps) GetAll() ([]Apps, error) {
	var apps []Apps
	err := DBConn.Table(a.tableName).Find(apps).Error
	return apps, err
}

func (a *Apps) IsExists(name string) (bool, error) {
	query := DBConn.Where("name = ?", name).First(a)
	return !query.RecordNotFound(), query.Error
}

func (a *Apps) Save() error {
	return DBConn.Save(a).Error
}

func (a *Apps) Create() error {
	return DBConn.Create(a).Error
}
