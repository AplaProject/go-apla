package model

// Confirmation is model
type Confirmation struct {
	BlockID int64 `gorm:"primary_key"`
	Good    int32 `gorm:"not null"`
	Bad     int32 `gorm:"not null"`
	Time    int32 `gorm:"not null"`
}

// GetGoodBlock returns last good block
func (c *Confirmation) GetGoodBlock(goodCount int) (bool, error) {
	return isFound(DBConn.Where("good >= ?", goodCount).Last(&c))
}

// GetConfirmation returns if block with blockID exists
func (c *Confirmation) GetConfirmation(blockID int64) (bool, error) {
	return isFound(DBConn.Where("block_id= ?", blockID).First(&c))
}

// Save is saving model
func (c *Confirmation) Save() error {
	return DBConn.Save(c).Error
}
