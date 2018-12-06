package block

import (
	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

// GetBlockDataFromBlockChain is retrieving block data from blockchain
func GetBlockDataFromBlockChain(ldbtx *leveldb.Transaction, hash []byte) (*blockchain.BlockHeader, error) {
	block := &blockchain.Block{}
	found, err := block.Get(ldbtx, hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting block by hash")
		return nil, utils.ErrInfo(err)
	}
	if !found {
		return nil, nil
	}
	return block.Header, nil
}
