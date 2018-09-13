package block

import (
	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// GetBlockDataFromBlockChain is retrieving block data from blockchain
func GetBlockDataFromBlockChain(hash []byte) (*blockchain.BlockHeader, error) {
	block := &blockchain.Block{}
	found, err := block.Get(hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting block by hash")
		return nil, utils.ErrInfo(err)
	}
	if !found {
		return nil, nil
	}
	return block.Header, nil
}
