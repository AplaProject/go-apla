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

	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
	"github.com/shopspring/decimal"
)

var (
	log = logging.MustGetLogger("daemons")
)

func init() {
	flag.Parse()
}

type txMapsType struct {
	Int64   map[string]int64
	String  map[string]string
	Bytes   map[string][]byte
	Float64 map[string]float64
	Money   map[string]float64
	Decimal map[string]decimal.Decimal
}
type Parser struct {
	*utils.DCDB
	TxMaps           *txMapsType
	TxMap            map[string][]byte
	TxMapS           map[string]string
	TxIds            int // count of transactions
	TxMapArr         []map[string][]byte
	TxMapsArr        []*txMapsType
	BlockData        *utils.BlockData
	PrevBlock        *utils.BlockData
	BinaryData       []byte
	blockHashHex     []byte
	dataType         int
	blockHex         []byte
	CurrentBlockId   int64
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
	nodePublicKey    []byte
	newPublicKeysHex [3][]byte
	TxPtr            interface{} // Pointer to the corresponding struct in consts/struct.go
	TxData           map[string]interface{}
	TxContract       *smart.Contract
	TxVars           map[string]string
	AllPkeys         map[string]string
	States           map[int64]string
}

func ClearTmp(blocks map[int64]string) {
	for _, tmpFileName := range blocks {
		os.Remove(tmpFileName)
	}
}

func (p *Parser) GetBlockInfo() *utils.BlockData {
	return &utils.BlockData{Hash: p.BlockData.Hash, Time: p.BlockData.Time, WalletId: p.BlockData.WalletId, CBID: p.BlockData.CBID, BlockId: p.BlockData.BlockId}
}

