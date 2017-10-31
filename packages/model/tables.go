package model

type Table struct {
	tableName   string
	Name        string `gorm:"primary_key;not null;size:100"`
	Permissions string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Columns     string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions  string `gorm:"not null"`
	RbID        int64  `gorm:"not null"`
}

func (t *Table) SetTablePrefix(prefix string) {
	t.tableName = prefix + "_tables"
}

func (t *Table) TableName() string {
	return t.tableName
}

func (t *Table) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(t))
}

func (t *Table) Create(transaction *DbTransaction) error {
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
