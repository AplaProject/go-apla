package model

import "github.com/EGaaS/go-egaas-mvp/packages/consts"

type Confirmation struct {
	BlockID int64 `gorm:"primary_key;not null"`
	Good    int32 `gorm:"not null"`
	Bad     int32 `gorm:"not null"`
	Time    int32 `gorm:"not null"`
}

func (c *Confirmation) GetGoodBlock(goodCount int) error {
	return DBConn.Where("good >= ?", goodCount).Last(&c).Error
}

func (c *Confirmation) GetConfirmation(blockID int64) error {
	return DBConn.Where("blockID = ?", blockID).First(&c).Error
}

func (c *Confirmation) GetMaxGoodBlock() error {
	return DBConn.Order("id desc").Where("good >= ?", consts.MIN_CONFIRMED_NODES).First(c).Error
}

func (c *Confirmation) Update() error {
	return DBConn.Update(c).Error
}

func (c *Confirmation) Create() error {
	return DBConn.Create(c).Error
}

func (c *Confirmation) IsExists() (bool, error) {
	query := DBConn.First(c)
	return !query.RecordNotFound(), query.Error
}

func (c *Confirmation) Save() error {
	return DBConn.Save(c).Error
}
