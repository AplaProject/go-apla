// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/autoupdate"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// GetTxTypeAndUserID returns tx type, wallet and citizen id from the block data
func GetTxTypeAndUserID(binaryBlock []byte) (txType int64, keyID int64) {
	tmp := binaryBlock[:]
	txType = converter.BinToDecBytesShift(&binaryBlock, 1)
	if consts.IsStruct(int(txType)) {
		var txHead consts.TxHeader
		converter.BinUnmarshal(&tmp, &txHead)
		keyID = txHead.KeyID
	}
	return
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

	header, err := ParseBlockHeader(bytes.NewBuffer(block.Data))
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	BlockData = &header
	BlockData.Hash = block.Hash
	return BlockData, nil
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
	if _, err := model.MarkTransactionUsed(transaction, txHash); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": txHash}).Error("marking transaction used")
		return err
	}
	if _, err := model.DeleteQueueTxByHash(transaction, txHash); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": txHash}).Error("DeleteQueueTxByHash")
	}
	return nil
}

// IsState returns if country is state
func IsState(transaction *model.DbTransaction, country string) (int64, error) {
	ids, err := model.GetAllSystemStatesIDs()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get all system states ids")
		return 0, err
	}
	for _, id := range ids {
		sp := &model.StateParameter{}
		sp.SetTablePrefix(converter.Int64ToStr(id))
		_, err = sp.Get(transaction, "state_name")
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("state get by name transaction")
			return 0, err
		}
		if strings.ToLower(sp.Name) == strings.ToLower(country) {
			return id, nil
		}
	}
	return 0, nil
}

// ParserInterface is parsing transactions
type ParserInterface interface {
	Init() error
	Validate() error
	Action() error
	Rollback() error
	Header() *tx.Header
}

// GetTablePrefix returns table prefix
func GetTablePrefix(global string, stateId int64) (string, error) {
	globalInt, err := strconv.Atoi(global)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ConversionError}).Error("converting global to int")
		return "", err
	}
	stateIdStr := converter.Int64ToStr(stateId)
	if globalInt == 1 {
		return "global", nil
	}
	return stateIdStr, nil
}

// GetParser returns ParserInterface
func GetParser(p *Parser, txType string) (ParserInterface, error) {
	switch txType {
	case "FirstBlock":
		return &FirstBlockParser{p}, nil
	}
	log.WithFields(log.Fields{"tx_type": txType, "type": consts.UnknownObject}).Error("unknown txType")
	return nil, fmt.Errorf("Unknown txType: %s", txType)
}

// Parser is a structure for parsing transactions
type Parser struct {
	BlockData      *utils.BlockData
	PrevBlock      *utils.BlockData
	dataType       int
	blockData      []byte
	CurrentVersion string
	MrklRoot       []byte
	PublicKeys     [][]byte

	TxBinaryData     []byte // transaction binary data
	TxFullData       []byte // full transaction, with type and data
	TxHash           []byte
	TxSlice          [][]byte
	TxMap            map[string][]byte
	TxIds            int // count of transactions
	TxKeyID          int64
	TxEcosystemIDStr string
	TxEcosystemID    int64
	TxNodePosition   uint32
	TxTime           int64
	TxType           int64
	TxCost           int64           // Maximum cost of executing contract
	TxUsedCost       decimal.Decimal // Used cost of CPU resources
	TxPtr            interface{}     // Pointer to the corresponding struct in consts/struct.go
	TxData           map[string]interface{}
	TxSmart          *tx.SmartContract
	TxContract       *smart.Contract
	TxHeader         *tx.Header
	txParser         ParserInterface
	DbTransaction    *model.DbTransaction
	SysUpdate        bool

	SmartContract smart.SmartContract
}

