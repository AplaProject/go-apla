package model

import (
	"fmt"
)

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
	if prefix == 0 {
		prefix = 1
	}
	m.tableName = fmt.Sprintf("%d_keys", prefix)
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
