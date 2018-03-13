package model

// BlockInterface is model
type BlockInterface struct {
	tableName  string
	ID         int64  `gorm:"primary_key;not null" json:"id"`
	Name       string `gorm:"not null" json:"name"`
	Value      string `gorm:"not null" json:"value"`
	Conditions string `gorm:"not null" json:"conditions"`
}

// SetTablePrefix is setting table prefix
func (bi *BlockInterface) SetTablePrefix(prefix string) {
	bi.tableName = prefix + "_blocks"
}

// TableName returns name of table
func (bi BlockInterface) TableName() string {
	return bi.tableName
}

// Get is retrieving model from database
func (bi *BlockInterface) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(bi))
}