// GetLogger returns logger
func (p Parser) GetLogger() *log.Entry {
	if p.BlockData != nil && p.PrevBlock != nil {
		logger := log.WithFields(log.Fields{"block_id": p.BlockData.BlockID, "block_time": p.BlockData.Time, "block_wallet_id": p.BlockData.KeyID, "block_state_id": p.BlockData.EcosystemID, "block_hash": p.BlockData.Hash, "block_version": p.BlockData.Version, "prev_block_id": p.PrevBlock.BlockID, "prev_block_time": p.PrevBlock.Time, "prev_block_wallet_id": p.PrevBlock.KeyID, "prev_block_state_id": p.PrevBlock.EcosystemID, "prev_block_hash": p.PrevBlock.Hash, "prev_block_version": p.PrevBlock.Version, "tx_type": p.TxType, "tx_time": p.TxTime, "tx_state_id": p.TxEcosystemID, "tx_wallet_id": p.TxKeyID})
		return logger
	}
	if p.BlockData != nil {
		logger := log.WithFields(log.Fields{"block_id": p.BlockData.BlockID, "block_time": p.BlockData.Time, "block_wallet_id": p.BlockData.KeyID, "block_state_id": p.BlockData.EcosystemID, "block_hash": p.BlockData.Hash, "block_version": p.BlockData.Version, "tx_type": p.TxType, "tx_time": p.TxTime, "tx_state_id": p.TxEcosystemID, "tx_wallet_id": p.TxKeyID})
		return logger
	}
	if p.PrevBlock != nil {
		logger := log.WithFields(log.Fields{"prev_block_id": p.PrevBlock.BlockID, "prev_block_time": p.PrevBlock.Time, "prev_block_wallet_id": p.PrevBlock.KeyID, "prev_block_state_id": p.PrevBlock.EcosystemID, "prev_block_hash": p.PrevBlock.Hash, "prev_block_version": p.PrevBlock.Version, "tx_type": p.TxType, "tx_time": p.TxTime, "tx_state_id": p.TxEcosystemID, "tx_wallet_id": p.TxKeyID})
		return logger
	}
	logger := log.WithFields(log.Fields{"tx_type": p.TxType, "tx_time": p.TxTime, "tx_state_id": p.TxEcosystemID, "tx_wallet_id": p.TxKeyID})
	return logger
}

// ClearTmp deletes temporary files
func ClearTmp(blocks map[int64]string) {
	for _, tmpFileName := range blocks {
		os.Remove(tmpFileName)
	}
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
		found, err := qtx.GetByHash(searchedHash)
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

// InsertIntoBlockchain inserts a block into the blockchain
func InsertIntoBlockchain(transaction *model.DbTransaction, block *Block) error {

	// for local tests
	blockID := block.Header.BlockID
	if block.Header.BlockID == 1 {
		if *conf.StartBlockID != 0 {
			blockID = *conf.StartBlockID
		}
	}

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
		if rollbackTxBytes, err := json.Marshal(rollbackTx); err != nil {
			log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling rollback_tx to json")
			return err
		} else {
			buffer.Write(rollbackTxBytes)
		}
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
	err = b.Create(transaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating block")
		return err
	}

	err = autoupdate.TryUpdate(uint64(blockID))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.AutoupdateError, "error": err, "blockID": blockID}).Fatal("update for blockID")
		return err
	}
	return nil
}

// FormatBlockData returns formated block data
func (p *Parser) FormatBlockData() string {
	result := ""
	if p.BlockData != nil {
		v := reflect.ValueOf(*p.BlockData)
		typeOfT := v.Type()
		if typeOfT.Kind() == reflect.Ptr {
			typeOfT = typeOfT.Elem()
		}
		for i := 0; i < v.NumField(); i++ {
			name := typeOfT.Field(i).Name
			switch name {
			case "BlockId", "Time", "UserId", "Level":
				result += "[" + name + "] = " + fmt.Sprintf("%d\n", v.Field(i).Interface())
			case "Sign", "Hash", "HeadHash":
				result += "[" + name + "] = " + fmt.Sprintf("%x\n", v.Field(i).Interface())
			default:
				result += "[" + name + "] = " + fmt.Sprintf("%s\n", v.Field(i).Interface())
			}
		}
	}
	return result
}

// FormatTxMap returns the formated TxMap
func (p *Parser) FormatTxMap() string {
	result := ""
	for k, v := range p.TxMap {
		switch k {
		case "sign":
			result += "[" + k + "] = " + fmt.Sprintf("%x\n", v)
		default:
			result += "[" + k + "] = " + fmt.Sprintf("%s\n", v)
		}
	}
	return result
}

// ErrInfo returns the more detailed error
func (p *Parser) ErrInfo(verr interface{}) error {
	var err error
	switch verr.(type) {
	case error:
		err = verr.(error)
	case string:
		err = fmt.Errorf(verr.(string))
	}
	return fmt.Errorf("[ERROR] %s (%s)\n%s\n%s", err, utils.Caller(1), p.FormatBlockData(), p.FormatTxMap())
}

