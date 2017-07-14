package model

type Config struct {
	BlockID                int32  `gorm:"not null";`
	DltWalletID            int64  `gorm:"not null";`
	StateID                int64  `gorm:"not null";`
	CitizenID              int64  `gorm:"not null";`
	BadBlocks              string `gorm:"not null";`
	AutoReload             int    `gorm:"not null";`
	FirstLoadBlockchainURL string `gorm:"column:first_load_blockchain_url;not null";`
	FirstLoadBlockchain    string `gorm:"not null";`
	CurrentLoadBlockchain  string `gorm:"not null";`
}

func (c *Config) GetConfig() error {
	return DBConn.First(&c).Error
}

func (c *Config) Save() error {
	return DBConn.Save(c).Error
}
