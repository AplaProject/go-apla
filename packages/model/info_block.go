package model

type InfoBlock struct {
	Hash           []byte `gorm:"not null"`
	NodePosition        int64  `gorm:"not null default 0"`
	EcosystemID       int64  `gorm:"not null default 0"`
	KeyID       int64  `gorm:"not null default 0"`
	BlockID        int64  `gorm:"not null"`
	Time           int64  `gorm:"not null"`
	CurrentVersion string `gorm:"not null"`
	Sent           int8   `gorm:"not null"`
}

func (ib *InfoBlock) TableName() string {
	return "info_block"
}

func (ib *InfoBlock) Get() (bool, error) {
	return isFound(DBConn.Last(ib))
}

func (ib *InfoBlock) Update(transaction *DbTransaction) error {
	return GetDB(transaction).Model(&InfoBlock{}).Updates(ib).Error
}

func (ib *InfoBlock) GetUnsent() (bool, error) {
	return isFound(DBConn.Where("sent = ?", "0").First(&ib))
}

func (ib *InfoBlock) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(ib).Error
}

func (ib *InfoBlock) MarkSent() error {
	return DBConn.Model(ib).Update("sent", "1").Error
}

func BlockGetUnsent() (*InfoBlock, error) {
	ib := &InfoBlock{}
	found, err := ib.GetUnsent()
	if !found {
		return nil, err
	}
	return ib, err
}