func (p *Parser) limitRequest(limit_ interface{}, txType string, period_ interface{}) error {

	var limit int
	switch limit_.(type) {
	case string:
		limit = utils.StrToInt(limit_.(string))
	case int:
		limit = limit_.(int)
	case int64:
		limit = int(limit_.(int64))
	}

	var period int
	switch period_.(type) {
	case string:
		period = utils.StrToInt(period_.(string))
	case int:
		period = period_.(int)
	}

	time := utils.BytesToInt(p.TxMap["time"])
	num, err := p.Single("SELECT count(time) FROM rb_time_"+txType+" WHERE user_id = ? AND time > ?", p.TxUserID, (time - period)).Int()
	if err != nil {
		return err
	}
	if num >= limit {
		return utils.ErrInfo(fmt.Errorf("[limit_requests] rb_time_%v %v >= %v", txType, num, limit))
	} else {
		err := p.ExecSql("INSERT INTO rb_time_"+txType+" (user_id, time) VALUES (?, ?)", p.TxUserID, time)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) dataPre() {
	p.blockHashHex = utils.DSha256(p.BinaryData)
	p.blockHex = utils.BinToHex(p.BinaryData)
	// определим тип данных
	p.dataType = int(utils.BinToDec(utils.BytesShift(&p.BinaryData, 1)))
	log.Debug("dataType", p.dataType)
}

// Это защита от dos, когда одну транзакцию можно было бы послать миллион раз,
// и она каждый раз успешно проходила бы фронтальную проверку
func (p *Parser) CheckLogTx(tx_binary []byte, transactions, queue_tx bool) error {
	hash, err := p.Single(`SELECT hash FROM log_transactions WHERE hex(hash) = ?`, utils.Md5(tx_binary)).String()
	log.Debug("SELECT hash FROM log_transactions WHERE hex(hash) = %s", utils.Md5(tx_binary))
	if err != nil {
		log.Error("%s", utils.ErrInfo(err))
		return utils.ErrInfo(err)
	}
	log.Debug("hash %x", hash)
	if len(hash) > 0 {
		return utils.ErrInfo(fmt.Errorf("double tx in log_transactions %s", utils.Md5(tx_binary)))
	}

	if transactions {
		// проверим, нет ли у нас такой тр-ии
		exists, err := p.Single("SELECT count(hash) FROM transactions WHERE hex(hash) = ?", utils.Md5(tx_binary)).Int64()
		if err != nil {
			log.Error("%s", utils.ErrInfo(err))
			return utils.ErrInfo(err)
		}
		if exists > 0 {
			return utils.ErrInfo(fmt.Errorf("double tx in transactions %s", utils.Md5(tx_binary)))
		}
	}

	if queue_tx {
		// проверим, нет ли у нас такой тр-ии
		exists, err := p.Single("SELECT count(hash) FROM queue_tx WHERE hex(hash) = ?", utils.Md5(tx_binary)).Int64()
		if err != nil {
			log.Error("%s", utils.ErrInfo(err))
			return utils.ErrInfo(err)
		}
		if exists > 0 {
			return utils.ErrInfo(fmt.Errorf("double tx in queue_tx %s", utils.Md5(tx_binary)))
		}
	}

	return nil
}

func (p *Parser) GetInfoBlock() error {

	// последний успешно записанный блок
	p.PrevBlock = new(utils.BlockData)
	var q string
	if p.ConfigIni["db_type"] == "mysql" || p.ConfigIni["db_type"] == "sqlite" {
		q = "SELECT LOWER(HEX(hash)) as hash, block_id, time FROM info_block"
	} else if p.ConfigIni["db_type"] == "postgresql" {
		q = "SELECT encode(hash, 'HEX')  as hash, block_id, time FROM info_block"
	}
	err := p.QueryRow(q).Scan(&p.PrevBlock.Hash, &p.PrevBlock.BlockId, &p.PrevBlock.Time)

	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) InsertIntoBlockchain() error {
	//var mutex = &sync.Mutex{}
	// для локальных тестов
	if p.BlockData.BlockId == 1 {
		if *utils.StartBlockId != 0 {
			p.BlockData.BlockId = *utils.StartBlockId
		}
	}
	//mutex.Lock()
	// пишем в цепочку блоков
	err := p.ExecSql("DELETE FROM block_chain WHERE id = ?", p.BlockData.BlockId)
	if err != nil {
		return err
	}
	err = p.ExecSql("INSERT INTO block_chain (id, hash, data, state_id, wallet_id, time, tx) VALUES (?, [hex], [hex], ?, ?, ?, ?)",
		p.BlockData.BlockId, p.BlockData.Hash, p.blockHex, p.BlockData.CBID, p.BlockData.WalletId, p.BlockData.Time, p.TxIds)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//mutex.Unlock()
	return nil
}

// старое
func (p *Parser) GetTxMap(fields []string) (map[string][]byte, error) {
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
}

func (p *Parser) CheckInputData(data map[string]string) error {
	for k, v := range data {
		if !utils.CheckInputData(p.TxMap[k], v) {
			return fmt.Errorf("incorrect " + k + "(" + string(p.TxMap[k]) + " : " + v + ")")
		}
	}
	return nil
}

func (p *Parser) limitRequestsRollback(txType string) error {
	time := p.TxMap["time"]
	if p.ConfigIni["db_type"] == "mysql" {
		return p.ExecSql("DELETE FROM rb_time_"+txType+" WHERE user_id = ? AND time = ? LIMIT 1", p.TxUserID, time)
	} else if p.ConfigIni["db_type"] == "postgresql" {
		return p.ExecSql("DELETE FROM rb_time_"+txType+" WHERE ctid IN (SELECT ctid FROM rb_time_"+txType+" WHERE  user_id = ? AND time = ? LIMIT 1)", p.TxUserID, time)
	} else {
		return p.ExecSql("DELETE FROM rb_time_"+txType+" WHERE id IN (SELECT id FROM rb_time_"+txType+" WHERE  user_id = ? AND time = ? LIMIT 1)", p.TxUserID, time)
	}
	return nil
}

func arrayIntersect(arr1, arr2 map[int]int) bool {
	for _, v := range arr1 {
		for _, v2 := range arr2 {
			if v == v2 {
				return true
			}
		}
	}
	return false
}

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

func (p *Parser) ErrInfo(err_ interface{}) error {
	var err error
	switch err_.(type) {
	case error:
		err = err_.(error)
	case string:
		err = fmt.Errorf(err_.(string))
	}
	return fmt.Errorf("[ERROR] %s (%s)\n%s\n%s", err, utils.Caller(1), p.FormatBlockData(), p.FormatTxMap())
}

func (p *Parser) limitRequestsMoneyOrdersRollback() error {
	err := p.ExecSql("DELETE FROM rb_time_money_orders WHERE hex(tx_hash) = ?", p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) getMyNodeCommission(currencyId, userId int64, amount float64) (float64, error) {
	return consts.COMMISSION, nil

}

func (p *Parser) checkSenderDLT(amount, commission int64) error {

	// получим сумму на кошельке юзера
	totalAmount, err := p.Single(`SELECT amount FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	amountAndCommission := amount + commission
	if totalAmount < amountAndCommission {
		return p.ErrInfo(fmt.Sprintf("%f < %f)", totalAmount, amountAndCommission))
	}
	return nil
}

func (p *Parser) MyTable(table, id_column string, id int64, ret_column string) (int64, error) {
	if utils.CheckInputData(table, "string") || utils.CheckInputData(ret_column, "string") {
		return 0, fmt.Errorf("!string")
	}
	return p.Single(`SELECT `+ret_column+` FROM `+table+` WHERE `+id_column+` = ?`, id).Int64()
}

func (p *Parser) MyTableChecking(table, id_column string, id int64, ret_column string) (bool, error) {
	if utils.CheckInputData(table, "string") || utils.CheckInputData(ret_column, "string") {
		return false, fmt.Errorf("!string")
	}

	if ok, err := p.CheckTableExists(table); !ok {
		return true, err
	}
	return false, nil
}

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

func (p *Parser) BlockError(err error) {
	if len(p.TxHash) == 0 {
		return
	}
	errText := err.Error()
	if len(errText) > 255 {
		errText = errText[:255]
	}
	p.DeleteQueueTx([]byte(p.TxHash))
	p.ExecSql("UPDATE transactions_status SET error = ? WHERE hex(hash) = ?", errText, p.TxHash)
}

func (p *Parser) AccessRights(condition string, iscondition bool) error {
	param := `value`
	if iscondition {
		param = `conditions`
	}
	conditions, err := p.Single(`SELECT `+param+` FROM "`+utils.Int64ToStr(int64(p.TxStateID))+`_state_parameters" WHERE name = ?`,
		condition).String()
	if err != nil {
		return err
	}
	if len(conditions) > 0 {
		ret, err := smart.EvalIf(conditions, &map[string]interface{}{`state`: p.TxStateID,
			`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID})
		if err != nil {
			return err
		}
		if !ret {
			return fmt.Errorf(`Access denied`)
		}
	}
	return nil
}

