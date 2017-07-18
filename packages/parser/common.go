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
	"database/sql"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	db "github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"github.com/op/go-logging"
	"github.com/shopspring/decimal"
)

var (
	log = logging.MustGetLogger("daemons")
)

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
		return "", err
	}
	stateIdStr := converter.Int64ToStr(stateId)
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
	case "DLTChangeHostVote":
		return &DLTChangeHostVoteParser{p, nil}, nil
	case "UpdFullNodes":
		return &UpdFullNodesParser{p, nil}, nil
	case "ChangeNodeKey":
		return &ChangeNodeKeyParser{p, nil}, nil
	case "NewState":
		return &NewStateParser{p, nil}, nil
	case "NewColumn":
		return &NewColumnParser{p, nil}, nil
	case "NewTable":
		return &NewTableParser{p, nil}, nil
	case "EditPage":
		return &EditPageParser{p, nil}, nil
	case "EditMenu":
		return &EditMenuParser{p, nil}, nil
	case "EditContract":
		return &EditContractParser{p, nil}, nil
	case "NewContract":
		return &NewContractParser{p, nil, nil}, nil
	case "EditColumn":
		return &EditColumnParser{p, nil}, nil
	case "EditTable":
		return &EditTableParser{p, nil}, nil
	case "EditStateParameters":
		return &EditStateParametersParser{p, nil}, nil
	case "NewStateParameters":
		return &NewStateParametersParser{p, nil}, nil
	case "NewPage":
		return &NewPageParser{p, nil}, nil
	case "NewMenu":
		return &NewMenuParser{p, nil}, nil
	case "ChangeNodeKeyDLT":
		return &ChangeNodeKeyDLTParser{p, nil}, nil
	case "AppendPage":
		return &AppendPageParser{p, nil}, nil
	case "RestoreAccessActive":
		return &RestoreAccessActiveParser{p, nil, "", 0}, nil
	case "RestoreAccessClose":
		return &RestoreAccessCloseParser{p, nil}, nil
	case "RestoreAccessRequest":
		return &RestoreAccessRequestParser{p, nil}, nil
	case "RestoreAccess":
		return &RestoreAccessParser{p, nil}, nil
	case "NewLang":
		return &NewLangParser{p, nil}, nil
	case "EditLang":
		return &EditLangParser{p, nil}, nil
	case "AppendMenu":
		return &AppendMenuParser{p, nil}, nil
	case "NewSign":
		return &NewSignParser{p, nil}, nil
	case "EditSign":
		return &EditSignParser{p, nil}, nil
	case "EditWallet":
		return &EditWalletParser{p, nil}, nil
	case "ActivateContract":
		return &ActivateContractParser{p, nil, ""}, nil
	case "NewAccount":
		return &NewAccountParser{p, nil}, nil
	}
	return nil, fmt.Errorf("Unknown txType: %s", txType)
}

type txMapsType struct {
	Int64   map[string]int64
	String  map[string]string
	Bytes   map[string][]byte
	Float64 map[string]float64
	Money   map[string]float64
	Decimal map[string]decimal.Decimal
}

// Parser is a structure for parsing transactions
type Parser struct {
	*db.DCDB
	TxMaps           *txMapsType
	TxMap            map[string][]byte
	TxMapS           map[string]string
	TxIds            int // count of transactions
	TxMapArr         []map[string][]byte
	TxMapsArr        []*txMapsType
	BlockData        *utils.BlockData
	PrevBlock        *utils.BlockData
	BinaryData       []byte
	TxBinaryData     []byte
	blockHashHex     []byte
	dataType         int
	blockHex         []byte
	CurrentBlockID   int64
	fullTxBinaryData []byte
	TxHash           string
	TxSlice          [][]byte
	MerkleRoot       []byte
	GoroutineName    string
	CurrentVersion   string
	MrklRoot         []byte
	PublicKeys       [][]byte
	TxUserID         int64
	TxCitizenID      int64
	TxWalletID       int64
	TxStateID        uint32
	TxStateIDStr     string
	TxTime           int64
	TxCost           int64           // Maximum cost of executing contract
	TxUsedCost       decimal.Decimal // Used cost of CPU resources
	nodePublicKey    []byte
	//	newPublicKeysHex [3][]byte
	TxPtr      interface{} // Pointer to the corresponding struct in consts/struct.go
	TxData     map[string]interface{}
	TxContract *smart.Contract
	TxVars     map[string]string
	AllPkeys   map[string]string
	States     map[int64]string
}

