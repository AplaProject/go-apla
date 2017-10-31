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
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/templatev2"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// GetTxTypeAndUserID returns tx type, wallet and citizen id from the block data
func GetTxTypeAndUserID(binaryBlock []byte) (txType int64, walletID int64, citizenID int64) {
	tmp := binaryBlock[:]
	txType = converter.BinToDecBytesShift(&binaryBlock, 1)
	if consts.IsStruct(int(txType)) {
		var txHead consts.TxHeader
		converter.BinUnmarshal(&tmp, &txHead)
		walletID = txHead.WalletID
		citizenID = txHead.CitizenID
	}
	return
}

func GetBlockDataFromBlockChain(blockID int64) (*utils.BlockData, error) {
	BlockData := new(utils.BlockData)
	block := &model.Block{}
	err := block.GetBlock(blockID)
	if err != nil {
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

func GetNodePublicKeyWalletOrCB(walletID, stateID int64) ([]byte, error) {
	if walletID != 0 {
		node := syspar.GetNode(walletID)
		if node != nil {
			return node.Public, nil
		}
	}
	return nil, fmt.Errorf(`unknown node %d`, walletID)
}

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

func IsCustomTable(table string) (isCustom bool, err error) {
	if table[0] >= '0' && table[0] <= '9' {
		if off := strings.IndexByte(table, '_'); off > 0 {
			prefix := table[:off]
			tables := &model.Table{}
			tables.SetTablePrefix(prefix)
			found, err := tables.Get(table[off+1:])
			if err != nil {
				log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting table")
				return false, err
			}
			if found {
				return true, nil
			}
		}
	}
	return false, nil
}

func IsState(transaction *model.DbTransaction, country string) (int64, error) {
	ids, err := model.GetAllSystemStatesIDs()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get all system states ids")
		return 0, err
	}
	for _, id := range ids {
		sp := &model.StateParameter{}
		sp.SetTablePrefix(strconv.Itoa(int(id)))
		err = sp.GetByNameTransaction(transaction, "state_name")
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

func init() {
	flag.Parse()
}

type ParserInterface interface {
	Init() error
	Validate() error
	Action() error
	Rollback() error
	Header() *tx.Header
}

func GetTablePrefix(global string, stateId int64) (string, error) {
	globalInt, err := strconv.Atoi(global)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ConvertionError}).Error("converting global to int")
		return "", err
	}
	stateIdStr := strconv.Itoa(int(stateId))
	if globalInt == 1 {
		return "global", nil
	}
	return stateIdStr, nil
}

func GetParser(p *Parser, txType string) (ParserInterface, error) {
	switch txType {
	case "FirstBlock":
		return &FirstBlockParser{p}, nil
	case "DLTTransfer":
		return &DLTTransferParser{p, nil}, nil
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

	TxBinaryData  []byte // transaction binary data
	TxFullData    []byte // full transaction, with type and data
	TxHash        []byte
	TxSlice       [][]byte
	TxMap         map[string][]byte
	TxIds         int // count of transactions
	TxUserID      int64
	TxCitizenID   int64
	TxWalletID    int64
	TxStateID     uint32
	TxStateIDStr  string
	TxTime        int64
	TxType        int64
	TxCost        int64           // Maximum cost of executing contract
	TxUsedCost    decimal.Decimal // Used cost of CPU resources
	TxPtr         interface{}     // Pointer to the corresponding struct in consts/struct.go
	TxData        map[string]interface{}
	TxSmart       *tx.SmartContract
	TxContract    *smart.Contract
	TxHeader      *tx.Header
	txParser      ParserInterface
	DbTransaction *model.DbTransaction

	AllPkeys map[string]string
}

func (p Parser) GetLogger() *log.Entry {
	if p.BlockData != nil && p.PrevBlock != nil {
		logger := log.WithFields(log.Fields{"block_id": p.BlockData.BlockID, "block_time": p.BlockData.Time, "block_wallet_id": p.BlockData.WalletID, "block_state_id": p.BlockData.StateID, "block_hash": p.BlockData.Hash, "block_version": p.BlockData.Version, "prev_block_id": p.PrevBlock.BlockID, "prev_block_time": p.PrevBlock.Time, "prev_block_wallet_id": p.PrevBlock.WalletID, "prev_block_state_id": p.PrevBlock.StateID, "prev_block_hash": p.PrevBlock.Hash, "prev_block_version": p.PrevBlock.Version, "tx_type": p.TxType, "tx_time": p.TxTime, "tx_state_id": p.TxStateID, "tx_wallet_id": p.TxWalletID})
		return logger
	}
	if p.BlockData != nil {
		logger := log.WithFields(log.Fields{"block_id": p.BlockData.BlockID, "block_time": p.BlockData.Time, "block_wallet_id": p.BlockData.WalletID, "block_state_id": p.BlockData.StateID, "block_hash": p.BlockData.Hash, "block_version": p.BlockData.Version, "tx_type": p.TxType, "tx_time": p.TxTime, "tx_state_id": p.TxStateID, "tx_wallet_id": p.TxWalletID})
		return logger
	}
	if p.PrevBlock != nil {
		logger := log.WithFields(log.Fields{"prev_block_id": p.PrevBlock.BlockID, "prev_block_time": p.PrevBlock.Time, "prev_block_wallet_id": p.PrevBlock.WalletID, "prev_block_state_id": p.PrevBlock.StateID, "prev_block_hash": p.PrevBlock.Hash, "prev_block_version": p.PrevBlock.Version, "tx_type": p.TxType, "tx_time": p.TxTime, "tx_state_id": p.TxStateID, "tx_wallet_id": p.TxWalletID})
		return logger
	}
	logger := log.WithFields(log.Fields{"tx_type": p.TxType, "tx_time": p.TxTime, "tx_state_id": p.TxStateID, "tx_wallet_id": p.TxWalletID})
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
		err := tx.GetVerified(searchedHash)
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
		if *utils.StartBlockID != 0 {
			blockID = *utils.StartBlockID
		}
	}

	// record into the block chain
	bl := &model.Block{}
	err := bl.DeleteById(transaction, blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting block by id")
		return err
	}
	b := &model.Block{
		ID:       blockID,
		Hash:     block.Header.Hash,
		Data:     block.BinData,
		StateID:  block.Header.StateID,
		WalletID: block.Header.WalletID,
		Time:     block.Header.Time,
		Tx:       int32(len(block.Parsers)),
	}
	err = b.Create(transaction)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating block")
		return err
	}
	return nil
}

