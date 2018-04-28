package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const keyBlockID = "id"

type maxBlockIDResult struct {
	MaxBlockID int64 `json:"max_block_id"`
}

type blockInfoResult struct {
	Hash          []byte `json:"hash"`
	EcosystemID   int64  `json:"ecosystem_id"`
	KeyID         int64  `json:"key_id"`
	Time          int64  `json:"time"`
	Tx            int32  `json:"tx_count"`
	RollbacksHash []byte `json:"rollbacks_hash"`
}

func maxBlockHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	block := &model.Block{}
	found, err := block.GetMaxBlock()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting max block")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("last block not found")
		errorResponse(w, errNotFound, http.StatusNotFound)
		return
	}

	jsonResponse(w, &maxBlockIDResult{block.ID})
}

func blockInfoHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	params := mux.Vars(r)

	blockID := converter.StrToInt64(params[keyBlockID])
	block := model.Block{}
	found, err := block.Get(blockID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "id": blockID}).Error("block with id not found")
		errorResponse(w, errNotFound, http.StatusNotFound)
		return
	}

	jsonResponse(w, &blockInfoResult{
		Hash:          block.Hash,
		EcosystemID:   block.EcosystemID,
		KeyID:         block.KeyID,
		Time:          block.Time,
		Tx:            block.Tx,
		RollbacksHash: block.RollbacksHash,
	})
}
