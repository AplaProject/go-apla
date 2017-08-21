package model

type MyNodeKey struct {
	ID         int32  `gorm:"primary_key;not null"`
	AddTime    int32  `gorm:"not null"`
	PublicKey  []byte `gorm:"not null"`
	PrivateKey []byte `gorm:"not null"`
	Status     string `gorm:"not null;default:'my_pending'"`
	MyTime     int32  `gorm:"not null"`
	Time       int32  `gorm:"not null"`
	BlockID    int64  `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (mnk *MyNodeKey) GetNodeWithMaxBlockID() error {
	blockID := int64(0)
	err := DBConn.Raw("SELECT max(block_id) FROM my_node_keys").Row().Scan(&blockID)
	if err != nil {
		return err
	}

	if err := DBConn.Where("block_id = ?", blockID).First(mnk).Error; err != nil {
		return err
	}
	return nil
}

func (mnk *MyNodeKey) Create() error {
	return DBConn.Create(mnk).Error
}

func (mnk *MyNodeKey) GetZeroBlock(publicKey []byte) error {
	return DBConn.Where("block_id = 0 AND public_key = ?", publicKey).First(mnk).Error
}

func MyNodeKeysCreateTable() error {
	return DBConn.CreateTable(&MyNodeKey{}).Error
}
