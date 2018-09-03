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
	log.WithFields(log.Fields{"id": blockID, "hash": block.Hash, "ecosystem_id": block.EcosystemID, "key_id": block.KeyID, "time": block.Time, "tx": block.Tx, "rollbacks_hash": block.RollbacksHash}).Debug("Block Information")
	return nil
}

type TxInfo struct {
	Hash         []byte                 `json:"hash"`
	ContractName string                 `json:"contract_name"`
	Params       map[string]interface{} `json:"params"`
	KeyID        int64                  `json:"key_id"`
}

func getBlocksTxInfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	startBlockID := data.params["block_id"].(int64)
	if startBlockID > 0 {
		startBlockID--
	}

	blocksCount := data.params["count"].(int64)

	blocks, err := model.GetBlockchain(startBlockID, startBlockID+blocksCount)
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
				Hash: tx.TxHash,
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

			log.WithFields(log.Fields{"block_id": blockModel.ID, "tx hash": txInfo.Hash, "contract_name": txInfo.ContractName, "key_id": txInfo.KeyID, "params": txInfoCollection}).Debug("Block Transactions Information")
		}

		result[blockModel.ID] = txInfoCollection
	}

	data.result = result
	return nil
}

type TxDetailedInfo struct {
	Hash         []byte                 `json:"hash"`
	ContractName string                 `json:"contract_name"`
	Params       map[string]interface{} `json:"params"`
	KeyID        int64                  `json:"key_id"`
	Time         int64                  `json:"time"`
	Type         int64                  `json:"type"`
}

type BlockHeaderInfo struct {
	BlockID      int64  `json:"block_id"`
	Time         int64  `json:"time"`
	EcosystemID  int64  `json:"ecosystem_id"`
	KeyID        int64  `json:"key_id"`
	NodePosition int64  `json:"node_position"`
	Sign         []byte `json:"sign"`
	Hash         []byte `json:"hash"`
	Version      int    `json:"version"`
}

type BlockDetailedInfo struct {
	Header        BlockHeaderInfo  `json:"header"`
	Hash          []byte           `json:"hash"`
	EcosystemID   int64            `json:"ecosystem_id"`
	NodePosition  int64            `json:"node_position"`
	KeyID         int64            `json:"key_id"`
	Time          int64            `json:"time"`
	Tx            int32            `json:"tx_count"`
	RollbacksHash []byte           `json:"rollbacks_hash"`
	MrklRoot      []byte           `json:"mrkl_root"`
	BinData       []byte           `json:"bin_data"`
	SysUpdate     bool             `json:"sys_update"`
	GenBlock      bool             `json:"gen_block"`
	StopCount     int              `json:"stop_count"`
	Transactions  []TxDetailedInfo `json:"transactions"`
}

func getBlocksDetailedInfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	startBlockID := data.params["block_id"].(int64)
	if startBlockID > 0 {
		startBlockID--
	}

	blocksCount := data.params["count"].(int64)

	blocks, err := model.GetBlockchain(startBlockID, startBlockID+blocksCount)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting blocks range")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	if len(blocks) == 0 {
		return errorAPI(w, "E_NOTFOUND", http.StatusNotFound)
	}

	result := map[int64]BlockDetailedInfo{}
	for _, blockModel := range blocks {
		blck, err := block.UnmarshallBlock(bytes.NewBuffer(blockModel.Data), blockModel.ID == 1)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err, "bolck_id": blockModel.ID}).Error("on unmarshalling block")
			return errorAPI(w, err, http.StatusInternalServerError)
		}

		txDetailedInfoCollection := make([]TxDetailedInfo, 0, len(blck.Transactions))
		for _, tx := range blck.Transactions {
			txDetailedInfo := TxDetailedInfo{
				Hash: tx.TxHash,
			}

			if tx.TxContract != nil {
				txDetailedInfo.ContractName = tx.TxContract.Name
				txDetailedInfo.Params = tx.TxData
				txDetailedInfo.KeyID = tx.TxKeyID
				txDetailedInfo.Time = tx.TxTime
				txDetailedInfo.Type = tx.TxType
			}

			txDetailedInfoCollection = append(txDetailedInfoCollection, txDetailedInfo)

			log.WithFields(log.Fields{"block_id": blockModel.ID, "tx hash": txDetailedInfo.Hash, "contract_name": txDetailedInfo.ContractName, "key_id": txDetailedInfo.KeyID, "time": txDetailedInfo.Time, "type": txDetailedInfo.Type, "params": txDetailedInfoCollection}).Debug("Block Transactions Information")
		}

		header := BlockHeaderInfo{
			BlockID:      blck.Header.BlockID,
			Time:         blck.Header.Time,
			EcosystemID:  blck.Header.EcosystemID,
			KeyID:        blck.Header.KeyID,
			NodePosition: blck.Header.NodePosition,
			Sign:         blck.Header.Sign,
			Hash:         blck.Header.Hash,
			Version:      blck.Header.Version,
		}

		bdi := BlockDetailedInfo{
			Header:        header,
			Hash:          blockModel.Hash,
			EcosystemID:   blockModel.EcosystemID,
			NodePosition:  blockModel.NodePosition,
			KeyID:         blockModel.KeyID,
			Time:          blockModel.Time,
			Tx:            blockModel.Tx,
			RollbacksHash: blockModel.RollbacksHash,
			MrklRoot:      blck.MrklRoot,
			BinData:       blck.BinData,
			SysUpdate:     blck.SysUpdate,
			GenBlock:      blck.GenBlock,
			StopCount:     blck.StopCount,
			Transactions:  txDetailedInfoCollection,
		}
		result[blockModel.ID] = bdi
	}

	data.result = result
	return nil
}
