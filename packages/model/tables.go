package model

type Tables struct {
	tableName             string
	Name                  string `gorm:"primary_key;not null;size:100"`
	ColumnsAndPermissions string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions            string `gorm:"not null"`
	RbID                  int64  `gorm:"not null"`
}

func (t *Tables) SetTableName(prefix string) {
	t.tableName = prefix + "_tables"
}

func (t *Tables) TableName() string {
	return t.tableName
}

func (t *Tables) Get(name []byte) error {
	return DBConn.Where("name = ?", name).First(t).Error
}

func (t *Tables) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = string(t.Name)
	result["columns_and_permissions"] = t.ColumnsAndPermissions
	result["conditions"] = t.Conditions
	result["rb_id"] = string(t.RbID)
	return result
}

func (t *Tables) GetAll(prefix string) ([]Tables, error) {
	var result []Tables
	err := DBConn.Table(prefix + "_tables").Find(result).Error
	return result, err
}

func (t *Tables) GetTablePermissions(tablePrefix string, tableName string) (map[string]string, error) {
	result := make(map[string]string, 0)
	err := DBConn.Table(tablePrefix+"tables").
		Select("jsonb_each_text(columns_and_permissions)").
		Where("name = ?", tableName).Scan(result).Error
	return result, err
}

func (t *Tables) GetColumnsAndPermissions(tablePrefix string, tableName string) (map[string]string, error) {
	result := make(map[string]string, 0)
	err := DBConn.Table(tablePrefix+"tables").
		Select("jsonb_each_text(columns_and_permissions->'update')").
		Where("name = ?", tableName).Scan(result).Error
	return result, err
}
