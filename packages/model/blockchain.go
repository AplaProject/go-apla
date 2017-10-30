package model

type Block struct {
	ID       int64  `gorm:"primary_key;not_null"`
	Hash     []byte `gorm:"not null"`
	Data     []byte `gorm:"not null"`
	StateID  int64  `gorm:"not null"`
	WalletID int64  `gorm:"not null"`
	Time     int64  `gorm:"not null"`
	Tx       int32  `gorm:"not null"`
}

func (Block) TableName() string {
	return "block_chain"
}

func (b *Block) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(b).Error
}

func (b *Block) Get(blockID int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", blockID).First(b))
}

func (b *Block) GetMaxBlock() (bool, error) {
	return isFound(DBConn.Last(b))
}

func GetBlockchain(startBlockID int64, endblockID int64) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	if endblockID > 0 {
		err = DBConn.Model(&Block{}).Order("id asc").Where("id > ? AND id <= ?", startBlockID, endblockID).Find(&blockchain).Error
	} else {
		err = DBConn.Model(&Block{}).Order("id asc").Where("id > ?", startBlockID).Find(&blockchain).Error
	}
	if err != nil {
		return nil, err
	}
	return *blockchain, nil
}

func (b *Block) GetBlocks(startFromID int64, limit int32) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	if startFromID > 0 {
		err = DBConn.Order("id desc").Limit(limit).Where("id > ?", startFromID).Find(&blockchain).Error
	} else {
		err = DBConn.Order("id desc").Limit(limit).Find(&blockchain).Error
	}
	return *blockchain, err
}

func (b *Block) GetBlocksFrom(startFromID int64, ordering string) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	err = DBConn.Order("id "+ordering).Where("id > ?", startFromID).Find(&blockchain).Error
	return *blockchain, err
}

func (b *Block) DeleteById(transaction *DbTransaction, id int64) error {
	return GetDB(transaction).Where("id = ?", id).Delete(Block{}).Error
}
