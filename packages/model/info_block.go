package model

type InfoBlock struct {
	Hash           []byte `gorm:"not null"`
	StateID        int64  `gorm:"not null"`
	WalletID       int64  `gorm:"not null"`
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
	return DBConn.First(ib).Error
}

func (ib *InfoBlock) Update() error {
	return DBConn.Model(&InfoBlock{}).Updates(ib).Error
}

func (ib *InfoBlock) GetUnsended() error {
	return DBConn.Where("sent = ?", "0").First(&ib).Error
}

func (ib *InfoBlock) MarkSended() error {
	return DBConn.Model(ib).Update("sent", "1").Error
}

func (ib *InfoBlock) Save() error {
	return DBConn.Save(ib).Error
}

func (ib *InfoBlock) Create() error {
	return DBConn.Create(ib).Error
}
