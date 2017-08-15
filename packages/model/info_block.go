package model

import "github.com/jinzhu/gorm"

type InfoBlock struct {
	Hash           []byte `gorm:"not null"`
	StateID        int64  `gorm:"not null default 0"`
	WalletID       int64  `gorm:"not null default 0"`
	BlockID        int64  `gorm:"not null"`
	Time           int64  `gorm:"not null"`
	Level          int8   `gorm:"not null"`
	CurrentVersion string `gorm:"not null"`
	Sent           int8   `gorm:"not null"`
}

func (ib *InfoBlock) TableName() string {
	return "info_block"
}

func (ib *InfoBlock) GetInfoBlock() error {
	return handleError(DBConn.Last(ib).Error)
}

func (ib *InfoBlock) Update() error {
	return DBConn.Model(&InfoBlock{}).Updates(ib).Error
}

func (ib *InfoBlock) GetUnsent() (bool, error) {
	err := DBConn.Where("sent = ?", "0").First(&ib).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return true, err
}

func (ib *InfoBlock) MarkSent() error {
	return DBConn.Model(ib).Update("sent", "1").Error
}

func (ib *InfoBlock) Save() error {
	return DBConn.Save(ib).Error
}

func (ib *InfoBlock) Create() error {
	return DBConn.Create(ib).Error
}

func GetCurBlockID() (int64, error) {
	curBlock := &InfoBlock{}
	err := curBlock.GetInfoBlock()
	if err != nil {
		return 0, err
	}
	return curBlock.BlockID, nil
}

func InfoBlockCreateTable() error {
	return DBConn.CreateTable(&InfoBlock{}).Error
}

func BlockGetUnsent() (*InfoBlock, error) {
	ib := &InfoBlock{}
	found, err := ib.GetUnsent()
	if !found {
		return nil, err
	}
	return ib, err
}
