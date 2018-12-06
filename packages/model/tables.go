package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/jinzhu/gorm"
)

// const TableName = "1_tables"

// Table is model
type Table struct {
	ID          int64       `gorm:"primary_key;not null"`
	Name        string      `gorm:"not null;size:100"`
	Permissions Permissions `gorm:"not null;type:jsonb(PostgreSQL)"`
	Columns     string      `gorm:"not null"`
	Conditions  string      `gorm:"not null"`
	AppID       int64       `gorm:"not null"`
	Ecosystem   int64       `gorm:"not null"`
}

type Permissions struct {
	Insert    string `json:"insert"`
	NewColumn string `json:"new_column"`
	Update    string `json:"update"`
	Read      string `json:"read"`
	Filter    string `json:"filter"`
}

func (p Permissions) Value() (driver.Value, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return string(data), err
}
func (p *Permissions) Scan(v interface{}) error {
	data, ok := v.([]byte)
	if !ok {
		return errors.New("Bad permissions")
	}
	return json.Unmarshal(data, p)
}

// SetTablePrefix is setting table prefix
func (t *Table) SetTablePrefix(prefix string) {
	t.Ecosystem = converter.StrToInt64(prefix)
}

// TableName returns name of table
func (t *Table) TableName() string {
	if t.Ecosystem == 0 {
		t.Ecosystem = 1
	}
	return `1_tables`
}

// Get is retrieving model from database
func (t *Table) Get(transaction *DbTransaction, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("ecosystem = ? and name = ?", t.Ecosystem, name).First(t))
}

// Create is creating record of model
func (t *Table) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(t).Error
}

// Delete is deleting model from database
func (t *Table) Delete(transaction *DbTransaction) error {
	return GetDB(transaction).Delete(t).Error
}

// ExistsByName finding table existence by name
func (t *Table) ExistsByName(transaction *DbTransaction, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("ecosystem= ? and name = ?", t.Ecosystem, name).First(t))
}

// IsExistsByPermissionsAndTableName returns columns existence by permission and table name
func (t *Table) IsExistsByPermissionsAndTableName(transaction *DbTransaction, columnName, tableName string) (bool, error) {
	return isFound(GetDB(transaction).Where(`ecosystem = ? AND (columns-> ? ) is not null AND name = ?`,
		t.Ecosystem, columnName, tableName).First(t))
}

// GetColumns returns columns from database
func (t *Table) GetColumns(transaction *DbTransaction, name, jsonKey string) (map[string]string, error) {
	keyStr := ""
	if jsonKey != "" {
		keyStr = `->'` + jsonKey + `'`
	}
	rows, err := GetDB(transaction).Raw(`SELECT data.* FROM "1_tables", jsonb_each_text(columns`+keyStr+`) AS data WHERE ecosystem = ? AND name = ?`, t.Ecosystem, name).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var key, value string
	result := map[string]string{}
	for rows.Next() {
		rows.Scan(&key, &value)
		result[key] = value
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPermissions returns table permissions by name
func (t *Table) GetPermissions(transaction *DbTransaction, name, jsonKey string) (map[string]string, error) {
	keyStr := ""
	if jsonKey != "" {
		keyStr = `->'` + jsonKey + `'`
	}
	rows, err := GetDB(transaction).Raw(`SELECT data.* FROM "1_tables", jsonb_each_text(permissions`+keyStr+`) AS data WHERE ecosystem = ? AND name = ?`, t.Ecosystem, name).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var key, value string
	result := map[string]string{}
	for rows.Next() {
		rows.Scan(&key, &value)
		result[key] = value
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *Table) Count() (count int64, err error) {
	err = GetDB(nil).Table(t.TableName()).Where("ecosystem= ?", t.Ecosystem).Count(&count).Error
	return
}

// CreateTable is creating table
func CreateTable(transaction *DbTransaction, tableName, colsSQL string) error {
	return GetDB(transaction).Exec(`CREATE TABLE "` + tableName + `" (
				"id" bigint NOT NULL DEFAULT '0',
				` + colsSQL + `
				);
				ALTER TABLE ONLY "` + tableName + `" ADD CONSTRAINT "` + tableName + `_pkey" PRIMARY KEY (id);`).Error
}

// GetAll returns all tables
func (t *Table) GetAll(prefix string) ([]Table, error) {
	result := make([]Table, 0)
	err := DBConn.Table("1_tables").Where("ecosystem = ?", prefix).Find(&result).Error
	return result, err
}

// func (t *Table) GetList(offset, limit int64) ([]Table, error) {
// 	var list []Table
// 	err := DBConn.Table(t.TableName()).Offset(offset).Limit(limit).Select("name").Order("name").Find(&list).Error
// 	return list, err
// }

// GetRowConditionsByTableNameAndID returns value of `conditions` field for table row by id
func GetRowConditionsByTableNameAndID(transaction *DbTransaction, tblname string, id int64) (string, error) {
	sql := `SELECT conditions FROM "` + tblname + `" WHERE id = ? LIMIT 1`
	return Single(transaction, sql, id).String()
}

func GetTableQuery(table string, ecosystemID int64) *gorm.DB {
	if FirstEcosystemTables[table] {
		return DBConn.Table("1_"+table).Where("ecosystem = ?", ecosystemID)
	}

	return DBConn.Table(converter.ParseTable(table, ecosystemID))
}
