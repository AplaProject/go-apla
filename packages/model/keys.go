package model

import "fmt"

// Key is model
type Key struct {
	ecosystem int64
	ID        int64  `gorm:"primary_key;not null"`
	PublicKey []byte `gorm:"column:pub;not null"`
	Amount    string `gorm:"not null"`
	Maxpay    string `gorm:"not null"`
	Deleted   int64  `gorm:"not null"`
	Blocked   int64  `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (m *Key) SetTablePrefix(prefix int64) *Key {
	m.ecosystem = prefix
	return m
}

// TableName returns name of table
func (m Key) TableName() string {
	if m.ecosystem == 0 {
		m.ecosystem = 1
	}
	return `1_keys`
}

// Get is retrieving model from database
func (m *Key) Get(wallet int64) (bool, error) {
	return isFound(DBConn.Where("id = ? and ecosystem = ?", wallet, m.ecosystem).First(m))
}

// KeyTableName returns name of key table
func KeyTableName(prefix int64) string {
	return fmt.Sprintf("%d_keys", prefix)
}
