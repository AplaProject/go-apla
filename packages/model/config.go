package model

import (
	"github.com/AplaProject/go-apla/packages/consts"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	MyBlockID              int32  `gorm:"not null"` // !!! remove
	KeyID                  int64  `gorm:"not null"`
	EcosystemID            int64  `gorm:"not null"`
	BadBlocks              string `gorm:"not null"`                                  // only read
	AutoReload             int    `gorm:"not null"`                                  // not used
	FirstLoadBlockchainURL string `gorm:"column:first_load_blockchain_url;not null"` // install -> blocks_colletcion
	FirstLoadBlockchain    string `gorm:"not null"`                                  // install -> blocks_collection == 'file'
	CurrentLoadBlockchain  string `gorm:"not null"`                                  // not used
}

// TableName returns name of table
func (c *Config) TableName() string {
	return "config"
}

// UpdateConfig is updates config
func UpdateConfig(field string, value interface{}) error {
	return DBConn.Model(&Config{}).Update(field, value).Error
}

// Get is retrieving model from database
func (c *Config) Get() (bool, error) {
	return isFound(DBConn.First(&c))
}

// Create is creating record of model
func (c *Config) Create() error {
	return DBConn.Create(c).Error
}

// ChangeBlockIDBatch is bulk changing block ids
func (c *Config) ChangeBlockIDBatch(transaction *DbTransaction, oldBlockID int64, newBlockID int64) error {
	return GetDB(transaction).Model(c).Where("my_block_id < ?", oldBlockID).Update("my_block_id", newBlockID).Error
}

// GetConfig returns config record
func GetConfig() (*Config, error) {
	config := &Config{}
	_, err := config.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("get config")
		return nil, err
	}
	return config, nil
}

//.