func (p *Parser) CheckInputData(data map[string][]interface{}) error {
	for k, list := range data {
		for _, v := range list {
			if !utils.CheckInputData(v, k) {
				return fmt.Errorf("incorrect %s: %s", v, k)
			}
		}
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

func (p *Parser) checkSenderDLT(amount, commission decimal.Decimal) error {
	logger := p.GetLogger()
	walletID := p.TxWalletID
	if walletID == 0 {
		walletID = p.TxCitizenID
	}

	wallet := &model.DltWallet{}
	err := wallet.GetWalletTransaction(p.DbTransaction, walletID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting wallet transaction")
		return err
	}
	amountAndCommission := amount
	amountAndCommission.Add(commission)
	wltAmount, err := decimal.NewFromString(wallet.Amount)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": wallet.Amount}).Error("convertion wallet amount to decimal from string")
		return err
	}
	if wltAmount.Cmp(amountAndCommission) < 0 {
		logger.Error("wallet amount is less than amount and commisssion")
		return fmt.Errorf("%v < %v)", wallet.Amount, amountAndCommission)
	}
	return nil
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
	sp.SetTablePrefix(p.TxStateIDStr)
	err := sp.GetByNameTransaction(p.DbTransaction, condition)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting state parameter by name transaction")
		return err
	}
	conditions := sp.Value
	if iscondition {
		conditions = sp.Conditions
	}
	if len(conditions) > 0 {
		ret, err := p.EvalIf(conditions)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.EvalError, "error": err}).Error("Evaluationg conditions")
			return err
		}
		if !ret {
			logger.WithFields(log.Fields{"type": consts.AccessDenied, "conditions": conditions}).Error("Access denied")
			return fmt.Errorf(`Access denied`)
		}
	} else {
		return fmt.Errorf(`There is not %s in state_parameters`, condition)
	}
	return nil
}

