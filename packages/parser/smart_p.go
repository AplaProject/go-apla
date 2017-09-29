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
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"encoding/hex"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/controllers"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/language"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

var (
	funcCallsDB = map[string]struct{}{
		"DBInsert":       struct{}{},
		"DBUpdate":       struct{}{},
		"DBUpdateExt":    struct{}{},
		"DBGetList":      struct{}{},
		"DBGetTable":     struct{}{},
		"DBString":       struct{}{},
		"DBInt":          struct{}{},
		"DBStringExt":    struct{}{},
		"DBIntExt":       struct{}{},
		"DBFreeRequest":  struct{}{},
		"DBStringWhere":  struct{}{},
		"DBIntWhere":     struct{}{},
		"DBAmount":       struct{}{},
		"UpdateContract": struct{}{},
		"UpdateParam":    struct{}{},
		"UpdateMenu":     struct{}{},
		"UpdatePage":     struct{}{},
		"DBInsertReport": struct{}{},
		"UpdateSysParam": struct{}{},
	}
	extendCost = map[string]int64{
		"AddressToId":    10,
		"IdToAddress":    10,
		"NewState":       1000, // ?? What cost must be?
		"Sha256":         50,
		"PubToID":        10,
		"StateVal":       10,
		"SysParamString": 10,
		"SysParamInt":    10,
		"SysCost":        10,
	}
)

func init() {
	smart.Extend(&script.ExtendData{Objects: map[string]interface{}{
		"DBInsert":           DBInsert,
		"DBUpdate":           DBUpdate,
		"DBUpdateExt":        DBUpdateExt,
		"DBGetList":          DBGetList,
		"DBGetTable":         DBGetTable,
		"DBString":           DBString,
		"DBInt":              DBInt,
		"DBStringExt":        DBStringExt,
		"DBFreeRequest":      DBFreeRequest,
		"DBIntExt":           DBIntExt,
		"DBStringWhere":      DBStringWhere,
		"DBIntWhere":         DBIntWhere,
		"Table":              StateTable,
		"TableTx":            StateTableTx,
		"AddressToId":        AddressToID,
		"IdToAddress":        IDToAddress,
		"DBAmount":           DBAmount,
		"ContractAccess":     IsContract,
		"ContractConditions": ContractConditions,
		"NewState":           NewStateFunc,
		"StateVal":           StateVal,
		"SysParamString":     SysParamString,
		"SysParamInt":        SysParamInt,
		"SysCost":            SysCost,
		"Int":                Int,
		"Str":                Str,
		"Money":              Money,
		"Float":              Float,
		"Len":                Len,
		"Sha256":             Sha256,
		"PubToID":            PubToID,
		"HexToBytes":         HexToBytes,
		"LangRes":            LangRes,
		"UpdateContract":     UpdateContract,
		"UpdateParam":        UpdateParam,
		"UpdateMenu":         UpdateMenu,
		"UpdatePage":         UpdatePage,
		"DBInsertReport":     DBInsertReport,
		"UpdateSysParam":     UpdateSysParam,
		"check_signature":    CheckSignature, // system function
	}, AutoPars: map[string]string{
		`*parser.Parser`: `parser`,
	}})
	smart.ExtendCost(getCost)
	smart.FuncCallsDB(funcCallsDB)
}

func getCost(name string) int64 {
	if val, ok := extendCost[name]; ok {
		return val
	}
	return -1
}

func (p *Parser) getExtend() *map[string]interface{} {
	head := p.TxSmart //consts.HeaderNew(contract.parser.TxPtr)
	var citizenID, walletID int64
	citizenID = int64(head.UserID)
	walletID = int64(head.UserID)
	// test
	block := int64(0)
	blockTime := int64(0)
	walletBlock := int64(0)
	if p.BlockData != nil {
		block = p.BlockData.BlockID
		walletBlock = p.BlockData.WalletID
		blockTime = p.BlockData.Time
	}
	extend := map[string]interface{}{`type`: head.Type, `time`: head.Time, `state`: head.StateID,
		`block`: block, `citizen`: citizenID, `wallet`: walletID, `wallet_block`: walletBlock,
		`parent`: ``, `txcost`: p.GetContractLimit(),
		`parser`: p, `contract`: p.TxContract, `block_time`: blockTime /*, `vars`: make(map[string]interface{})*/}
	for key, val := range p.TxData {
		extend[key] = val
	}

	return &extend
}

