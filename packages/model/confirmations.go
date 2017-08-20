package model

import (
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/jinzhu/gorm"
)

type Confirmation struct {
	BlockID int64 `gorm:"primary_key"`
	Good    int32 `gorm:"not null"`
	Bad     int32 `gorm:"not null"`
	Time    int32 `gorm:"not null"`
}

func (c *Confirmation) GetGoodBlock(goodCount int) error {
	return handleError(DBConn.Where("good >= ?", goodCount).Last(&c).Error)
}

func (c *Confirmation) GetConfirmation(blockID int64) error {
	return handleError(DBConn.Where("block_id= ?", blockID).First(&c).Error)
}

func (c *Confirmation) GetMaxGoodBlock() error {
	return handleError(DBConn.Order("block_id desc").Where("good >= ?", consts.MIN_CONFIRMED_NODES).First(c).Error)
}

func (c *Confirmation) Update() error {
	return DBConn.Update(c).Error
}

func (c *Confirmation) Create() error {
	return DBConn.Create(c).Error
}

func (c *Confirmation) IsExists() (bool, error) {
	query := DBConn.First(c)
	if query.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	return !query.RecordNotFound(), query.Error
}

func (c *Confirmation) Save() error {
	return DBConn.Save(c).Error
}
