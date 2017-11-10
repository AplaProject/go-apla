package model

type Config struct {
	MyBlockID              int32  `gorm:"not null"`
	KeyID            int64  `gorm:"not null"`
	EcosystemID                int64  `gorm:"not null"`
	BadBlocks              string `gorm:"not null"`
	AutoReload             int    `gorm:"not null"`
	FirstLoadBlockchainURL string `gorm:"column:first_load_blockchain_url;not null"`
	FirstLoadBlockchain    string `gorm:"not null"`
	CurrentLoadBlockchain  string `gorm:"not null"`
}

func (c *Config) TableName() string {
	return "config"
}

func UpdateConfig(field string, value interface{}) error {
	return DBConn.Model(&Config{}).Update(field, value).Error
}

func (c *Config) Get() (bool, error) {
	return isFound(DBConn.First(&c))
}

func (c *Config) Create() error {
	return DBConn.Create(c).Error
}

func (c *Config) ChangeBlockIDBatch(transaction *DbTransaction, oldBlockID int64, newBlockID int64) error {
	return GetDB(transaction).Model(c).Where("my_block_id < ?", oldBlockID).Update("my_block_id", newBlockID).Error
}
