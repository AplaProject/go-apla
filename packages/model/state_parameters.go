package model

import (
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
)

type StateParameters struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:100"`
	Value      string `gorm:"not null"`
	ByteCode   []byte `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func NewStateParameters(tablePrefix string) *StateParameters {
	return &StateParameters{tableName: tablePrefix + "_state_parameters"}
}

func (sp *StateParameters) GetMoneyDigit() int {
	DBConn.Where("name = ?", "money_digit").First(&sp)
	return converter.StrToInt(sp.Value)
}

func (sp *StateParameters) GetCurrency() string {
	DBConn.Where("name = ?", "currency_name").First(&sp)
	return sp.Value
}
