package model

type Citizens struct {
	tableName string
	ID        int64  `gorm:"primary_key;not null"`
	PublicKey []byte `gorm:"column:public_key_0;not_null"`
	BlockID   int64  `gorm:"not null"`
	RbID      int64  `gorm:"not null"`
}

func (c *Citizens) TableName() string {
	return c.tableName
}

func (c *Citizens) SetTableName(tableName string) {
	c.tableName = tableName
}

func (c *Citizens) Create() error {
	return DBConn.Create(c).Error
}

func CreateCitizensStateTable(stateID string) error {
	return DBConn.Exec(`CREATE TABLE "` + stateID + `_citizens" (
				"id" bigint NOT NULL DEFAULT '0',
				"public_key_0" bytea  NOT NULL DEFAULT '',				
				"block_id" bigint NOT NULL DEFAULT '0',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "` + stateID + `_citizens" ADD CONSTRAINT "` + stateID + `_citizens_pkey" PRIMARY KEY (id);
		   `).Error
}
