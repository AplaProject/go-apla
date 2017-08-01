package model

import "github.com/EGaaS/go-egaas-mvp/packages/converter"

type Block struct {
	ID           int64  `gorm:"primary_key;not_null"`
	Hash         []byte `gorm:not null`
	Data         []byte `gorm:not null`
	StateID      int64  `gorm:not null`
	WalletID     int64  `gorm:not null`
	Time         int64  `gorm:not null`
	Tx           int32  `gorm:not null`
	Cur0lMinerID int32  `gorm:not null;column:cur_0l_miner_id`
	MaxMinerID   int32  `gorm:not null`
}

func GetBlockchain(startBlockID int64, endblockID int64) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	if endblockID == -1 {
		err = DBConn.Order("id asc").Where("id > ? AND id <= ?", startBlockID, endblockID).Find(blockchain).Error
	} else {
		err = DBConn.Order("id asc").Where("id > ?", startBlockID).Find(blockchain).Error
	}
	if err != nil {
		return nil, err
	}
	return *blockchain, nil
}

func TableName() string {
	return "block_chain"
}

func (b *Block) IsExists() (bool, error) {
	query := DBConn.First(b)
	return !query.RecordNotFound(), query.Error
}

func (b *Block) IsExistsID(blockID int64) (bool, error) {
	query := DBConn.Where("id = ?").First(b)
	return !query.RecordNotFound(), query.Error
}

func (b *Block) Create() error {
	return DBConn.Create(b).Error
}

func (b *Block) GetBlock(blockID int64) error {
	return DBConn.Where("id = ?", blockID).First(&b).Error
}

func (b *Block) GetMaxBlock() error {
	return DBConn.First(b).Error
}

func (b *Block) GetBlocksFrom(startFromID int64, ordering string) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	err = DBConn.Order("id "+ordering).Where("id > ?", startFromID).Find(blockchain).Error
	if err != nil {
		return nil, err
	}
	return *blockchain, nil
}

func (b *Block) GetBlocks(startFromID int64, limit int32) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	if startFromID != -1 {
		err = DBConn.Order("id desc").Limit(limit).Where("id > ?", startFromID).Find(blockchain).Error
	} else {
		err = DBConn.Order("id desc").Limit(limit).Last(blockchain).Error
	}
	if err != nil {
		return nil, err
	}
	return *blockchain, nil
}

func (b *Block) Delete() error {
	return DBConn.Delete(b).Error
}

func (b *Block) DeleteById(id int64) error {
	return DBConn.Where("id = ?", id).Delete(Block{}).Error
}

func (b *Block) DeleteChain() error {
	return DBConn.Where("id > ", b.ID).Delete(Block{}).Error
}

func (b *Block) GetLastBlockData() (map[string]int64, error) {
	result := make(map[string]int64)
	confirmation := &Confirmation{}
	err := confirmation.GetMaxGoodBlock()
	if err != nil {
		return result, err
	}
	if confirmation.BlockID == 0 {
		confirmation.BlockID = 1
	}

	err = b.GetBlock(confirmation.BlockID)
	if err != nil {
		return result, err
	}
	result["blockId"] = int64(converter.BinToDec(b.Data[1:5]))
	result["lastBlockTime"] = int64(converter.BinToDec(b.Data[5:9]))
	return result, nil
}

func (b *Block) ToMap() map[string]string {
	result := make(map[string]string)
	result["hash"] = string(b.Hash)
	result["state_id"] = string(b.StateID)
	result["wallet_id"] = converter.AddressToString(b.WalletID)
	result["time"] = string(b.Time)
	result["tx"] = string(b.Tx)
	result["id"] = string(b.ID)
	return result
}
