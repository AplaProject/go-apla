package model

type Menu struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:not null;type:jsonb(PostgreSQL)`
	Conditions string `gorm:not null`
	RbID       int64  `gorm:not null`
}

func (m *Menu) SetTableName(newName string) {
	m.tableName = newName
}

func (m *Menu) TableName() string {
	return m.tableName
}

func (m *Menu) Create() error {
	return DBConn.Create(m).Error
}

func (m *Menu) GetByName(name string) error {
	return DBConn.Where("name = ?", name).Find(m).Error
}

func CreateStateMenuTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_menu" (
				"name" varchar(255)  NOT NULL DEFAULT '',
				"value" text  NOT NULL DEFAULT '',
				"conditions" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + stateID + `_menu" ADD CONSTRAINT "` + stateID + `_menu_pkey" PRIMARY KEY (name);
				`).Error
}
