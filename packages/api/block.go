package api

import (
	"encoding/hex"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/block"
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
	block, hash, found, err := blockchain.GetLastBlock(nil)
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
	block := &blockchain.Block{}
	found, err := block.Get(nil, blockHash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "hash": blockHash}).Error("block with hash not found")
		return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}
	data.result = &getBlockInfoResult{Hash: blockHash, EcosystemID: block.Header.EcosystemID, KeyID: block.Header.KeyID, Time: block.Header.Time, Tx: int32(len(block.Transactions)), RollbacksHash: block.Header.RollbacksHash}
	log.WithFields(log.Fields{"id": block.Header.BlockID, "hash": blockHash, "ecosystem_id": block.Header.EcosystemID, "key_id": block.Header.KeyID, "time": block.Header.Time, "rollbacks_hash": block.RollbacksHash}).Debug("Block Information")
	return nil
}

type TxInfo struct {
	Hash         []byte                 `json:"hash"`
	ContractName string                 `json:"contract_name"`
	Params       map[string]interface{} `json:"params"`
	KeyID        int64                  `json:"key_id"`
}

func getBlocksTxInfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	startBlockHashHex := data.params["block_hash"].(string)
	blocksCount := data.params["count"].(int64)
	startBlockHash, err := hex.DecodeString(startBlockHashHex)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding block hash from string")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	blocks, err := blockchain.GetNBlocksFrom(nil, startBlockHash, int(blocksCount), 1)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting blocks range")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	if len(blocks) == 0 {
		return errorAPI(w, "E_NOTFOUND", http.StatusNotFound)
	}

	result := map[int64][]TxInfo{}
	for _, blck := range blocks {
		txInfoCollection := make([]TxInfo, 0, len(blck.Transactions))
		b, err := block.FromBlockchainBlock(blck.Block, blck.Hash, nil)
		if err != nil {
			return err
		}
		for _, tx := range b.Transactions {
			txInfo := TxInfo{
				Hash: tx.TxHash,
			}

			if tx.TxContract != nil {
				txInfo.ContractName = tx.TxContract.Name
				txInfo.Params = tx.TxData
			}

			txInfo.KeyID = tx.TxKeyID
			txInfoCollection = append(txInfoCollection, txInfo)

			log.WithFields(log.Fields{"block_id": blck.Header.BlockID, "tx hash": txInfo.Hash, "contract_name": txInfo.ContractName, "key_id": txInfo.KeyID, "params": txInfoCollection}).Debug("Block Transactions Information")
		}

		result[blck.Block.Header.BlockID] = txInfoCollection
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
	startBlockHashHex := data.params["block_hash"].(string)
	blocksCount := data.params["count"].(int64)
	startBlockHash, err := hex.DecodeString(startBlockHashHex)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding block hash from string")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	blocks, err := blockchain.GetNBlocksFrom(nil, startBlockHash, int(blocksCount), 1)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting blocks range")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	if len(blocks) == 0 {
		return errorAPI(w, "E_NOTFOUND", http.StatusNotFound)
	}

	result := map[int64]BlockDetailedInfo{}
	for _, blockModel := range blocks {
		blck, err := block.FromBlockchainBlock(blockModel.Block, blockModel.Hash, nil)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err, "block_id": blockModel.Block.Header.BlockID}).Error("on unmarshalling block")
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

		}

		header := BlockHeaderInfo{
			BlockID:      blck.Header.BlockID,
			Time:         blck.Header.Time,
			EcosystemID:  blck.Header.EcosystemID,
			KeyID:        blck.Header.KeyID,
			NodePosition: blck.Header.NodePosition,
			Sign:         blck.Header.Sign,
			Hash:         blockModel.Hash,
			Version:      blck.Header.Version,
		}

		bdi := BlockDetailedInfo{
			Header:        header,
			Hash:          blockModel.Hash,
			EcosystemID:   blockModel.Block.Header.EcosystemID,
			NodePosition:  blockModel.Block.Header.NodePosition,
			KeyID:         blockModel.Block.Header.KeyID,
			Time:          blockModel.Block.Header.Time,
			Tx:            int32(len(blockModel.Block.Transactions)),
			RollbacksHash: blockModel.Block.RollbacksHash,
			MrklRoot:      blck.MrklRoot,
			BinData:       blck.BinData,
			SysUpdate:     blck.SysUpdate,
			GenBlock:      blck.GenBlock,
			StopCount:     blck.StopCount,
			Transactions:  txDetailedInfoCollection,
		}
		result[blockModel.Block.Header.BlockID] = bdi
	}

	data.result = result
	return nil
}
