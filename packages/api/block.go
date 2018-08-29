package api

import (
	"bytes"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/block"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type getMaxBlockIDResult struct {
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
	data.result = &getMaxBlockIDResult{block.ID}
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
	data.result = &getBlockInfoResult{Hash: block.Hash, EcosystemID: block.EcosystemID, KeyID: block.KeyID, Time: block.Time, Tx: block.Tx, RollbacksHash: block.RollbacksHash}
	return nil
}

type TxInfo struct {
	Hash         string
	ContractName string
	Params       map[string]interface{}
	KeyID        int64
}

func getBlocksTxInfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	startBlockID := data.params["block_id"].(int64)
	if startBlockID > 0 {
		startBlockID--
	}

	blocksCount := data.params["count"].(int64)

	blocks, err := model.GetBlockchain(startBlockID, startBlockID+blocksCount, model.OrderASC)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting blocks range")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	if len(blocks) == 0 {
		return errorAPI(w, "E_NOTFOUND", http.StatusNotFound)
	}

	result := map[int64][]TxInfo{}
	for _, blockModel := range blocks {
		blck, err := block.UnmarshallBlock(bytes.NewBuffer(blockModel.Data), blockModel.ID == 1)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err, "bolck_id": blockModel.ID}).Error("on unmarshalling block")
			return errorAPI(w, err, http.StatusInternalServerError)
		}

		txInfoCollection := make([]TxInfo, 0, len(blck.Transactions))
		for _, tx := range blck.Transactions {
			txInfo := TxInfo{
				Hash: string(tx.TxHash),
			}

			if tx.TxContract != nil {
				txInfo.ContractName = tx.TxContract.Name
				txInfo.Params = tx.TxData
			}

			if blck.Header.BlockID == 1 {
				txInfo.KeyID = blck.Header.KeyID
			} else {
				txInfo.KeyID = tx.TxHeader.KeyID
			}

			txInfoCollection = append(txInfoCollection, txInfo)
		}

		result[blockModel.ID] = txInfoCollection
	}

	data.result = result
	return nil
}
