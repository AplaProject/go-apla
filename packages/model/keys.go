package model

import "fmt"

const keyTableSuffix = "_keys"

// Key is model
type Key struct {
	tableName string
	ID        int64  `gorm:"primary_key;not null"`
	PublicKey []byte `gorm:"column:pub;not null"`
	Amount    string `gorm:"not null"`
	Delete    int64  `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (m *Key) SetTablePrefix(prefix int64) *Key {
	m.tableName = KeyTableName(prefix)
	return m
}

// TableName returns name of table
func (m Key) TableName() string {
	return m.tableName
}

// Get is retrieving model from database
func (m *Key) Get(wallet int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", wallet).First(m))
}

// KeyTableName returns name of keys table
func KeyTableName(prefix int64) string {
	return fmt.Sprintf("%d%s", prefix, keyTableSuffix)
}
