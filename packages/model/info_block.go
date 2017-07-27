package model

type InfoBlock struct {
	Hash           []byte `gorm:"not null"`
	StateID        int64  `gorm:"not null default 0"`
	WalletID       int64  `gorm:"not null default 0"`
	BlockID        int64  `gorm:"not null"`
	Time           int32  `gorm:"not null default 0"`
	Level          int8   `gorm:"not null default 0"`
	CurrentVersion string `gorm:"not null default ''"`
	Sent           int8   `gorm:"not null default 0"`
}

func (ib *InfoBlock) GetInfoBlock() error {
	return DBConn.Last(ib).Error
}

func (ib *InfoBlock) GetUnsent() error {
	return DBConn.Where("sent = ?", "0").First(&ib).Error
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
