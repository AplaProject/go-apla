package model

import (
	"strconv"
)

type Menu struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (m *Menu) SetTablePrefix(prefix string) {
	m.tableName = prefix + "_menu"
}

func (m Menu) TableName() string {
	return m.tableName
}

func (m *Menu) Get(name string) error {
	return DBConn.Where("name = ?", name).First(m).Error
}

func (m *Menu) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(m).Error
}

func (m *Menu) GetAll(prefix string) ([]Menu, error) {
	result := make([]Menu, 0)
	err := DBConn.Table(prefix + "_menu").Order("name").Find(&result).Error
	return result, err
}

func (m *Menu) ToMap() map[string]string {
	result := make(map[string]string)
	result["name"] = m.Name
	result["value"] = m.Value
	result["conditions"] = m.Conditions
	result["rb_id"] = strconv.FormatInt(m.RbID, 10)
	return result
}

func CreateStateMenuTable(transaction *DbTransaction, stateID string) error {
	return GetDB(transaction).Exec(`CREATE TABLE "` + stateID + `_menu" (
				"name" varchar(255)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"conditions" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + stateID + `_menu" ADD CONSTRAINT "` + stateID + `_menu_pkey" PRIMARY KEY (name);
				`).Error
}
