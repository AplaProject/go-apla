package block

import (
	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
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

// GetDataFromFirstBlock returns data of first block
func GetDataFromFirstBlock() (data *consts.FirstBlock, ok bool) {
	bBlock, _, found, err := blockchain.GetFirstBlock()
	if !found {
		return
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting record of first block")
		return
	}
	if len(bBlock.Transactions) == 0 {
		log.WithFields(log.Fields{"type": consts.ParserError}).Error("list of parsers is empty")
		return
	}

	t := bBlock.Transactions[0]
	fb := &consts.FirstBlock{}
	if err := msgpack.Unmarshal(t, fb); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("getting data of first block")
		return
	} else {
		data = fb
	}

	return
}
