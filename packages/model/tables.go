package model

type Tables struct {
	tableName             string
	Name                  []byte `gorm:"primary_key;not null"`
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
