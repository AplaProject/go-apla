package model

// AppParam is model
type AppParam struct {
	tableName  string
	ID         int64  `gorm:"primary_key;not null"`
	AppID      int64  `gorm:"not null"`
	Name       string `gorm:"not null;size:100"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
}

// TableName returns name of table
func (sp *AppParam) TableName() string {
	return sp.tableName
}

// SetTablePrefix is setting table prefix
func (sp *AppParam) SetTablePrefix(tablePrefix string) {
	sp.tableName = tablePrefix + "_app_params"
}

// Get is retrieving model from database
func (sp *AppParam) Get(transaction *DbTransaction, app int64, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("app_id=? and name = ?", app, name).First(sp))
}

// GetAllAppParameters is returning all state parameters
func (sp *AppParam) GetAllAppParameters(app int64) ([]AppParam, error) {
	parameters := make([]AppParam, 0)
	err := DBConn.Table(sp.TableName()).Find(&parameters).Error
	if err != nil {
		return nil, err
	}
	return parameters, nil
}