func (p *Parser) AccessTable(table, action string) error {

	if p.TxStateID == 0 {
		return nil
	}
	prefix := utils.Int64ToStr(int64(p.TxStateID))

	tablePermission, err := p.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions) as data WHERE name = ?`, "key", "value", table)
	if err != nil {
		return err
	}
	if len(tablePermission[action]) > 0 {
		ret, err := smart.EvalIf(tablePermission[action], &map[string]interface{}{`state`: p.TxStateID,
			`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID})
		if err != nil {
			return err
		}
		if !ret {
			return fmt.Errorf(`Access denied`)
		}
	}
	return nil
}

func (p *Parser) AccessColumns(table string, columns []string) error {

	if p.TxStateID == 0 {
		return nil
	}
	prefix := utils.Int64ToStr(int64(p.TxStateID))

	columnsAndPermissions, err := p.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions->'update') as data WHERE name = ?`,
		"key", "value", table)
	if err != nil {
		return err
	}
	for _, col := range columns {
		if cond, ok := columnsAndPermissions[col]; ok && len(cond) > 0 {
			ret, err := smart.EvalIf(cond, &map[string]interface{}{`state`: p.TxStateID,
				`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID})
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

func (p *Parser) AccessChange(table, name string) error {
	if p.TxStateID == 0 {
		return nil
	}

	prefix := utils.Int64ToStr(int64(p.TxStateID))

	conditions, err := p.Single(`SELECT conditions FROM "`+prefix+`_`+table+`" WHERE name = ?`, name).String()
	if err != nil {
		return err
	}

	if len(conditions) > 0 {
		ret, err := smart.EvalIf(conditions, &map[string]interface{}{`state`: p.TxStateID,
			`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID})
		if err != nil {
			return err
		}
		if !ret {
			return fmt.Errorf(`Access denied`)
		}
	}
	return nil
}
