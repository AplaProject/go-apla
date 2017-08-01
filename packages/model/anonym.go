package model

type Anonym struct {
	tableName string
	IDCitizen int64  `gorm:"not null"`
	IDAnonym  int64  `gorm:"not null"`
	Encrypted []byte `gorm:"not null"`
}

func (a *Anonym) SetTableName(prefix int64) {
	a.tableName = string(prefix) + "_anonyms"
}

func (a *Anonym) TableName() string {
	return a.tableName
}

func (a *Anonym) Create() error {
	return DBConn.Create(a).Error
}
