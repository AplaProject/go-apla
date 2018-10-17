package model

import "github.com/GenesisKernel/go-genesis/packages/converter"

// BlockInterface is model
type BlockInterface struct {
	ecosystem  int64
	ID         int64  `gorm:"primary_key;not null" json:"id"`
	Name       string `gorm:"not null" json:"name"`
	Value      string `gorm:"not null" json:"value"`
	Conditions string `gorm:"not null" json:"conditions"`
}

// SetTablePrefix is setting table prefix
func (bi *BlockInterface) SetTablePrefix(prefix string) {
	bi.ecosystem = converter.StrToInt64(prefix)
}

// TableName returns name of table
func (bi BlockInterface) TableName() string {
	if bi.ecosystem == 0 {
		bi.ecosystem = 1
	}
	return `1_blocks`
}

// Get is retrieving model from database
func (bi *BlockInterface) Get(name string) (bool, error) {
	return isFound(DBConn.Where("ecosystem=? and name = ?", bi.ecosystem, name).First(bi))
}
