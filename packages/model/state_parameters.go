package model

type StateParameter struct {
	tableName  string
	ID         int64  `gorm:"primary_key;not null"`
	Name       string `gorm:"not null;size:100"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (sp *StateParameter) TableName() string {
	return sp.tableName
}

func (sp *StateParameter) SetTablePrefix(tablePrefix string) {
	sp.tableName = tablePrefix + "_parameters"
}

func (sp *StateParameter) Get(transaction *DbTransaction, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("name = ?", name).First(sp))
}

func (sp *StateParameter) GetAllStateParameters() ([]StateParameter, error) {
	parameters := make([]StateParameter, 0)
	err := DBConn.Table(sp.TableName()).Find(&parameters).Error
	if err != nil {
		return nil, err
	}
	return parameters, nil
}
