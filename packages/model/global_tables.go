package model

type GlobalTables struct {
	Name                  []byte `gorm:"primary_key;not null"`
	ColumnsAndPermissions string `gorm:not null;type:jsonb(PostgreSQL)`
	Conditions            string `gorm:not null`
	RbID                  int64  `gorm:not null`
}
