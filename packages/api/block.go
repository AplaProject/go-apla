//MIT License
//
//Copyright (c) 2016-2018 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

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
