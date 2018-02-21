package model

import "fmt"

// Key is model
type Key struct {
	prefix    int64
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
	m.prefix = prefix
	return m
}

// TableName returns name of table
func (m Key) TableName() string {
	if m.prefix == 0 {
		m.prefix = 1
	}
	return fmt.Sprintf("%d_keys", m.prefix)
}

// Get is retrieving model from database
func (m *Key) Get(wallet int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", wallet).First(m))
}
