package api

import (
	"encoding/hex"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/block"
	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type maxBlockResult struct {
	MaxBlockID int64  `json:"max_block_id"`
	Hash       string `json:"hash"`
}

func getMaxBlockHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	block := &blockchain.Block{}
	_, _, found, err := blockchain.GetLastBlock(nil)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting max block")
		errorResponse(w, err)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("last block not found")
		errorResponse(w, errNotFound)
		return
	}
	hash, err := block.Hash()
	if err != nil {
		errorResponse(w, err)
		return
	}
	result := &maxBlockResult{
		MaxBlockID: block.Header.BlockID,
		Hash:       string(converter.BinToHex(hash)),
	}
	jsonResponse(w, result)
}

type blockInfoResult struct {
	Hash          []byte `json:"hash"`
	EcosystemID   int64  `json:"ecosystem_id"`
	KeyID         int64  `json:"key_id"`
	Time          int64  `json:"time"`
	Tx            int32  `json:"tx_count"`
	RollbacksHash []byte `json:"rollbacks_hash"`
}

func getBlockInfoHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	params := mux.Vars(r)

	blockHash := converter.HexToBin(params["hash"])
	block := &blockchain.Block{}
	found, err := block.Get(nil, blockHash)

	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		errorResponse(w, err)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "hash": blockHash}).Error("block with hash not found")
		errorResponse(w, errNotFound)
		return
	}

	jsonResponse(w, &blockInfoResult{
		Hash:          blockHash,
		EcosystemID:   block.Header.EcosystemID,
		KeyID:         block.Header.KeyID,
		Time:          block.Header.Time,
		Tx:            int32(len(block.TxHashes)),
		RollbacksHash: block.RollbacksHash,
	})
}

type TxInfo struct {
	Hash         []byte                 `json:"hash"`
	ContractName string                 `json:"contract_name"`
	Params       map[string]interface{} `json:"params"`
	KeyID        int64                  `json:"key_id"`
}

type blocksTxInfoForm struct {
	nopeValidator
	BlockHashHex string `schema:"block_hash"`
	Count        int64  `schema:"count"`
}

func getBlocksTxInfoHandler(w http.ResponseWriter, r *http.Request) {
	form := &blocksTxInfoForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	logger := getLogger(r)
	startBlockHash, err := hex.DecodeString(form.BlockHashHex)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding block hash from string")
		errorResponse(w, err)
	}

	blocks, err := blockchain.GetNBlocksFrom(nil, startBlockHash, int(form.Count), 1)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting blocks range")
		errorResponse(w, err)
		return
	}

	if len(blocks) == 0 {
		errorResponse(w, errNotFound)
		return
	}

	result := map[int64][]TxInfo{}
	for _, blck := range blocks {
		txInfoCollection := make([]TxInfo, 0, len(blck.TxHashes))
		txs, err := blck.Block.Transactions(nil)
		if err != nil {
			errorResponse(w, errNotFound)
			return
		}
		b, err := block.FromBlockchainBlock(blck.Block, txs, blck.Hash, nil)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err, "block_id": blck.Header.BlockID}).Error("on unmarshalling block")
			errorResponse(w, err)
			return
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

	jsonResponse(w, &result)
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

func getBlocksDetailedInfoHandler(w http.ResponseWriter, r *http.Request) {
	form := &blocksTxInfoForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	logger := getLogger(r)
	startBlockHash, err := hex.DecodeString(form.BlockHashHex)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding block hash from string")
		errorResponse(w, err)
		return
	}

	blocks, err := blockchain.GetNBlocksFrom(nil, startBlockHash, int(form.Count), 1)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting blocks range")
		errorResponse(w, err)
		return
	}

	if len(blocks) == 0 {
		errorResponse(w, errNotFound)
		return
	}

	result := map[int64]BlockDetailedInfo{}
	for _, blockModel := range blocks {
		txs, err := blockModel.Transactions(nil)
		if err != nil {
			errorResponse(w, errNotFound)
			return
		}
		blck, err := block.FromBlockchainBlock(blockModel.Block, txs, blockModel.Hash, nil)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err, "block_id": blockModel.Block.Header.BlockID}).Error("on unmarshalling block")
			errorResponse(w, err)
			return
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
			Tx:            int32(len(blockModel.Block.TxHashes)),
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

	jsonResponse(w, &result)
}
