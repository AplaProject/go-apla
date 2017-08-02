package model

type Citizens struct {
	tableName string
	ID        int64  `gorm:"primary_key;not null"`
	PublicKey []byte `gorm:"not null;column:publick_key_0"`
	BlockID   int64  `gorm:"not null"`
	RbID      int64  `gorm:"not null"`
	Avatar    string
	Name      string
}

func (c *Citizens) SetTableName(tablePrefix int64) {
	c.tableName = string(tablePrefix) + "_citizens"
}

func (c *Citizens) TableName() string {
	return c.tableName
}

func (c *Citizens) Create() error {
	return DBConn.Create(c).Error
}

func GetAllCitizensWhereIdMoreThan(tablePrefix int64, id int64, limit int64) ([]Citizens, error) {
	citizens := new([]Citizens)
	err := DBConn.Table(string(tablePrefix)+"_citizens").Order("id").Where("id >= ?", id).Limit(limit).Find(citizens).Error
	if err != nil {
		return nil, err
	}
	return *citizens, nil
}

func (c *Citizens) IsExists() (bool, error) {
	query := DBConn.Where("id = ?", c.ID).First(c)
	return !query.RecordNotFound(), query.Error
}

func (c *Citizens) Get(id int64) error {
	return DBConn.Where("id = ?", id).First(c).Error
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
