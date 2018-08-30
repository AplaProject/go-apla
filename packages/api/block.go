package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"

	log "github.com/sirupsen/logrus"
)

type getMaxBlockIDResult struct {
	MaxBlockID int64  `json:"max_block_id"`
	Hash       string `json:"hash"`
}

func getMaxBlockID(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	block, hash, found, err := blockchain.GetLastBlock()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting max block")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("last block not found")
		return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}
	data.result = &getMaxBlockIDResult{
		MaxBlockID: block.Header.BlockID,
		Hash:       string(converter.BinToHex(hash)),
	}
	return nil
}

type getBlockInfoResult struct {
	Hash          []byte `json:"hash"`
	EcosystemID   int64  `json:"ecosystem_id"`
	KeyID         int64  `json:"key_id"`
	Time          int64  `json:"time"`
	Tx            int32  `json:"tx_count"`
	RollbacksHash []byte `json:"rollbacks_hash"`
}

func getBlockInfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	blockHash := converter.HexToBin(data.params["hash"].(string))
	block, found, err := blockchain.GetBlock(blockHash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "hash": blockHash}).Error("block with hash not found")
		return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}
	data.result = &getBlockInfoResult{Hash: blockHash, EcosystemID: block.Header.EcosystemID, KeyID: block.Header.KeyID, Time: block.Header.Time, RollbacksHash: block.Header.RollbacksHash}
	return nil
}
