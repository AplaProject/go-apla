package blockchain

import (
	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

const confirmationPrefix = "prefix-"
const lastBlockCount = 20

// Confirmation is model
type Confirmation struct {
	BlockID int64 `gorm:"primary_key"`
	Good    int   `gorm:"not null"`
	Bad     int   `gorm:"not null"`
	Time    int64 `gorm:"not null"`
}

func (c *Confirmation) Marshal() ([]byte, error) {
	val, err := msgpack.Marshal(c)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling confirmation")
		return nil, err
	}
	return val, err
}

func (c *Confirmation) Unmarshal(b []byte) error {
	if err := msgpack.Unmarshal(b, c); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
		return err
	}
	return nil
}

func (c *Confirmation) Get(tx *leveldb.Transaction, blockHash []byte) (bool, error) {
	val, err := GetDB(tx).Get([]byte(confirmationPrefix+string(blockHash)), nil)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting block")
		return false, err
	}
	if err := c.Unmarshal(val); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
		return true, err
	}
	return true, nil
}

func (c *Confirmation) Insert(tx *leveldb.Transaction, blockHash []byte) error {
	val, err := c.Marshal()
	if err != nil {
		return err
	}
	err = GetDB(tx).Put([]byte(confirmationPrefix+string(blockHash)), val, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting confirmation")
		return err
	}
	return nil
}

func GetUnconfirmedBlocks(tx *leveldb.Transaction, confirmedNodeCount int) ([]*BlockWithHash, error) {
	var result []*BlockWithHash
	blocks, err := GetLastNBlocks(tx, lastBlockCount)
	if err != nil {
		return result, err
	}
	for _, block := range blocks {
		confirmation := &Confirmation{}
		found, err := confirmation.Get(tx, []byte(block.Hash))
		if err != nil {
			return result, err
		}
		if !found {
			continue
		}
		if confirmation.Good < confirmedNodeCount {
			result = append(result, block)
		}
	}
	return result, nil
}
