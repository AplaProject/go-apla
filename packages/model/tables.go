package model

import "strconv"

// Table is model
type Table struct {
	tableName   string
	ID          int64  `gorm:"primary_key;not null"`
	Name        string `gorm:"not null;size:100"`
	Permissions string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Columns     string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions  string `gorm:"not null"`
	RbID        int64  `gorm:"not null"`
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
func (t *Table) IsExistsByPermissionsAndTableName(columnName, tableName string) (bool, error) {
	return isFound(DBConn.Where(`(columns-> ? ) is not null AND name = ?`, columnName, tableName).First(t))
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
				"rb_id" bigint NOT NULL DEFAULT '0'
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

// GetColumnsAndPermissionsAndRbIDWhereTable returns columns and permissions
func GetColumnsAndPermissionsAndRbIDWhereTable(transaction *DbTransaction, table, tableName string) (map[string]string, error) {
	type proxy struct {
		ColumnsAndPermissions string
		RbID                  int64
	}
	temp := &proxy{}
	err := GetDB(transaction).Table(table).Where("name = ?", tableName).Select("columns_and_permissions, rb_id").Find(temp).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, 0)
	result["columns_and_permissions"] = temp.ColumnsAndPermissions
	result["rb_id"] = strconv.FormatInt(temp.RbID, 10)
	return result, nil
}

// GetTableWhereUpdatePermissionAndTableName returns tables
func GetTableWhereUpdatePermissionAndTableName(table, columnName, tableName string) (map[string]string, error) {
	type proxy struct {
		ColumnsAndPermissions string
		RbID                  int64
	}
	temp := &proxy{}
	err := DBConn.Table(table).Where("(columns-> ? ) is not null AND name = ?", columnName, tableName).Select("columns_and_permissions, rb_id").Find(temp).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, 0)
	result["columns_and_permissions"] = temp.ColumnsAndPermissions
	result["rb_id"] = strconv.FormatInt(temp.RbID, 10)
	return result, nil
}

// GetAll returns all tables
func (t *Table) GetAll(prefix string) ([]Table, error) {
	result := make([]Table, 0)
	err := DBConn.Table(prefix + "_tables").Find(&result).Error
	return result, err
}