// StackCont adds an element to the stack of contract call or removes the top element when name is empty
func StackCont(p interface{}, name string) {
	cont := p.(*Parser).TxContract
	if len(name) > 0 {
		cont.StackCont = append(cont.StackCont, name)
	} else {
		cont.StackCont = cont.StackCont[:len(cont.StackCont)-1]
	}
	return
}

// CallContract calls the contract functions according to the specified flags
func (p *Parser) CallContract(flags int) (err error) {
	var public []byte
	if flags&smart.CallRollback == 0 {
		if len(p.TxSmart.PublicKey) > 0 && string(p.TxSmart.PublicKey) != `null` {
			public = p.TxSmart.PublicKey
		}
		if len(p.PublicKeys) == 0 {
			wallet := &model.DltWallet{}
			err := wallet.GetWallet(p.TxSmart.UserID)
			if err != nil {
				return err
			}
			if len(wallet.PublicKey) == 0 {
				if len(public) > 0 {
					p.PublicKeys = append(p.PublicKeys, public)
				} else {
					return fmt.Errorf("unknown wallet id")
				}
			} else {
				p.PublicKeys = append(p.PublicKeys, []byte(wallet.PublicKey))
			}
		}
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.TxData[`forsign`].(string), p.TxSmart.BinSignatures, false)
		if err != nil {
			return err
		}
		if !CheckSignResult {
			return fmt.Errorf("incorrect sign")
		}
	}

	methods := []string{`init`, `conditions`, `action`, `rollback`}
	p.TxContract.Extend = p.getExtend()
	p.TxContract.StackCont = []string{p.TxContract.Name}
	(*p.TxContract.Extend)[`stack_cont`] = StackCont
	before := (*p.TxContract.Extend)[`txcost`].(int64)
	var price int64 = -1
	if cprice := p.TxContract.GetFunc(`price`); cprice != nil {
		var ret []interface{}
		if ret, err = smart.Run(cprice, nil, p.TxContract.Extend); err != nil {
			return err
		} else if len(ret) == 1 {
			if _, ok := ret[0].(int64); !ok {
				return fmt.Errorf(`Wrong result type of price function`)
			}
			price = ret[0].(int64)
		} else {
			return fmt.Errorf(`Wrong type of price function`)
		}
	}
	systemParam := &model.SystemParameter{}
	err = systemParam.Get("fuel_rate")
	if err != nil {
		log.Fatal(err)
	}
	fuelRate, err := decimal.NewFromString(systemParam.Value)
	if err != nil {
		return err
	}
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
		return fmt.Errorf(`Fuel rate must be greater than 0`)
	}
	if !p.TxContract.Block.Info.(*script.ContractInfo).Active {
		return fmt.Errorf(`Contract %s is not active`, p.TxContract.Name)
	}
	p.TxContract.FreeRequest = false
	for i := uint32(0); i < 4; i++ {
		if (flags & (1 << i)) > 0 {
			cfunc := p.TxContract.GetFunc(methods[i])
			if cfunc == nil {
				continue
			}
			p.TxContract.Called = 1 << i
			_, err = smart.Run(cfunc, nil, p.TxContract.Extend)

			if err != nil {
				break
			}
		}
	}
	p.TxUsedCost = decimal.New(before-(*p.TxContract.Extend)[`txcost`].(int64), 0)
	p.TxContract.TxPrice = price
	return
}

// DBInsert inserts a record into the specified database table
func DBInsert(p *Parser, tblname string, params string, val ...interface{}) (qcost int64, ret int64, err error) { // map[string]interface{}) {
	if err = p.AccessTable(tblname, "insert"); err != nil {
		return
	}
	var ind int
	var lastID string
	if ind, err = model.NumIndexes(tblname); err != nil {
		return
	}
	qcost, lastID, err = p.selectiveLoggingAndUpd(strings.Split(params, `,`), val, tblname, nil, nil, true)
	if ind > 0 {
		qcost *= int64(ind)
	}
	if err == nil {
		ret, _ = strconv.ParseInt(lastID, 10, 64)
	}
	return
}