// ClearTmp deletes temporary files
func ClearTmp(blocks map[int64]string) {
	for _, tmpFileName := range blocks {
		os.Remove(tmpFileName)
	}
}

// GetBlockInfo returns BlockData structure
func (p *Parser) GetBlockInfo() *utils.BlockData {
	return &utils.BlockData{Hash: p.BlockData.Hash, Time: p.BlockData.Time, WalletID: p.BlockData.WalletID, StateID: p.BlockData.StateID, BlockID: p.BlockData.BlockID}
}

/*
func (p *Parser) limitRequest(vimit interface{}, txType string, vperiod interface{}) error {

	var limit int
	switch vimit.(type) {
	case string:
		limit = utils.StrToInt(vimit.(string))
	case int:
		limit = vimit.(int)
	case int64:
		limit = int(vimit.(int64))
	}

	var period int
	switch vperiod.(type) {
	case string:
		period = utils.StrToInt(vperiod.(string))
	case int:
		period = vperiod.(int)
	}

	time := utils.BytesToInt(p.TxMap["time"])
	num, err := p.Single("SELECT count(time) FROM rb_time_"+txType+" WHERE user_id = ? AND time > ?", p.TxUserID, (time - period)).Int()
	if err != nil {
		return err
	}
	if num >= limit {
		return utils.ErrInfo(fmt.Errorf("[limit_requests] rb_time_%v %v >= %v", txType, num, limit))
	} else {
		err := p.ExecSQL("INSERT INTO rb_time_"+txType+" (user_id, time) VALUES (?, ?)", p.TxUserID, time)
		if err != nil {
			return err
		}
	}
	return nil
}*/

func (p *Parser) dataPre() {
	hash, err := crypto.DoubleHash(p.BinaryData)
	if err != nil {
		log.Fatal(err)
	}
	p.blockHashHex = converter.BinToHex(hash)

	p.blockHex = converter.BinToHex(p.BinaryData)
	// определим тип данных
	// define the data type
	p.dataType = int(converter.BinToDec(converter.BytesShift(&p.BinaryData, 1)))
	log.Debug("dataType", p.dataType)
}

// CheckLogTx checks if this transaction exists
// Это защита от dos, когда одну транзакцию можно было бы послать миллион раз,
// This is protection against dos, when one transaction could be sent a million times
// и она каждый раз успешно проходила бы фронтальную проверку
// And it would have successfully passed a frontal test
func (p *Parser) CheckLogTx(txBinary []byte, transactions, txQueue bool) error {
	searchedHash, err := crypto.Hash(txBinary)
	if err != nil {
		log.Fatal(err)
	}
	searchedHash = converter.BinToHex(searchedHash)
	hash, err := p.Single(`SELECT hash FROM log_transactions WHERE hex(hash) = ?`, searchedHash).String()
	log.Debug("SELECT hash FROM log_transactions WHERE hex(hash) = %s", searchedHash)
	if err != nil {
		log.Error("%s", utils.ErrInfo(err))
		return utils.ErrInfo(err)
	}
	log.Debug("hash %x", hash)
	if len(hash) > 0 {
		return utils.ErrInfo(fmt.Errorf("double tx in log_transactions %s", searchedHash))
	}

	if transactions {
		// проверим, нет ли у нас такой тр-ии
		// check whether we have such a transaction
		exists, err := p.Single("SELECT count(hash) FROM transactions WHERE hex(hash) = ? and verified = 1", searchedHash).Int64()
		if err != nil {
			log.Error("%s", utils.ErrInfo(err))
			return utils.ErrInfo(err)
		}
		if exists > 0 {
			return utils.ErrInfo(fmt.Errorf("double tx in transactions %s", searchedHash))
		}
	}

	if txQueue {
		// проверим, нет ли у нас такой тр-ии
		// check whether we have such a transaction
		exists, err := p.Single("SELECT count(hash) FROM queue_tx WHERE hex(hash) = ?", searchedHash).Int64()
		if err != nil {
			log.Error("%s", utils.ErrInfo(err))
			return utils.ErrInfo(err)
		}
		if exists > 0 {
			return utils.ErrInfo(fmt.Errorf("double tx in queue_tx %s", searchedHash))
		}
	}

	return nil
}

