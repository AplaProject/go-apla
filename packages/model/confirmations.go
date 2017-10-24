package model

import (
	"github.com/AplaProject/go-apla/packages/consts"
)

type Confirmation struct {
	BlockID int64 `gorm:"primary_key"`
	Good    int32 `gorm:"not null"`
	Bad     int32 `gorm:"not null"`
	Time    int32 `gorm:"not null"`
}

func (c *Confirmation) GetGoodBlock(goodCount int) (bool, error) {
	return isFound(DBConn.Where("good >= ?", goodCount).Last(&c))
}

func (c *Confirmation) GetConfirmation(blockID int64) (bool, error) {
	return isFound(DBConn.Where("block_id= ?", blockID).First(&c))
}

func (c *Confirmation) GetMaxGoodBlock() (bool, error) {
	return isFound(DBConn.Order("block_id desc").Where("good >= ?", consts.MIN_CONFIRMED_NODES).First(c))
}

func (c *Confirmation) Update() error {
	return DBConn.Update(c).Error
}

func (c *Confirmation) Create() error {
	return DBConn.Create(c).Error
}

func (c *Confirmation) IsExists() (bool, error) {
	return isFound(DBConn.First(c))
}

func (c *Confirmation) Save() error {
	return DBConn.Save(c).Error
}