// DBInsertReport inserts a record into the specified report table
func DBInsertReport(p *Parser, tblname string, params string, val ...interface{}) (qcost int64, ret int64, err error) {
	var state int64
	qcost = 0
	names := strings.Split(tblname, `_`)
	if names[0] != `global` {
		state, err = strconv.ParseInt(names[0], 10, 64)
		if err != nil {
			logger.LogInfo(consts.StrToIntError, names[0])
		}
		if state != int64(p.TxStateID) {
			err = fmt.Errorf(`Wrong state in DBInsertReport`)
			return
		}
		if !model.IsNodeState(state, ``) {
			return
		}
	}
	tblname = names[0] + `_reports_` + strings.Join(names[1:], `_`)
	if err = p.AccessTable(tblname, "insert"); err != nil {
		return
	}
	var lastID string
	qcost, lastID, err = p.selectiveLoggingAndUpd(strings.Split(params, `,`), val, tblname, nil, nil, true)
	if err == nil {
		ret, _ = strconv.ParseInt(lastID, 10, 64)
	}
	return
}

func checkReport(tblname string) error {
	if strings.Contains(tblname, `_reports_`) {
		return fmt.Errorf(`Access denied to report table`)
	}
	return nil
}

// DBUpdate updates the item with the specified id in the table
func DBUpdate(p *Parser, tblname string, id int64, params string, val ...interface{}) (qcost int64, err error) { // map[string]interface{}) {
	qcost = 0
	/*	if err = p.AccessTable(tblname, "general_update"); err != nil {
		return
	}*/
	if err = checkReport(tblname); err != nil {
		return
	}
	columns := strings.Split(params, `,`)
	if err = p.AccessColumns(tblname, columns); err != nil {
		return
	}
	qcost, _, err = p.selectiveLoggingAndUpd(columns, val, tblname, []string{`id`}, []string{converter.Int64ToStr(id)}, true)
	return
}

// DBUpdateExt updates the record in the specified table. You can specify 'where' query in params and then the values for this query
func DBUpdateExt(p *Parser, tblname string, column string, value interface{}, params string, val ...interface{}) (qcost int64, err error) { // map[string]interface{}) {
	qcost = 0
	var isIndex bool

	if err = checkReport(tblname); err != nil {
		return
	}

	columns := strings.Split(params, `,`)
	if err = p.AccessColumns(tblname, columns); err != nil {
		return
	}
	if isIndex, err = model.IsIndex(tblname, column); err != nil {
		return
	} else if !isIndex {
		err = fmt.Errorf(`there is not index on %s`, column)
	} else {
		qcost, _, err = p.selectiveLoggingAndUpd(columns, val, tblname, []string{column}, []string{fmt.Sprint(value)}, true)
	}
	return
}

// DBString returns the value of the field of the record with the specified id
func DBString(tblname string, name string, id int64) (int64, string, error) {
	if err := checkReport(tblname); err != nil {
		return 0, ``, err
	}
	cost, err := model.GetQueryTotalCost(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where id=?`, id)
	if err != nil {
		return 0, "", nil
	}
	res, err := model.Single(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where id=?`, id).String()
	return cost, res, err
}

// Sha256 returns SHA256 hash value
func Sha256(text string) string {
	hash, err := crypto.Hash([]byte(text))
	if err != nil {
		log.Fatal(err)
	}
	hash = converter.BinToHex(hash)
	return string(hash)
}

// PubToID returns a numeric identifier for the public key specified in the hexadecimal form.
func PubToID(hexkey string) int64 {
	pubkey, err := hex.DecodeString(hexkey)
	if err != nil {
		return 0
	}
	return crypto.Address(pubkey)
}

// HexToBytes converts the hexadecimal representation to []byte
func HexToBytes(hexdata string) ([]byte, error) {
	return hex.DecodeString(hexdata)
}

