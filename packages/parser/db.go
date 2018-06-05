package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// UpdBlockInfo updates info_block table
func UpdBlockInfo(dbTransaction *model.DbTransaction, block *Block) error {
	blockID := block.Header.BlockID
	// for the local tests
	forSha := fmt.Sprintf("%d,%x,%s,%d,%d,%d,%d", blockID, block.PrevHeader.Hash, block.MrklRoot,
		block.Header.Time, block.Header.EcosystemID, block.Header.KeyID, block.Header.NodePosition)

	hash, err := crypto.DoubleHash([]byte(forSha))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("double hashing block")
	}

	block.Header.Hash = hash
	if block.Header.BlockID == 1 {
		ib := &model.InfoBlock{
			Hash:           hash,
			BlockID:        blockID,
			Time:           block.Header.Time,
			EcosystemID:    block.Header.EcosystemID,
			KeyID:          block.Header.KeyID,
			NodePosition:   converter.Int64ToStr(block.Header.NodePosition),
			CurrentVersion: fmt.Sprintf("%d", block.Header.Version),
		}
		err := ib.Create(dbTransaction)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating info block")
			return fmt.Errorf("error insert into info_block %s", err)
		}
	} else {
		ibUpdate := &model.InfoBlock{
			Hash:         hash,
			BlockID:      blockID,
			Time:         block.Header.Time,
			EcosystemID:  block.Header.EcosystemID,
			KeyID:        block.Header.KeyID,
			NodePosition: converter.Int64ToStr(block.Header.NodePosition),
			Sent:         0,
		}
		if err := ibUpdate.Update(dbTransaction); err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating info block")
			return fmt.Errorf("error while updating info_block: %s", err)
		}
	}

	return nil
}

// InsertIntoBlockchain inserts a block into the blockchain
func InsertIntoBlockchain(transaction *model.DbTransaction, block *Block) error {
	// for local tests
	blockID := block.Header.BlockID

	// record into the block chain
	bl := &model.Block{}
	err := bl.DeleteById(transaction, blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting block by id")
		return err
	}
	rollbackTx := &model.RollbackTx{}
	blockRollbackTxs, err := rollbackTx.GetBlockRollbackTransactions(transaction, blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block rollback txs")
		return err
	}
	buffer := bytes.Buffer{}
	for _, rollbackTx := range blockRollbackTxs {
		rollbackTxBytes, err := json.Marshal(rollbackTx)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling rollback_tx to json")
			return err
		}

		buffer.Write(rollbackTxBytes)
	}
	rollbackTxsHash, err := crypto.Hash(buffer.Bytes())
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("hashing block rollback_txs")
		return err
	}
	b := &model.Block{
		ID:            blockID,
		Hash:          block.Header.Hash,
		Data:          block.BinData,
		EcosystemID:   block.Header.EcosystemID,
		KeyID:         block.Header.KeyID,
		NodePosition:  block.Header.NodePosition,
		Time:          block.Header.Time,
		RollbacksHash: rollbackTxsHash,
		Tx:            int32(len(block.Parsers)),
	}
	blockTimeCalculator, err := utils.BuildBlockTimeCalculator()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating block")
		return err
	}
	validBlockTime := true
	if blockID > 1 {
		validBlockTime, err = blockTimeCalculator.ValidateBlock(b.NodePosition, time.Unix(b.Time, 0))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("block validation")
			return err
		}
	}
	if validBlockTime {
		err = b.Create(transaction)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating block")
			return err
		}
	} else {
		err := fmt.Errorf("Invalid block time: %d", block.Header.Time)
		log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error("invalid block time")
		return err
	}

	return nil
}

// InsertInLogTx is inserting tx in log
func InsertInLogTx(transaction *model.DbTransaction, binaryTx []byte, time int64) error {
	txHash, err := crypto.Hash(binaryTx)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Fatal("hashing binary tx")
	}
	ltx := &model.LogTransaction{Hash: txHash, Time: time}
	err = ltx.Create(transaction)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("insert logged transaction")
		return utils.ErrInfo(err)
	}
	return nil
}

// CheckLogTx checks if this transaction exists
// And it would have successfully passed a frontal test
func CheckLogTx(txBinary []byte, transactions, txQueue bool) error {
	searchedHash, err := crypto.Hash(txBinary)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal(err)
	}
	logTx := &model.LogTransaction{}
	found, err := logTx.GetByHash(searchedHash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting log transaction by hash")
		return utils.ErrInfo(err)
	}
	if found {
		log.WithFields(log.Fields{"tx_hash": searchedHash, "type": consts.DuplicateObject}).Error("double tx in log transactions")
		return utils.ErrInfo(fmt.Errorf("double tx in log_transactions %x", searchedHash))
	}

	if transactions {
		// check for duplicate transaction
		tx := &model.Transaction{}
		_, err := tx.GetVerified(searchedHash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting verified transaction")
			return utils.ErrInfo(err)
		}
		if len(tx.Hash) > 0 {
			log.WithFields(log.Fields{"tx_hash": tx.Hash, "type": consts.DuplicateObject}).Error("double tx in transactions")
			return utils.ErrInfo(fmt.Errorf("double tx in transactions %x", searchedHash))
		}
	}

	if txQueue {
		// check for duplicate transaction from queue
		qtx := &model.QueueTx{}
		found, err := qtx.GetByHash(nil, searchedHash)
		if found {
			log.WithFields(log.Fields{"tx_hash": searchedHash, "type": consts.DuplicateObject}).Error("double tx in queue")
			return utils.ErrInfo(fmt.Errorf("double tx in queue_tx %x", searchedHash))
		}
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting transaction from queue")
			return utils.ErrInfo(err)
		}
	}

	return nil
}

// GetBlockDataFromBlockChain is retrieving block data from blockchain
func GetBlockDataFromBlockChain(blockID int64) (*utils.BlockData, error) {
	BlockData := new(utils.BlockData)
	block := &model.Block{}
	_, err := block.Get(blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting block by ID")
		return BlockData, utils.ErrInfo(err)
	}

	header, err := utils.ParseBlockHeader(bytes.NewBuffer(block.Data), false)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	BlockData = &header
	BlockData.Hash = block.Hash
	return BlockData, nil
}
