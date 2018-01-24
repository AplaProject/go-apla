package api

import (
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type GetMaxBlockIDResult struct {
	MaxBlockID int64 `json:"max_block_id"`
}

func getMaxBlockID(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	block := &model.Block{}
	found, err := block.GetMaxBlock()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting max block")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("last block not found")
		return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}
	data.result = &GetMaxBlockIDResult{block.ID}
	return nil
}

type GetBlockInfoResult struct {
	Hash          []byte `json:"hash"`
	EcosystemID   int64  `json:"ecosystem_id"`
	KeyID         int64  `json:"key_id"`
	Time          int64  `json:"time"`
	Tx            int32  `json:"tx_count"`
	RollbacksHash []byte `json:"rollbacks_hash"`
}

func getBlockInfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	blockID := converter.StrToInt64(data.params["id"].(string))
	block := model.Block{}
	found, err := block.Get(blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "id": blockID}).Error("block with id not found")
		return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}
	data.result = &GetBlockInfoResult{Hash: block.Hash, EcosystemID: block.EcosystemID, KeyID: block.KeyID, Time: block.Time, Tx: block.Tx, RollbacksHash: block.RollbacksHash}
	return nil
}