// BlockError writes the error of the transaction in the transactions_status table
func (p *Parser) BlockError(err error) {
	if len(p.TxHash) == 0 {
		return
	}
	errText := err.Error()
	if len(errText) > 255 {
		errText = errText[:255]
	}
	p.DeleteQueueTx(p.TxHash)
	ts := &model.TransactionStatus{}
	ts.SetError(errText, p.TxHash)
}

// AccessRights checks the access right by executing the condition value
func (p *Parser) AccessRights(condition string, iscondition bool) error {
	logger := p.GetLogger()
	sp := &model.StateParameter{}
	sp.SetTablePrefix(converter.Int64ToStr(p.TxSmart.EcosystemID))
	_, err := sp.Get(p.DbTransaction, condition)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting state parameter by name transaction")
		return err
	}
	conditions := sp.Value
	if iscondition {
		conditions = sp.Conditions
	}
	if len(conditions) > 0 {
		ret, err := p.SmartContract.EvalIf(conditions)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.EvalError, "error": err, "conditions": conditions}).Error("evaluating conditions")
			return err
		}
		if !ret {
			logger.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
			return fmt.Errorf(`Access denied`)
		}
	} else {
		logger.WithFields(log.Fields{"type": consts.EmptyObject, "conditions": condition}).Error("No condition in state_parameters")
		return fmt.Errorf(`There is not %s in state_parameters`, condition)
	}
	return nil
}

// AccessChange is changing access
func (p *Parser) AccessChange(table, name, global string, stateId int64) error {
	logger := p.GetLogger()
	prefix, err := GetTablePrefix(global, stateId)
	if err != nil {
		return err
	}
	var conditions string
	switch table {
	case "pages":
		page := &model.Page{}
		page.SetTablePrefix(prefix)
		if _, err := page.Get(name); err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page")
			return err
		}
		conditions = page.Conditions
	case "menus":
		menu := &model.Menu{}
		menu.SetTablePrefix(prefix)
		if _, err := menu.Get(name); err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting menu")
			return err
		}
		conditions = menu.Conditions
	}

	if len(conditions) > 0 {
		ret, err := p.SmartContract.EvalIf(conditions)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.EvalError, "error": err}).Error("evaluating conditions")
			return err
		}
		if !ret {
			log.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
			return fmt.Errorf(`Access denied`)
		}
	} else {
		log.WithFields(log.Fields{"type": consts.EmptyObject, "table": prefix + "_" + table}).Error("There is not conditions in")
		return fmt.Errorf(`There is not conditions in %s`, prefix+`_`+table)
	}
	return nil
}

func (p *Parser) getEGSPrice(name string) (decimal.Decimal, error) {
	logger := p.GetLogger()
	syspar := &model.SystemParameter{}
	fPrice, err := syspar.GetValueParameterByName("op_price", name)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting value parameter by name")
		return decimal.New(0, 0), p.ErrInfo(err)
	}
	if fPrice == nil {
		return decimal.New(0, 0), nil
	}
	p.TxCost = 0
	p.TxUsedCost, _ = decimal.NewFromString(*fPrice)
	systemParam := &model.SystemParameter{}
	_, err = systemParam.Get("fuel_rate")
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Fatal("getting system parameter")
	}
	fuelRate, err := decimal.NewFromString(systemParam.Value)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.ConversionError, "value": systemParam.Value}).Error("converting fuel rate system parameter from string to decimal")
		return decimal.New(0, 0), p.ErrInfo(err)
	}
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
		logger.Error("fuel rate is less than zero")
		return decimal.New(0, 0), fmt.Errorf(`fuel rate must be greater than 0`)
	}
	return p.TxUsedCost.Mul(fuelRate), nil
}

// CallContract calls the contract functions according to the specified flags
func (p *Parser) CallContract(flags int) (resultContract string, err error) {
	sc := smart.SmartContract{
		VDE:           false,
		Rollback:      true,
		SysUpdate:     false,
		VM:            smart.GetVM(false, 0),
		TxSmart:       *p.TxSmart,
		TxData:        p.TxData,
		TxContract:    p.TxContract,
		TxCost:        p.TxCost,
		TxUsedCost:    p.TxUsedCost,
		BlockData:     p.BlockData,
		TxHash:        p.TxHash,
		PublicKeys:    p.PublicKeys,
		DbTransaction: p.DbTransaction,
	}
	resultContract, err = sc.CallContract(flags)
	p.SysUpdate = sc.SysUpdate
	return
}
