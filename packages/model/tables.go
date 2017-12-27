package model

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/jinzhu/gorm"
)

func NewBufferedTables() *bufferedTables {
	return &bufferedTables{tables: make(map[int64]map[string]Table)}
}

type bufferedTables struct {
	tables      map[int64]map[string]Table
	rwMutex     sync.RWMutex
	updateMutex sync.Mutex
}

func loadEcosystemTables(ecosystemID int64) (*[]Table, error) {
	var tables []Table
	err := DBConn.Raw(fmt.Sprintf(`select * from "%d_tables";`, ecosystemID)).Scan(&tables).Error
	if err != nil {
		return nil, err
	}
	return &tables, nil
}

func loadTable(ecosystemID int64, tableName string) (Table, error) {
	table := &Table{}
	table.SetTablePrefix(strconv.FormatInt(ecosystemID, 10))
	found, err := table.Get(nil, tableName)
	if !found {
		return *table, gorm.ErrRecordNotFound
	}
	return *table, err
}

func (bk *bufferedTables) updateEcosystemCache(tablePrefix int64) error {
	bk.updateMutex.Lock()
	defer bk.updateMutex.Unlock()
	tables, err := loadEcosystemTables(tablePrefix)
	if err != nil {
		return err
	}
	newEcosystemBuffer := make(map[string]Table, len(*tables))
	for _, t := range *tables {
		newEcosystemBuffer[t.Name] = t
	}

	bk.tables[tablePrefix] = newEcosystemBuffer
	return nil
}

func (bk *bufferedTables) updateKeyCache(tablePrefix int64, tableName string) error {
	bk.updateMutex.Lock()
	defer bk.updateMutex.Unlock()
	table, err := loadTable(tablePrefix, tableName)
	if err != nil {
		return err
	}
	bk.tables[tablePrefix][tableName] = table
	return nil
}

func (bk *bufferedTables) Initialize() error {
	IDs, err := GetAllSystemStatesIDs()
	if err != nil {
		return err
	}
	for _, ID := range IDs {
		err := bk.updateEcosystemCache(ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bk *bufferedTables) GetTable(tablePrefix int64, tableName string) (table Table, found bool, err error) {
	result := Table{}
	bk.rwMutex.RLock()
	defer bk.rwMutex.RUnlock()

	_, ok := bk.tables[tablePrefix]
	if !ok {
		err := bk.updateEcosystemCache(tablePrefix)
		if err != nil && err != gorm.ErrRecordNotFound {
			return result, false, err
		}
		if err == gorm.ErrRecordNotFound {
			return result, false, nil
		}
	}

	_, ok = bk.tables[tablePrefix][tableName]
	if !ok {
		err = bk.updateKeyCache(tablePrefix, tableName)
		if err != nil && err != gorm.ErrRecordNotFound {
			return result, false, err
		}
		if err == gorm.ErrRecordNotFound {
			return result, false, nil
		}
	}

	result = bk.tables[tablePrefix][tableName]
	return result, true, nil
}

func (bk *bufferedTables) SetTable(tablePrefix int64, tableName string, table Table) (found bool, err error) {
	bk.rwMutex.RLock()
	_, ok := bk.tables[tablePrefix]
	if !ok {
		err := bk.updateEcosystemCache(tablePrefix)
		if err != nil && err != gorm.ErrRecordNotFound {
			return false, err
		}
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
	}

	_, ok = bk.tables[tablePrefix]
	if !ok {
		err = bk.updateKeyCache(tablePrefix, tableName)
		if err != nil && err != gorm.ErrRecordNotFound {
			return false, err
		}
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
	}
	bk.rwMutex.RUnlock()

	bk.rwMutex.Lock()
	bk.tables[tablePrefix][tableName] = table
	bk.rwMutex.Unlock()
	return true, nil
}

// Table is model
type Table struct {
	tableName   string
	ID          int64  `gorm:"primary_key;not null"`
	Name        string `gorm:"not null;size:100"`
	Permissions string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Columns     string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions  string `gorm:"not null"`
}

// TableVDE is model
type TableVDE struct {
	tableName   string
	ID          int64  `gorm:"primary_key;not null"`
	Name        string `gorm:"not null;size:100"`
	Permissions string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Columns     string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions  string `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (t *Table) SetTablePrefix(prefix string) {
	t.tableName = prefix + "_tables"
}

// SetTablePrefix is setting table prefix
func (t *TableVDE) SetTablePrefix(prefix string) {
	t.tableName = prefix + "_tables"
}

// TableName returns name of table
func (t *Table) TableName() string {
	return t.tableName
}

// TableName returns name of table
func (t *TableVDE) TableName() string {
	return t.tableName
}

// Get is retrieving model from database
func (t *Table) Get(transaction *DbTransaction, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("name = ?", name).First(t))
}

// Create is creating record of model
func (t *Table) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(t).Error
}

// Create is creating record of model
func (t *TableVDE) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(t).Error
}

// Delete is deleting model from database
func (t *Table) Delete(transaction *DbTransaction) error {
	return GetDB(transaction).Delete(t).Error
}

// ExistsByName finding table existence by name
func (t *Table) ExistsByName(transaction *DbTransaction, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("name = ?", name).First(t))
}

// IsExistsByPermissionsAndTableName returns columns existence by permission and table name
func (t *Table) IsExistsByPermissionsAndTableName(transaction *DbTransaction, columnName, tableName string) (bool, error) {
	return isFound(GetDB(transaction).Where(`(columns-> ? ) is not null AND name = ?`, columnName, tableName).First(t))
}

// GetColumns returns columns from database
func (t *Table) GetColumns(transaction *DbTransaction, name, jsonKey string) (map[string]string, error) {
	keyStr := ""
	if jsonKey != "" {
		keyStr = `->'` + jsonKey + `'`
	}
	rows, err := GetDB(transaction).Raw(`SELECT data.* FROM "`+t.tableName+`", jsonb_each_text(columns`+keyStr+`) AS data WHERE name = ?`, name).Rows()
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
	rows, err := GetDB(transaction).Raw(`SELECT data.* FROM "`+t.tableName+`", jsonb_each_text(permissions`+keyStr+`) AS data WHERE name = ?`, name).Rows()
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

// CreateTable is creating table
func CreateTable(transaction *DbTransaction, tableName, colsSQL string) error {
	return GetDB(transaction).Exec(`CREATE TABLE "` + tableName + `" (
				"id" bigint NOT NULL DEFAULT '0',
				` + colsSQL + `
				);
				ALTER TABLE ONLY "` + tableName + `" ADD CONSTRAINT "` + tableName + `_pkey" PRIMARY KEY (id);`).Error
}

// CreateVDETable is creating VDE table
func CreateVDETable(transaction *DbTransaction, tableName, colsSQL string) error {
	return GetDB(transaction).Exec(`CREATE TABLE "` + tableName + `" (
				"id" bigint NOT NULL DEFAULT '0',
				` + colsSQL + `
				);
				ALTER TABLE ONLY "` + tableName + `" ADD CONSTRAINT "` + tableName + `_pkey" PRIMARY KEY (id);`).Error
}

// GetAll returns all tables
func (t *Table) GetAll(prefix string) ([]Table, error) {
	result := make([]Table, 0)
	err := DBConn.Table(prefix + "_tables").Find(&result).Error
	return result, err
}
