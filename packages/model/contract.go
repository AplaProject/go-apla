package model

// Signature is model
type Contract struct {
	tableName    string
	ID           int64  `gorm:"primary_key;not null;default:0"`
	Name         string `gorm:"unique;not null;default:''"`
	Value        string `gorm:"unique;not null;default:''"`
	WalletID     int64  `gorm:"not null;default:0"`
	TokenID      int64  `gorm:"not null;default:1"`
	Active       string `gorm:"not null;default 0;size:1"`
	Confirmation string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions   string `gorm:"not null;default:''"`
	AppID        int64  `gorm:"not null;default:1"`
}

// SetTablePrefix is setting table prefix
func (c *Contract) SetTablePrefix(prefix string) {
	c.tableName = prefix + "_contracts"
}

// TableName returns name of table
func (c *Contract) TableName() string {
	return c.tableName
}

// Get is retrieving model from database
func (c *Contract) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(c))
}