// GetInfoBlock returns the latest block
func (p *Parser) GetInfoBlock() error {

	// последний успешно записанный блок
	// the last successfully recorded block
	p.PrevBlock = new(utils.BlockData)
	var q string
	if p.ConfigIni["db_type"] == "mysql" || p.ConfigIni["db_type"] == "sqlite" {
		q = "SELECT LOWER(HEX(hash)) as hash, block_id, time FROM info_block"
	} else if p.ConfigIni["db_type"] == "postgresql" {
		q = "SELECT encode(hash, 'HEX')  as hash, block_id, time FROM info_block"
	}
	err := p.QueryRow(q).Scan(&p.PrevBlock.Hash, &p.PrevBlock.BlockID, &p.PrevBlock.Time)

	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	return nil
}

// InsertIntoBlockchain inserts a block into the blockchain
func (p *Parser) InsertIntoBlockchain() error {
	//var mutex = &sync.Mutex{}
	// для локальных тестов
	// for local tests
	if p.BlockData.BlockID == 1 {
		if *utils.StartBlockID != 0 {
			p.BlockData.BlockID = *utils.StartBlockID
		}
	}
	//mutex.Lock()
	// пишем в цепочку блоков
	// record into the block chain
	err := p.ExecSQL("DELETE FROM block_chain WHERE id = ?", p.BlockData.BlockID)
	if err != nil {
		return err
	}
	err = p.ExecSQL("INSERT INTO block_chain (id, hash, data, state_id, wallet_id, time, tx) VALUES (?, [hex], [hex], ?, ?, ?, ?)",
		p.BlockData.BlockID, p.BlockData.Hash, p.blockHex, p.BlockData.StateID, p.BlockData.WalletID, p.BlockData.Time, p.TxIds)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//mutex.Unlock()
	return nil
}

// старое
// the old
/*func (p *Parser) GetTxMap(fields []string) (map[string][]byte, error) {
	if len(p.TxSlice) != len(fields)+4 {
		return nil, fmt.Errorf("bad transaction_array %d != %d (type=%d)", len(p.TxSlice), len(fields)+4, p.TxSlice[0])
	}
	TxMap := make(map[string][]byte)
	TxMap["hash"] = p.TxSlice[0]
	TxMap["type"] = p.TxSlice[1]
	TxMap["time"] = p.TxSlice[2]
	TxMap["user_id"] = p.TxSlice[3]
	for i, field := range fields {
		TxMap[field] = p.TxSlice[i+4]
	}
	p.TxUserID = utils.BytesToInt64(TxMap["user_id"])
	p.TxTime = utils.BytesToInt64(TxMap["time"])
	p.PublicKeys = nil
	//log.Debug("TxMap", TxMap)
	//log.Debug("TxMap[hash]", TxMap["hash"])
	//log.Debug("p.TxSlice[0]", p.TxSlice[0])
	return TxMap, nil
}*/

