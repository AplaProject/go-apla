package model

type Config struct {
	MyBlockID              int32  `gorm:";not null";`
	DltWalletID            int64  `gorm:"not null";`
	StateID                int64  `gorm:"not null";`
	CitizenID              int64  `gorm:"not null";`
	BadBlocks              string `gorm:"not null";`
	AutoReload             int    `gorm:"not null";`
	FirstLoadBlockchainURL string `gorm:"column:first_load_blockchain_url;not null";`
	FirstLoadBlockchain    string `gorm:"not null";`
	CurrentLoadBlockchain  string `gorm:"not null";`
}

func (c *Config) TableName() string {
	return "config"
}

func UpdateConfig(field string, value string) error {
	return DBConn.Model(&Config{}).Update(field, value).Error
}

func (c *Config) GetConfig() error {
	return DBConn.First(&c).Error
}

func (c *Config) Save() error {
	return DBConn.Save(c).Error
}

func (c *Config) Create() error {
	return DBConn.Create(c).Error
}

func (c *Config) ChangeBlockID(oldBlockID int64, newBlockID int64) error {
	return DBConn.Model(c).Where("id = ?", oldBlockID).Update("id", newBlockID).Error
}

func (c *Config) ChangeBlockIDBatch(oldBlockID int64, newBlockID int64) error {
	return DBConn.Model(c).Where("id < ?", oldBlockID).Update("id", newBlockID).Error
}