// AccessTable checks the access right to the table
func (p *Parser) AccessTable(table, action string) error {
	logger := p.GetLogger()
	govAccount, _ := templatev2.StateParam(int64(p.TxStateID), `founder_account`)
	if table == fmt.Sprintf(`%d_parameters`, p.TxStateID) {
		govAccountInt, err := strconv.ParseInt(govAccount, 10, 64)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("parsing gov account to int")
			return err
		}
		if p.TxContract != nil && p.TxCitizenID == govAccountInt {
			return nil
		} else {
			logger.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
			return fmt.Errorf(`Access denied`)
		}
	}

	if isCustom, err := IsCustomTable(table); err != nil {
		return err
		// TODO: table != ... is left for compatibility temporarily. Remove it
	} else if !isCustom && !strings.HasSuffix(table, `_citizenship_requests`) {
		logger.WithFields(log.Fields{"table": table, "type": consts.InvalidObject}).Error("is not custom table")
		return fmt.Errorf(table + ` is not a custom table`)
	}
	prefix := table[:strings.IndexByte(table, '_')]
	tables := &model.Table{}
	tables.SetTablePrefix(prefix)
	tablePermission, err := tables.GetPermissions(table[strings.IndexByte(table, '_')+1:], "")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting table permissions")
		return err
	}
	if len(tablePermission[action]) > 0 {
		ret, err := p.EvalIf(tablePermission[action])
		if err != nil {
			logger.WithFields(log.Fields{"action": action, "permissions": tablePermission[action], "error": err, "type": consts.EvalError}).Error("evaluating table permissions for action")
			return err
		}
		if !ret {
			logger.WithFields(log.Fields{"action": action, "permissions": tablePermission[action], "type": consts.EvalError}).Error("access denied")
			return fmt.Errorf(`Access denied`)
		}
	}
	return nil
}

// AccessColumns checks access rights to the columns
func (p *Parser) AccessColumns(table string, columns []string) error {
	logger := p.GetLogger()
	if table == fmt.Sprintf(`%d_parameters`, p.TxStateID) {
		govAccount, _ := templatev2.StateParam(int64(p.TxStateID), `founder_account`)
		govAccountInt, err := strconv.ParseInt(govAccount, 10, 64)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("parsing gov account to int")
			return err
		}
		if p.TxContract != nil && p.TxCitizenID == govAccountInt {
			return nil
		}
		logger.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access Denied")
		return fmt.Errorf(`Access denied`)
	}
	if isCustom, err := IsCustomTable(table); err != nil {
		return err
	} else if !isCustom && !strings.HasSuffix(table, `_parameters`) {
		logger.WithFields(log.Fields{"table": table, "type": consts.InvalidObject}).Error("is not custom table")
		return fmt.Errorf(table + ` is not a custom table`)
	}
	prefix := table[:strings.IndexByte(table, '_')]
	tables := &model.Table{}
	tables.SetTablePrefix(prefix)
	columnsAndPermissions, err := tables.GetColumns(table, "")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting table columns")
		return err
	}
	for _, col := range columns {
		var (
			cond string
			ok   bool
		)
		cond, ok = columnsAndPermissions[converter.Sanitize(col, ``)]
		if !ok {
			cond, ok = columnsAndPermissions[`*`]
		}
		if ok && len(cond) > 0 {
			ret, err := p.EvalIf(cond)
			if err != nil {
				logger.WithFields(log.Fields{"condition": cond, "column": col, "type": consts.EvalError}).Error("evaluating condition")
				return err
			}
			if !ret {
				logger.WithFields(log.Fields{"condition": cond, "column": col, "type": consts.AccessDenied}).Error("action denied")
				return fmt.Errorf(`Access denied`)
			}
		}
	}
	return nil
}

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
		if err := page.Get(name); err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page")
			return err
		}
		conditions = page.Conditions
	case "menus":
		menu := &model.Menu{}
		menu.SetTablePrefix(prefix)
		if err := menu.Get(name); err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting menu")
			return err
		}
		conditions = menu.Conditions
	}

	if len(conditions) > 0 {
		ret, err := p.EvalIf(conditions)
		if err != nil {
			logger.WithFields(log.Fields{"conditions": conditions, "type": consts.EvalError}).Error("evaluating conditions")
			return err
		}
		if !ret {
			logger.WithFields(log.Fields{"conditions": conditions, "type": consts.AccessDenied}).Error("access denied")
			return fmt.Errorf(`Access denied`)
		}
	} else {
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
	err = systemParam.Get("fuel_rate")
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("getting system parameter")
	}
	fuelRate, err := decimal.NewFromString(systemParam.Value)
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.ConvertionError, "value": systemParam.Value}).Error("converting fuel rate system parameter from string to decimal")
		return decimal.New(0, 0), p.ErrInfo(err)
	}
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
		logger.Error("fuel rate is less than zero")
		return decimal.New(0, 0), fmt.Errorf(`fuel rate must be greater than 0`)
	}
	return p.TxUsedCost.Mul(fuelRate), nil
}

func (p *Parser) checkPrice(name string) error {
	EGSPrice, err := p.getEGSPrice(name)
	if err != nil {
		return err
	}
	// Is there a correct amount on the wallet?
	err = p.checkSenderDLT(EGSPrice, decimal.New(0, 0))
	if err != nil {
		return err
	}
	return nil
}
