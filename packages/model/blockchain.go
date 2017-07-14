package model

type Block struct {
	ID           int64  `gorm:"primary_key;not_null"`
	Hash         []byte `gorm:not null`
	Data         []byte `gorm:not null`
	StateID      int32  `gorm:not null`
	WalletID     int64  `gorm:not null`
	Time         int32  `gorm:not null`
	Tx           int32  `gorm:not null`
	Cur0lMinerID int32  `gorm:not null;column:cur_0l_miner_id`
	MaxMinerID   int32  `gorm:not null`
}

func TableName() string {
	return "block_chain"
}

func (b *Block) GetBlock(blockID int64) error {
	if err := DBConn.Where("block_id = ?", blockID).First(&b).Error; err != nil {
		return err
	}
	return nil
}

func GetBlockchain(startBlockID int64, endblockID int64) ([]Block, error) {
	blockchain := new([]Block)
	if err := DBConn.Order("id asc").Where("id > ? AND id <= ?", startBlockID, endblockID).Find(blockchain).Error; err != nil {
		return nil, err
	}
	return *blockchain, nil
}