// DBInt returns the numeric value of the column for the record with the specified id
func DBInt(tblname string, name string, id int64) (int64, int64, error) {
	if err := checkReport(tblname); err != nil {
		return 0, 0, err
	}
	cost, err := model.GetQueryTotalCost(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where id=?`, id)
	if err != nil {
		return 0, 0, err
	}
	res, err := model.Single(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where id=?`, id).Int64()
	return cost, res, err
}

func getBytea(table string) map[string]bool {
	isBytea := make(map[string]bool)
	colTypes, err := model.GetAll(`select column_name, data_type from information_schema.columns where table_name=?`, -1, table)
	if err != nil {
		return isBytea
	}
	for _, icol := range colTypes {
		isBytea[icol[`column_name`]] = icol[`data_type`] == `bytea`
	}
	return isBytea
}

// DBStringExt returns the value of 'name' column for the record with the specified value of the 'idname' field
func DBStringExt(tblname string, name string, id interface{}, idname string) (int64, string, error) {
	if err := checkReport(tblname); err != nil {
		return 0, ``, err
	}

	isBytea := getBytea(tblname)
	if isBytea[idname] {
		switch id.(type) {
		case string:
			if vbyte, err := hex.DecodeString(id.(string)); err == nil {
				id = vbyte
			}
		}
	}

	if isIndex, err := model.IsIndex(tblname, idname); err != nil {
		return 0, ``, err
	} else if !isIndex {
		return 0, ``, fmt.Errorf(`there is not index on %s`, idname)
	}
	cost, err := model.GetQueryTotalCost(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where `+converter.EscapeName(idname)+`=?`, id)
	if err != nil {
		return 0, "", err
	}
	res, err := model.Single(`select `+converter.EscapeName(name)+` from `+converter.EscapeName(tblname)+` where `+converter.EscapeName(idname)+`=?`, id).String()
	return cost, res, err
}

// DBIntExt returns the numeric value of the 'name' column for the record with the specified value of the 'idname' field
func DBIntExt(tblname string, name string, id interface{}, idname string) (cost int64, ret int64, err error) {
	var val string
	var qcost int64
	qcost, val, err = DBStringExt(tblname, name, id, idname)
	if err != nil {
		return 0, 0, err
	}
	if len(val) == 0 {
		return 0, 0, nil
	}
	res, err := strconv.ParseInt(val, 10, 64)
	return qcost, res, err
}

// DBFreeRequest is a free function that is needed to find the record with the specified value in the 'idname' column.
func DBFreeRequest(p *Parser, tblname string /*name string,*/, id interface{}, idname string) (int64, error) {
	if p.TxContract.FreeRequest {
		return 0, fmt.Errorf(`DBFreeRequest can be executed only once`)
	}
	p.TxContract.FreeRequest = true
	cost, ret, err := DBStringExt(tblname, idname, id, idname)
	if err != nil {
		return 0, err
	}
	if len(ret) > 0 || ret == fmt.Sprintf(`%v`, id) {
		return 0, nil
	}
	return cost, fmt.Errorf(`DBFreeRequest: cannot find %v in %s of %s`, id, idname, tblname)
}

// DBStringWhere returns the column value based on the 'where' condition and 'params' values for this condition
func DBStringWhere(tblname string, name string, where string, params ...interface{}) (int64, string, error) {
	if err := checkReport(tblname); err != nil {
		return 0, ``, err
	}

	re := regexp.MustCompile(`([a-z]+[\w_]*)\"?\s*[><=]`)
	ret := re.FindAllStringSubmatch(where, -1)
	for _, iret := range ret {
		if len(iret) != 2 {
			continue
		}
		if isIndex, err := model.IsIndex(tblname, iret[1]); err != nil {
			return 0, ``, err
		} else if !isIndex {
			return 0, ``, fmt.Errorf(`there is not index on %s`, iret[1])
		}
	}
	selectQuery := `select ` + converter.EscapeName(name) + ` from ` + converter.EscapeName(tblname) + ` where ` + strings.Replace(converter.Escape(where), `$`, `?`, -1)
	qcost, err := model.GetQueryTotalCost(selectQuery, params...)
	if err != nil {
		return 0, "", err
	}
	res, err := model.Single(selectQuery, params).String()
	if err != nil {
		return 0, "", err
	}
	return qcost, res, err
}

// DBIntWhere returns the column value based on the 'where' condition and 'params' values for this condition
func DBIntWhere(tblname string, name string, where string, params ...interface{}) (cost int64, ret int64, err error) {
	var val string
	cost, val, err = DBStringWhere(tblname, name, where, params...)
	if err != nil {
		return 0, 0, err
	}
	if len(val) == 0 {
		return 0, 0, nil
	}
	res, err := strconv.ParseInt(val, 10, 64)
	return cost, res, err
}

// StateTable adds a prefix with the state number to the table name
func StateTable(p *Parser, tblname string) string {
	return fmt.Sprintf("%d_%s", p.TxStateID, tblname)
}

// StateTable adds a prefix with the state number to the table name
func StateTableTx(p *Parser, tblname string) string {
	return fmt.Sprintf("%v_%s", p.TxData[`StateId`], tblname)
}

// ContractConditions calls the 'conditions' function for each of the contracts specified in the parameters
func ContractConditions(p *Parser, names ...interface{}) (bool, error) {
	for _, iname := range names {
		name := iname.(string)
		if len(name) > 0 {
			contract := smart.GetContract(name, p.TxStateID)
			if contract == nil {
				contract = smart.GetContract(name, 0)
				if contract == nil {
					return false, fmt.Errorf(`Unknown contract %s`, name)
				}
			}
			block := contract.GetFunc(`conditions`)
			if block == nil {
				return false, fmt.Errorf(`There is not conditions in contract %s`, name)
			}
			_, err := smart.Run(block, []interface{}{}, &map[string]interface{}{`state`: int64(p.TxStateID),
				`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID, `parser`: p})
			if err != nil {
				return false, err
			}
		} else {
			return false, fmt.Errorf(`empty contract name in ContractConditions`)
		}
	}
	return true, nil
}

// IsContract checks whether the name of the executable contract matches one of the names listed in the parameters.
func IsContract(p *Parser, names ...interface{}) bool {
	for _, iname := range names {
		name := iname.(string)
		if p.TxContract != nil && len(name) > 0 {
			if name[0] != '@' {
				name = fmt.Sprintf(`@%d`, p.TxStateID) + name
			}
			//		return p.TxContract.Name == name
			if p.TxContract.StackCont[len(p.TxContract.StackCont)-1] == name {
				return true
			}
		} else if len(p.TxSlice) > 1 {
			if consts.TxTypes[converter.BytesToInt(p.TxSlice[1])] == name {
				return true
			}
		}
	}
	return false
}

// IsGovAccount checks whether the specified account is the owner of the state
func IsGovAccount(p *Parser, citizen int64) bool {
	account, err := strconv.ParseInt(StateVal(p, `gov_account`), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, StateVal(p, `gov_account`))
	}
	return account == citizen
}

// AddressToID converts the string representation of the wallet number to a numeric
func AddressToID(input string) (addr int64) {
	input = strings.TrimSpace(input)
	if len(input) < 2 {
		return 0
	}
	if input[0] == '-' {
		addr, _ = strconv.ParseInt(input, 10, 64)
	} else if strings.Count(input, `-`) == 4 {
		addr = converter.StringToAddress(input)
	} else {
		uaddr, _ := strconv.ParseUint(input, 10, 64)
		addr = int64(uaddr)
	}
	if !converter.IsValidAddress(converter.AddressToString(addr)) {
		return 0
	}
	return
}

// IDToAddress converts the identifier of account to a string of the form XXXX -...- XXXX
func IDToAddress(id int64) (out string) {
	out = converter.AddressToString(id)
	if !converter.IsValidAddress(out) {
		out = `invalid`
	}
	return
}

// DBAmount returns the value of the 'amount' column for the record with the 'id' value in the 'column' column
func DBAmount(tblname, column string, id int64) (int64, decimal.Decimal) {
	if err := checkReport(tblname); err != nil {
		return 0, decimal.New(0, 0)
	}

	balance, err := model.Single("SELECT amount FROM "+converter.EscapeName(tblname)+" WHERE "+converter.EscapeName(column)+" = ?", id).String()
	if err != nil {
		return 0, decimal.New(0, 0)
	}
	val, _ := decimal.NewFromString(balance)
	return 0, val
}

// EvalIf counts and returns the logical value of the specified expression
func (p *Parser) EvalIf(conditions string) (bool, error) {
	time := int64(0)
	if p.TxSmart != nil {
		time = p.TxSmart.Time
	}
	/*	if p.TxPtr != nil {
		switch val := p.TxPtr.(type) {
		case *consts.TXHeader:
			time = int64(val.Time)
		}
	}*/
	blockTime := int64(0)
	if p.BlockData != nil {
		blockTime = p.BlockData.Time
	}

	return smart.EvalIf(conditions, converter.Int64ToStr(int64(p.TxStateID)), &map[string]interface{}{`state`: p.TxStateID,
		`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID, `parser`: p,
		`block_time`: blockTime, `time`: time})
}

// StateVal returns the value of the specified parameter for the state
func StateVal(p *Parser, name string) string {
	val, _ := template.StateParam(int64(p.TxStateID), name)
	return val
}

// SysParamString returns the value of the system parameter
func SysParamString(name string) string {
	return syspar.SysString(name)
}

// SysParamInt returns the value of the system parameter
func SysParamInt(name string) int64 {
	return syspar.SysInt64(name)
}

// SysCost returns the cost of the transaction from the system parameter
func SysCost(name string) int64 {
	return syspar.SysCost(name)
}

// Int converts a string to a number
func Int(val string) int64 {
	value, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, val)
	}
	return value
}

// Str converts the value to a string
func Str(v interface{}) (ret string) {
	switch val := v.(type) {
	case float64:
		ret = fmt.Sprintf(`%f`, val)
	default:
		ret = fmt.Sprintf(`%v`, val)
	}
	return
}

// Money converts the value into a numeric type for money
func Money(v interface{}) (ret decimal.Decimal) {
	return script.ValueToDecimal(v)
}

// Float converts the value to float64
func Float(v interface{}) (ret float64) {
	return script.ValueToFloat(v)
}

// UpdateContract updates the content and condition of contract with the specified name
func UpdateContract(p *Parser, name, value, conditions string) (int64, error) {
	var (
		fields []string
		values []interface{}
	)
	prefix := converter.Int64ToStr(int64(p.TxStateID))
	sc := &model.SmartContract{}
	sc.SetTablePrefix(prefix)
	err := sc.GetByName(name)
	if err != nil {
		return 0, err
	}
	cond := sc.Conditions
	if len(cond) > 0 {
		ret, err := p.EvalIf(cond)
		if err != nil {
			return 0, err
		}
		if !ret {
			if err = p.AccessRights(`changing_smart_contracts`, false); err != nil {
				return 0, err
			}
		}
	}
	if len(value) > 0 {
		fields = append(fields, "value")
		values = append(values, value)
	}
	if len(conditions) > 0 {
		if err := smart.CompileEval(conditions, p.TxStateID); err != nil {
			return 0, err
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	if len(fields) == 0 {
		return 0, fmt.Errorf(`empty value and condition`)
	}
	root, err := smart.CompileBlock(value, prefix, false, sc.ID)
	if err != nil {
		return 0, err
	}
	_, _, err = p.selectiveLoggingAndUpd(fields, values,
		prefix+"_smart_contracts", []string{"id"}, []string{converter.Int64ToStr(sc.ID)}, true)
	if err != nil {
		return 0, err
	}
	for i, item := range root.Children {
		if item.Type == script.ObjContract {
			root.Children[i].Info.(*script.ContractInfo).TableID = sc.ID
			root.Children[i].Info.(*script.ContractInfo).Active = sc.Active == "1"
		}
	}
	smart.FlushBlock(root)

	return 0, nil
}

// UpdateParam updates the value and condition of parameter with the specified name for the state
func UpdateParam(p *Parser, name, value, conditions string) (int64, error) {
	var (
		fields []string
		values []interface{}
	)

	if err := p.AccessRights(name, true); err != nil {
		return 0, err
	}
	if len(value) > 0 {
		fields = append(fields, "value")
		values = append(values, value)
	}
	if len(conditions) > 0 {
		if err := smart.CompileEval(conditions, uint32(p.TxStateID)); err != nil {
			return 0, err
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	if len(fields) == 0 {
		return 0, fmt.Errorf(`empty value and condition`)
	}
	_, _, err := p.selectiveLoggingAndUpd(fields, values,
		converter.Int64ToStr(int64(p.TxStateID))+"_state_parameters", []string{"name"}, []string{name}, true)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

// UpdateMenu updates the value and condition for the specified menu
func UpdateMenu(p *Parser, name, value, conditions, global string, stateID int64) error {
	if err := p.AccessChange(`menu`, p.TxStateIDStr, global, stateID); err != nil {
		return err
	}
	fields := []string{"value"}
	values := []interface{}{value}
	if len(conditions) > 0 {
		if err := smart.CompileEval(conditions, uint32(p.TxStateID)); err != nil {
			return err
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	_, _, err := p.selectiveLoggingAndUpd(fields, values, converter.Int64ToStr(int64(p.TxStateID))+"_menu",
		[]string{"name"}, []string{name}, true)
	if err != nil {
		return err
	}
	return nil
}

// CheckSignature checks the additional signatures for the contract
func CheckSignature(i *map[string]interface{}, name string) error {
	state, name := script.ParseContract(name)
	pref := converter.Int64ToStr(int64(state))
	if state == 0 {
		pref = `global`
	}
	//	fmt.Println(`CheckSignature`, i, state, name)
	p := (*i)[`parser`].(*Parser)
	value, err := model.Single(`select value from "`+pref+`_signatures" where name=?`, name).String()
	if err != nil {
		return err
	}
	if len(value) == 0 {
		return nil
	}
	hexsign, err := hex.DecodeString((*i)[`Signature`].(string))
	if len(hexsign) == 0 || err != nil {
		return fmt.Errorf(`wrong signature`)
	}

	var sign controllers.TxSignJSON
	err = json.Unmarshal([]byte(value), &sign)
	if err != nil {
		return err
	}
	wallet := (*i)[`wallet`].(int64)
	if wallet == 0 {
		wallet = (*i)[`citizen`].(int64)
	}
	forsign := fmt.Sprintf(`%d,%d`, uint64((*i)[`time`].(int64)), uint64(wallet))
	for _, isign := range sign.Params {
		forsign += fmt.Sprintf(`,%v`, (*i)[isign.Param])
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forsign, hexsign, true)
	if err != nil {
		return err
	}
	if !CheckSignResult {
		return fmt.Errorf(`incorrect signature ` + forsign)
	}
	return nil
}

// UpdatePage updates the text, menu and condition of the specified page
func UpdatePage(p *Parser, name, value, menu, conditions, global string, stateID int64) error {
	if err := p.AccessChange(`pages`, name, global, stateID); err != nil {
		return p.ErrInfo(err)
	}
	fields := []string{"value"}
	values := []interface{}{value}
	if len(conditions) > 0 {
		if err := smart.CompileEval(conditions, uint32(p.TxStateID)); err != nil {
			return err
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	if len(menu) > 0 {
		fields = append(fields, "menu")
		values = append(values, menu)
	}
	_, _, err := p.selectiveLoggingAndUpd(fields, values, converter.Int64ToStr(int64(p.TxStateID))+"_pages",
		[]string{"name"}, []string{name}, true)
	if err != nil {
		return err
	}

	return nil
}

// Len returns the length of the slice
func Len(in []interface{}) int64 {
	return int64(len(in))
}

// LangRes returns the language resource
func LangRes(p *Parser, idRes, lang string) string {
	ret, _ := language.LangText(idRes, int(p.TxStateID), lang)
	return ret
}

func checkWhere(tblname string, where string, order string) (string, string, error) {
	re := regexp.MustCompile(`([a-z]+[\w_]*)\"?\s*[><=]`)
	ret := re.FindAllStringSubmatch(where, -1)

	for _, iret := range ret {
		if len(iret) != 2 {
			continue
		}
		if isIndex, err := model.IsIndex(tblname, iret[1]); err != nil {
			return ``, ``, err
		} else if !isIndex {
			return ``, ``, fmt.Errorf(`there is not index on %s`, iret[1])
		}
	}
	if len(order) > 0 {
		order = ` order by ` + converter.EscapeName(order)
	}
	return strings.Replace(converter.Escape(where), `$`, `?`, -1), order, nil
}

// DBGetList returns a list of column values with the specified 'offset', 'limit', 'where'
func DBGetList(tblname string, name string, offset, limit int64, order string,
	where string, params ...interface{}) (int64, []interface{}, error) {

	if err := checkReport(tblname); err != nil {
		return 0, nil, err
	}

	re := regexp.MustCompile(`([a-z]+[\w_]*)\"?\s*[><=]`)
	ret := re.FindAllStringSubmatch(where, -1)

	for _, iret := range ret {
		if len(iret) != 2 {
			continue
		}
		if isIndex, err := model.IsIndex(tblname, iret[1]); err != nil {
			return 0, nil, err
		} else if !isIndex {
			return 0, nil, fmt.Errorf(`there is not index on %s`, iret[1])
		}
	}
	if len(order) > 0 {
		order = ` order by ` + converter.EscapeName(order)
	}
	if limit <= 0 {
		limit = -1
	}
	list, err := model.GetAll(`select `+converter.Escape(name)+` from `+converter.EscapeName(tblname)+` where `+
		strings.Replace(converter.Escape(where), `$`, `?`, -1)+order+fmt.Sprintf(` offset %d `, offset), int(limit), params...)
	result := make([]interface{}, len(list))
	for i := 0; i < len(list); i++ {
		result[i] = reflect.ValueOf(list[i]).Interface()
	}
	return 0, result, err
}

// DBGetTable returns an array of values of the specified columns when there is selection of data 'offset', 'limit', 'where'
func DBGetTable(tblname string, columns string, offset, limit int64, order string,
	where string, params ...interface{}) (int64, []interface{}, error) {
	var err error
	if err = checkReport(tblname); err != nil {
		return 0, nil, err
	}

	where, order, err = checkWhere(tblname, where, order)
	if limit <= 0 {
		limit = -1
	}
	cols := strings.Split(converter.Escape(columns), `,`)
	list, err := model.GetAll(`select `+strings.Join(cols, `,`)+` from `+converter.EscapeName(tblname)+` where `+
		where+order+fmt.Sprintf(` offset %d `, offset), int(limit), params...)
	result := make([]interface{}, len(list))
	for i := 0; i < len(list); i++ {
		//result[i] = make(map[string]interface{})
		result[i] = reflect.ValueOf(list[i]).Interface()
		/*		for _, key := range cols {
				result[i][key] = reflect.ValueOf(list[i][key]).Interface()
			}*/
	}
	return 0, result, err
}

// NewStateFunc creates a new country
func NewStateFunc(p *Parser, country, currency string) (err error) {
	newStateParser := NewStateParser{p, nil}
	_, err = newStateParser.Main(country, currency)
	return
}

// UpdateSysParam updates the system parameter
func UpdateSysParam(p *Parser, name, value, conditions string) (int64, error) {
	var (
		fields []string
		values []interface{}
	)

	par := &model.SystemParameter{}
	err := par.Get(name)
	if err != nil {
		return 0, err
	}
	cond := par.Conditions
	if len(cond) > 0 {
		ret, err := p.EvalIf(cond)
		if err != nil {
			return 0, err
		}
		if !ret {
			return 0, fmt.Errorf(`Access denied`)
		}
	}
	if len(value) > 0 {
		fields = append(fields, "value")
		values = append(values, value)
	}
	if len(conditions) > 0 {
		if err := smart.CompileEval(conditions, 0); err != nil {
			return 0, err
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	if len(fields) == 0 {
		return 0, fmt.Errorf(`empty value and condition`)
	}
	_, _, err = p.selectiveLoggingAndUpd(fields, values, "system_parameters", []string{"name"}, []string{name}, true)
	if err != nil {
		return 0, err
	}
	err = syspar.SysUpdate()
	if err != nil {
		return 0, err
	}
	return 0, nil
}