func (p *Parser) CheckInputData(data map[string][]interface{}) error {
	for k, list := range data {
		for _, v := range list {
			fmt.Println("v==", v, k)
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
	wallet := p.TxWalletID
	if wallet == 0 {
		wallet = p.TxCitizenID
	}
	// получим сумму на кошельке юзера
	// recieve the amount on the user's wallet
	strAmount, err := p.Single(`SELECT amount FROM dlt_wallets WHERE wallet_id = ?`, wallet).String()
	if err != nil {
		return err
	}
	totalAmount, _ := decimal.NewFromString(strAmount)

	amountAndCommission := amount
	amountAndCommission.Add(commission)
	if totalAmount.Cmp(amountAndCommission) < 0 {
		return fmt.Errorf("%v < %v)", totalAmount, amountAndCommission)
	}
	return nil
}

// CheckTableExists checks if the table exists
func (p *Parser) CheckTableExists(table string) (bool, error) {
	var q string
	switch p.ConfigIni["db_type"] {
	case "sqlite":
		q = `SELECT name FROM sqlite_master WHERE type='table' AND name='` + table + `';`
	case "postgresql":
		q = `SELECT relname FROM pg_class WHERE relname = '` + table + `';`
	case "mysql":
		q = `SHOW TABLES LIKE '` + table + `'`
	}
	exists, err := p.Single(q).Int64()
	if err != nil {
		return false, err
	}
	if exists > 0 {
		return true, nil
	}

	return false, nil
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
	p.DeleteQueueTx([]byte(p.TxHash))
	log.Debug("UPDATE transactions_status SET error = %s WHERE hex(hash) = %x", errText, p.TxHash)
	p.ExecSQL("UPDATE transactions_status SET error = ? WHERE hex(hash) = ?", errText, p.TxHash)
}

// AccessRights checks the access right by executing the condition value
func (p *Parser) AccessRights(condition string, iscondition bool) error {
	param := `value`
	if iscondition {
		param = `conditions`
	}
	conditions, err := p.Single(`SELECT `+param+` FROM "`+converter.Int64ToStr(int64(p.TxStateID))+`_state_parameters" WHERE name = ?`,
		condition).String()
	if err != nil {
		return err
	}
	if len(conditions) > 0 {
		ret, err := p.EvalIf(conditions)
		if err != nil {
			return err
		}
		if !ret {
			return fmt.Errorf(`Access denied`)
		}
	} else {
		return fmt.Errorf(`There is not %s in state_parameters`, condition)
	}
	return nil
}

// AccessTable checks the access right to the table
func (p *Parser) AccessTable(table, action string) error {

	//	prefix := utils.Int64ToStr(int64(p.TxStateID))
	govAccount, _ := template.StateParam(int64(p.TxStateID), `gov_account`)
	if table == `dlt_wallets` && p.TxContract != nil && p.TxCitizenID == converter.StrToInt64(govAccount) {
		return nil
	}

	if isCustom, err := p.IsCustomTable(table); err != nil {
		return err // table != ... временно оставлено для совместимости. После переделки new_state убрать
		// table != ... is left for compatibility temporarily. Remove new_state after rebuilding.
	} else if !isCustom && !strings.HasSuffix(table, `_citizenship_requests`) {
		return fmt.Errorf(table + ` is not a custom table`)
	}
	prefix := table[:strings.IndexByte(table, '_')]

	/*	if p.TxStateID == 0 {
		return nil
	}*/

	tablePermission, err := p.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions) as data WHERE name = ?`, "key", "value", table)
	if err != nil {
		return err
	}
	if len(tablePermission[action]) > 0 {
		ret, err := p.EvalIf(tablePermission[action])
		if err != nil {
			return err
		}
		if !ret {
			return fmt.Errorf(`Access denied`)
		}
	}
	return nil
}

// AccessColumns checks access rights to the columns
func (p *Parser) AccessColumns(table string, columns []string) error {

	//prefix := utils.Int64ToStr(int64(p.TxStateID))

	if isCustom, err := p.IsCustomTable(table); err != nil {
		return err // table != ... временно оставлено для совместимости. После переделки new_state убрать // table != ... is left for compatibility temporarily. Remove new_state after rebuilding
	} else if !isCustom && !strings.HasSuffix(table, `_citizenship_requests`) {
		return fmt.Errorf(table + ` is not a custom table`)
	}
	prefix := table[:strings.IndexByte(table, '_')]
	/*	if p.TxStateID == 0 {
		return nil
	}*/

	columnsAndPermissions, err := p.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions->'update') as data WHERE name = ?`,
		"key", "value", table)
	if err != nil {
		return err
	}
	for _, col := range columns {
		if cond, ok := columnsAndPermissions[col]; ok && len(cond) > 0 {
			ret, err := p.EvalIf(cond)
			if err != nil {
				return err
			}
			if !ret {
				return fmt.Errorf(`Access denied`)
			}
		}
	}
	return nil
}

