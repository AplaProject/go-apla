package model

type Table struct {
	tableName   string
	ID        int64  `gorm:"primary_key;not null"`
	Name        string `gorm:"not null;size:100"`
	Permissions string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Columns     string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions  string `gorm:"not null"`
	RbID        int64  `gorm:"not null"`
}

type TableVDE struct {
	tableName   string
	ID        int64  `gorm:"primary_key;not null"`
	Name        string `gorm:"not null;size:100"`
	Permissions string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Columns     string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions  string `gorm:"not null"`
}

func (t *Table) SetTablePrefix(prefix string) {
	t.tableName = prefix + "_tables"
}

func (t *TableVDE) SetTablePrefix(prefix string) {
	t.tableName = prefix + "_tables"
}

func (t *Table) TableName() string {
	return t.tableName
}

func (t *TableVDE) TableName() string {
	return t.tableName
}

func (t *Table) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(t))
}

func (t *Table) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(t).Error
}

func (t *TableVDE) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(t).Error
}

func (t *Table) Delete() error {
	return DBConn.Delete(t).Error
}

func (t *Table) ExistsByName(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(t))
}

func (t *Table) IsExistsByPermissionsAndTableName(columnName, tableName string) (bool, error) {
	return isFound(DBConn.Where(`(columns-> ? ) is not null AND name = ?`, columnName, tableName).First(t))
}

func (t *Table) GetColumns(name, jsonKey string) (map[string]string, error) {
	keyStr := ""
	if jsonKey != "" {
		keyStr = `->'` + jsonKey + `'`
	}
	rows, err := DBConn.Raw(`SELECT data.* FROM "`+t.tableName+`", jsonb_each_text(columns`+keyStr+`) AS data WHERE name = ?`, name).Rows()
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

func (t *Table) GetPermissions(name, jsonKey string) (map[string]string, error) {
	keyStr := ""
	if jsonKey != "" {
		keyStr = `->'` + jsonKey + `'`
	}
	rows, err := DBConn.Raw(`SELECT data.* FROM "`+t.tableName+`", jsonb_each_text(permissions`+keyStr+`) AS data WHERE name = ?`, name).Rows()
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

func CreateTable(transaction *DbTransaction, tableName, colsSQL string) error {
	return GetDB(transaction).Exec(`CREATE TABLE "` + tableName + `" (
				"id" bigint NOT NULL DEFAULT '0',
				` + colsSQL + `
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + tableName + `" ADD CONSTRAINT "` + tableName + `_pkey" PRIMARY KEY (id);`).Error
}

func CreateVDETable(transaction *DbTransaction, tableName, colsSQL string) error {
	return GetDB(transaction).Exec(`CREATE TABLE "` + tableName + `" (
				"id" bigint NOT NULL DEFAULT '0',
				` + colsSQL + `
				);
				ALTER TABLE ONLY "` + tableName + `" ADD CONSTRAINT "` + tableName + `_pkey" PRIMARY KEY (id);`).Error
}

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
