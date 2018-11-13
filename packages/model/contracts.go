package model

// Contract is model
type Contract struct {
	ID        int64  `gorm:"primary_key;not null"`
	Value     string `gorm:"not null"`
	TokenID   int64  `gorm:"not null"`
	WalletID  int64  `gorm:"not null"`
	Ecosystem int64  `gorm:"not null"`
}

// TableName returns name of table
func (c *Contract) TableName() string {
	return `1_contracts`
}

// GetList is retrieving records from database
func (c *Contract) GetList(offset, limit int64) ([]Contract, error) {
	result := new([]Contract)
	err := DBConn.Table(c.TableName()).Offset(offset).Limit(limit).Order("id asc").Find(&result).Error
	return *result, err
}

// GetFromEcosystem retrieving ecosystem contracts from database
func (c *Contract) GetFromEcosystem(db *DbTransaction, ecosystem int64) ([]Contract, error) {
	result := new([]Contract)
	err := GetDB(db).Table(c.TableName()).Where("ecosystem = ?", ecosystem).Order("id asc").Find(&result).Error
	return *result, err
}

// Count returns count of records in table
func (c *Contract) Count() (count int64, err error) {
	err = DBConn.Table(c.TableName()).Count(&count).Error
	return
}