func (p *Parser) AccessChange(table, name, global string, stateId int64) error {
	prefix, err := GetTablePrefix(global, stateId)
	if err != nil {
		return err
	}
	conditions, err := p.Single(`SELECT conditions FROM "`+prefix+`_`+table+`" WHERE name = ?`, name).String()
	if err != nil {
		return err
	}

	if len(conditions) > 0 {
		ret, err := p.EvalIf(conditions)
		if err != nil {
			return err
		}
		if !ret {
			return fmt.Errorf(`Access denied`)
		}
	} else {
		return fmt.Errorf(`There is not conditions in %s`, prefix+`_`+table)
	}
	return nil
}

func (p *Parser) getEGSPrice(name string) (decimal.Decimal, error) {
	fPrice, err := p.Single(`SELECT value->'`+name+`' FROM system_parameters WHERE name = ?`, "op_price").String()
	if err != nil {
		return decimal.New(0, 0), p.ErrInfo(err)
	}
	p.TxCost = 0
	p.TxUsedCost, _ = decimal.NewFromString(fPrice)
	fuelRate := p.GetFuel()
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
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

// GetContractLimit returns the default maximal cost of contract
func (p *Parser) GetContractLimit() (ret int64) {
	// default maximum cost of F
	p.TxCost = script.CostDefault // ret * fuel
	return p.TxCost
}

func (p *Parser) payFPrice() error {
	var (
		fromID int64
		err    error
	)
	//return nil
	toID := p.BlockData.WalletID // account of node
	fuel := p.GetFuel()
	if fuel.Cmp(decimal.New(0, 0)) <= 0 {
		return fmt.Errorf(`fuel rate must be greater than 0`)
	}

	if p.TxCost == 0 { // embedded transaction
		fromID = p.TxWalletID
		if fromID == 0 {
			fromID = p.TxCitizenID
		}
	} else { // contract
		if p.TxStateID > 0 && p.TxCitizenID != 0 && p.TxContract != nil {
			//fromID = p.TxContract.TxGovAccount
			fromID = converter.StrToInt64(StateVal(p, `gov_account`))
		} else {
			// списываем напрямую с dlt_wallets у юзера
			// write directly from dlt_wallets of user
			fromID = p.TxWalletID
		}
	}
	egs := p.TxUsedCost.Mul(fuel)
	fmt.Printf("Pay fuel=%v fromID=%d toID=%d cost=%v egs=%v", fuel, fromID, toID, p.TxUsedCost, egs)
	if egs.Cmp(decimal.New(0, 0)) == 0 { // Is it possible to pay nothing?
		return nil
	}
	var amount string
	if amount, err = p.Single(`select amount from dlt_wallets where wallet_id=?`, fromID).String(); err != nil {
		return err
	}
	damount, err := decimal.NewFromString(amount)
	if err != nil {
		return err
	}
	if damount.Cmp(egs) < 0 {
		egs = damount
	}
	commission := egs.Mul(decimal.New(3, 0)).Div(decimal.New(100, 0)).Floor()
	if _, _, err := p.selectiveLoggingAndUpd([]string{`-amount`}, []interface{}{egs}, `dlt_wallets`, []string{`wallet_id`},
		[]string{converter.Int64ToStr(fromID)}, true); err != nil {
		return err
	}
	if _, _, err := p.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{egs.Sub(commission)}, `dlt_wallets`, []string{`wallet_id`},
		[]string{converter.Int64ToStr(toID)}, true); err != nil {
		return err
	}
	if _, _, err := p.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{commission}, `dlt_wallets`, []string{`wallet_id`},
		[]string{converter.Int64ToStr(consts.COMMISSION_WALLET)}, true); err != nil {
		return err
	}
	fmt.Printf(" Paid commission %v\r\n", commission)
	return nil
}
