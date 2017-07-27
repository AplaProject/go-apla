package model

type Pages struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:not null;type:jsonb(PostgreSQL)`
	Menu       string `gorm:"not null;size:255"`
	Conditions string `gorm:not null`
	RbID       int64  `gorm:not null`
}

func (p *Pages) SetTableName(newName string) {
	p.tableName = newName
}

func (p *Pages) TableName() string {
	return p.tableName
}

func (p *Pages) Create() error {
	return DBConn.Create(p).Error
}

func (p *Pages) GetByName(name string) error {
	return DBConn.Where("name = ?", name).Find(p).Error
}

func CreateStatePagesTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_pages" (
			    "name" varchar(255)  NOT NULL DEFAULT '',
			    "value" text  NOT NULL DEFAULT '',
			    "menu" varchar(255)  NOT NULL DEFAULT '',
			    "conditions" bytea  NOT NULL DEFAULT '',
			    "rb_id" bigint NOT NULL DEFAULT '0'
			    );
			    ALTER TABLE ONLY "` + stateID + `_pages" ADD CONSTRAINT "` + stateID + `_pages_pkey" PRIMARY KEY (name);`).Error

}
