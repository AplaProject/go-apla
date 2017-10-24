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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"encoding/hex"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/language"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/templatev2"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/shopspring/decimal"

	"github.com/jinzhu/gorm"
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
		"DBRowExt":       struct{}{},
		"DBRow":          struct{}{},
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
		"FindEcosystem":  struct{}{},
	}
	extendCost = map[string]int64{
		"AddressToId":       10,
		"IdToAddress":       10,
		"NewState":          1000, // ?? What cost must be?
		"Sha256":            50,
		"PubToID":           10,
		"StateVal":          10,
		"SysParamString":    10,
		"SysParamInt":       10,
		"SysCost":           10,
		"SysFuel":           10,
		"ValidateCondition": 30,
		"PrefixTable":       10,
		"EvalCondition":     20,
		"HasPrefix":         10,
		"Contains":          10,
		"Replace":           10,
		"UpdateLang":        10,
		"Size":              10,
		"Substr":            10,
		"ContractsList":     10,
		"IsContract":        10,
		"CompileContract":   100,
		"FlushContract":     50,
		"Eval":              10,
		"Activate":          10,
		"CreateEcosystem":   100,
		"RollbackEcosystem": 100,
		"TableConditions":   100,
		"CreateTable":       100,
		"RollbackTable":     100,
		"PermTable":         100,
		"ColumnCondition":   50,
		"CreateColumn":      50,
		"RollbackColumn":    50,
		"PermColumn":        50,
	}
)

//SignRes contains the data of the signature
type SignRes struct {
	Param string `json:"name"`
	Text  string `json:"text"`
}

// TxSignJSON is a structure for additional signs of transaction
type TxSignJSON struct {
	ForSign string    `json:"forsign"`
	Field   string    `json:"field"`
	Title   string    `json:"title"`
	Params  []SignRes `json:"params"`
}

func init() {
	smart.Extend(&script.ExtendData{Objects: map[string]interface{}{
		"DBInsert":           DBInsert,
		"DBUpdate":           DBUpdate,
		"DBUpdateExt":        DBUpdateExt,
		"DBGetList":          DBGetList,
		"DBGetTable":         DBGetTable,
		"DBString":           DBString,
		"DBInt":              DBInt,
		"DBRowExt":           DBRowExt,
		"DBRow":              DBRow,
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
		"ContractAccess":     ContractAccess,
		"ContractConditions": ContractConditions,
		"StateVal":           StateVal,
		"SysParamString":     SysParamString,
		"SysParamInt":        SysParamInt,
		"SysCost":            SysCost,
		"SysFuel":            SysFuel,
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
		"ValidateCondition":  ValidateCondition,
		"PrefixTable":        PrefixTable,
		"EvalCondition":      EvalCondition,
		"HasPrefix":          strings.HasPrefix,
		"Contains":           strings.Contains,
		"Replace":            Replace,
		"FindEcosystem":      FindEcosystem,
		"CreateEcosystem":    CreateEcosystem,
		"RollbackEcosystem":  RollbackEcosystem,
		"CreateTable":        CreateTable,
		"RollbackTable":      RollbackTable,
		"PermTable":          PermTable,
		"TableConditions":    TableConditions,
		"ColumnCondition":    ColumnCondition,
		"CreateColumn":       CreateColumn,
		"RollbackColumn":     RollbackColumn,
		"PermColumn":         PermColumn,
		"UpdateLang":         UpdateLang,
		"Size":               Size,
		"Substr":             Substr,
		"ContractsList":      ContractsList,
		"IsContract":         IsContract,
		"CompileContract":    CompileContract,
		"FlushContract":      FlushContract,
		"Eval":               Eval,
		"Activate":           ActivateContract,
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

// GetContractLimit returns the default maximal cost of contract
func (p *Parser) GetContractLimit() (ret int64) {
	// default maximum cost of F
	if len(p.TxSmart.MaxSum) > 0 {
		p.TxCost = converter.StrToInt64(p.TxSmart.MaxSum)
	} else {
		cost, _ := templatev2.StateParam(p.TxSmart.StateID, `max_sum`)
		if len(cost) > 0 {
			p.TxCost = converter.StrToInt64(cost)
		}
	}
	if p.TxCost == 0 {
		p.TxCost = script.CostDefault // fuel
	}
	return p.TxCost
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
		`parent`: ``, `txcost`: p.GetContractLimit(), `txhash`: p.TxHash, `result`: ``,
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

/*func (p *Parser) payContract() error {
	var (
		fromID int64
		err    error
	)
	//return nil
	toID := p.BlockData.WalletID // account of node
	fuel, err := decimal.NewFromString(syspar.GetFuelRate(p.TxSmart.TokenEcosystem))
	if err != nil {
		return err
	}
	if fuel.Cmp(decimal.New(0, 0)) <= 0 {
		return fmt.Errorf(`fuel rate must be greater than 0`)
	}


	egs := p.TxUsedCost.Mul(fuel)
	fmt.Printf("Pay fuel=%v fromID=%d toID=%d cost=%v egs=%v", fuel, fromID, toID, p.TxUsedCost, egs)
	if egs.Cmp(decimal.New(0, 0)) == 0 { // Is it possible to pay nothing?
		return nil
	}
	wallet := &model.DltWallet{}
	if err := wallet.GetWallet(fromID); err != nil {
		return err
	}
	wltAmount, err := decimal.NewFromString(wallet.Amount)
	if err != nil {
		return err
	}

	if wltAmount.Cmp(egs) < 0 {
		egs = wltAmount
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
		[]string{converter.Int64ToStr(syspar.GetCommissionWallet())}, true); err != nil {
		return err
	}
	//	fmt.Printf(" Paid commission %v\r\n", commission)
	return nil
}*/

// CallContract calls the contract functions according to the specified flags
func (p *Parser) CallContract(flags int) (err error) {
	var (
		public                 []byte
		sizeFuel, toID, fromID int64
		fuelRate               decimal.Decimal
	)
	payWallet := &model.Key{}
	p.TxContract.Extend = p.getExtend()
	var price int64

	methods := []string{`init`, `conditions`, `action`, `rollback`}
	p.TxContract.StackCont = []string{p.TxContract.Name}
	(*p.TxContract.Extend)[`stack_cont`] = StackCont

	if flags&smart.CallRollback == 0 && (flags&smart.CallAction) != 0 {
		toID = p.BlockData.WalletID
		fromID = p.TxSmart.UserID
		if len(p.TxSmart.PublicKey) > 0 && string(p.TxSmart.PublicKey) != `null` {
			public = p.TxSmart.PublicKey
		}
		wallet := &model.Key{}
		wallet.SetTablePrefix(p.TxSmart.StateID)
		err := wallet.Get(p.TxSmart.UserID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if len(wallet.PublicKey) > 0 {
			public = wallet.PublicKey
		}
		if p.TxSmart.Type == 258 { // UpdFullNodes
			node := syspar.GetNode(p.TxSmart.UserID)
			if node == nil {
				return fmt.Errorf("unknown node id")
			}
			public = node.Public
		}
		if len(public) == 0 {
			return fmt.Errorf("empty public key")
		}
		p.PublicKeys = append(p.PublicKeys, public)
		//		fmt.Println(`CALL CONTRACT`, p.TxData[`forsign`].(string))
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.TxData[`forsign`].(string), p.TxSmart.BinSignatures, false)
		if err != nil {
			fmt.Println(`ForSign`, p.TxData[`forsign`].(string))
			return err
		}
		if !CheckSignResult {
			return fmt.Errorf("incorrect sign")
		}
		if p.TxSmart.StateID > 0 {
			if p.TxSmart.TokenEcosystem == 0 {
				p.TxSmart.TokenEcosystem = p.TxSmart.StateID
			}
			fuelRate, err = decimal.NewFromString(syspar.GetFuelRate(p.TxSmart.TokenEcosystem))
			if err != nil {
				return err
			}
			if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
				return fmt.Errorf(`Fuel rate must be greater than 0`)
			}
			if len(p.TxSmart.PayOver) > 0 {
				payOver, err := decimal.NewFromString(p.TxSmart.PayOver)
				if err != nil {
					return err
				}
				fuelRate = fuelRate.Add(payOver)
			}
			if p.TxContract.Block.Info.(*script.ContractInfo).Owner.Active {
				fromID = p.TxContract.Block.Info.(*script.ContractInfo).Owner.WalletID
				p.TxSmart.TokenEcosystem = p.TxContract.Block.Info.(*script.ContractInfo).Owner.TokenID
			} else if len(p.TxSmart.PayOver) > 0 {
				payOver, err := decimal.NewFromString(p.TxSmart.PayOver)
				if err != nil {
					return err
				}
				fuelRate = fuelRate.Add(payOver)
			}
			payWallet.SetTablePrefix(p.TxSmart.TokenEcosystem)
			if err = payWallet.Get(fromID); err != nil {
				return err
			}
			if !bytes.Equal(wallet.PublicKey, payWallet.PublicKey) && !bytes.Equal(p.TxSmart.PublicKey, payWallet.PublicKey) {
				return fmt.Errorf(`Token and user public keys are different`)
			}
			amount, err := decimal.NewFromString(payWallet.Amount)
			if err != nil {
				return err
			}
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
			sizeFuel = syspar.GetSizeFuel() * int64(len(p.TxSmart.Data)) / 1024
			if amount.Cmp(decimal.New(sizeFuel+price, 0).Mul(fuelRate)) <= 0 {
				return fmt.Errorf(`current balance is not enough`)
			}
		}
	}
	before := (*p.TxContract.Extend)[`txcost`].(int64) + price

	// Payment for the size
	(*p.TxContract.Extend)[`txcost`] = (*p.TxContract.Extend)[`txcost`].(int64) - sizeFuel

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
				before -= price
				break
			}
		}
	}
	p.TxUsedCost = decimal.New(before-(*p.TxContract.Extend)[`txcost`].(int64), 0)
	p.TxContract.TxPrice = price
	if (flags&smart.CallAction) != 0 && p.TxSmart.StateID > 0 {
		apl := p.TxUsedCost.Mul(fuelRate)
		fmt.Printf("Pay fuel=%v fromID=%d toID=%d cost=%v apl=%v", fuelRate, fromID, toID, p.TxUsedCost, apl)
		wltAmount, err := decimal.NewFromString(payWallet.Amount)
		if err != nil {
			return err
		}
		if wltAmount.Cmp(apl) < 0 {
			apl = wltAmount
		}
		commission := apl.Mul(decimal.New(syspar.SysInt64(`commission_size`), 0)).Div(decimal.New(100, 0)).Floor()
		walletTable := fmt.Sprintf(`%d_keys`, p.TxSmart.TokenEcosystem)
		if _, _, err := p.selectiveLoggingAndUpd([]string{`-amount`}, []interface{}{apl}, walletTable, []string{`id`},
			[]string{converter.Int64ToStr(fromID)}, true); err != nil {
			return err
		}
		if _, _, err := p.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{apl.Sub(commission)}, walletTable, []string{`id`},
			[]string{converter.Int64ToStr(toID)}, true); err != nil {
			return err
		}
		if _, _, err := p.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{commission}, walletTable, []string{`id`},
			[]string{syspar.GetCommissionWallet(p.TxSmart.TokenEcosystem)}, true); err != nil {
			return err
		}
		fmt.Printf(" Paid commission %v\r\n", commission)
	}
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
	qcost = 0
	names := strings.Split(tblname, `_`)
	if names[0] != `global` {
		state := converter.StrToInt64(names[0])
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
		err = fmt.Errorf(`there is no index on %s`, column)
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
		isBytea[icol[`column_name`]] = icol[`column_name`] != `conditions` && icol[`data_type`] == `bytea`
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
		return 0, ``, fmt.Errorf(`there is no index on %s`, idname)
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
			return 0, ``, fmt.Errorf(`there is no index on %s`, iret[1])
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

// StateTableTx adds a prefix with the state number to the table name
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

// ContractAccess checks whether the name of the executable contract matches one of the names listed in the parameters.
func ContractAccess(p *Parser, names ...interface{}) bool {
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
	return converter.StrToInt64(StateVal(p, `founder_account`)) == citizen
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

	return smart.EvalIf(conditions, p.TxStateID, &map[string]interface{}{`state`: p.TxStateID,
		`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID, `parser`: p,
		`block_time`: blockTime, `time`: time})
}

// StateVal returns the value of the specified parameter for the state
func StateVal(p *Parser, name string) string {
	val, _ := templatev2.StateParam(int64(p.TxStateID), name)
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

// SysFuel returns the fuel rate
func SysFuel(state int64) string {
	return syspar.GetFuelRate(state)
}

// Int converts a string to a number
func Int(val string) int64 {
	return converter.StrToInt64(val)
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
	found, err := sc.GetByName(name)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, errors.New("can't find smart contract: Name " + name)
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
	root, err := smart.CompileBlock(value, &script.OwnerInfo{StateID: uint32(converter.StrToInt64(prefix)),
		Active: false, TableID: sc.ID, WalletID: sc.WalletID, TokenID: 0})
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
			root.Children[i].Info.(*script.ContractInfo).Owner.TableID = sc.ID
			root.Children[i].Info.(*script.ContractInfo).Owner.Active = sc.Active == "1"
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

	var sign TxSignJSON
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
			return ``, ``, fmt.Errorf(`there is no index on %s`, iret[1])
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

// DBRowExt returns one row from the table StringExt
func DBRowExt(tblname string, columns string, id interface{}, idname string) (int64, map[string]string, error) {

	if err := checkReport(tblname); err != nil {
		return 0, nil, err
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
		return 0, nil, err
	} else if !isIndex {
		return 0, nil, fmt.Errorf(`there is no index on %s`, idname)
	}
	query := `select ` + converter.Sanitize(columns, ` ,()*`) + ` from ` + converter.EscapeName(tblname) + ` where ` + converter.EscapeName(idname) + `=?`
	cost, err := model.GetQueryTotalCost(query, id)
	if err != nil {
		return 0, nil, err
	}
	res, err := model.GetOneRow(query, id).String()

	return cost, res, err
}

// DBRow returns one row from the table StringExt
func DBRow(tblname string, columns string, id int64) (int64, map[string]string, error) {

	if err := checkReport(tblname); err != nil {
		return 0, nil, err
	}

	query := `select ` + converter.Sanitize(columns, ` ,()*`) + ` from ` + converter.EscapeName(tblname) + ` where id=?`
	cost, err := model.GetQueryTotalCost(query, id)
	if err != nil {
		return 0, nil, err
	}
	res, err := model.GetOneRow(query, id).String()

	return cost, res, err
}

// UpdateSysParam updates the system parameter
func UpdateSysParam(p *Parser, name, value, conditions string) (int64, error) {
	var (
		fields []string
		values []interface{}
	)

	par := &model.SystemParameter{}
	found, err := par.Get(name)
	if !found {
		return 0, errors.New("can't find system parameter: Name " + name)
	}
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

// ValidateCondition checks if the condition can be compiled
func ValidateCondition(condition string, state int64) error {
	if len(condition) == 0 {
		return fmt.Errorf("Conditions cannot be empty")
	}
	return smart.CompileEval(condition, uint32(state))
}

// PrefixTable returns table name with global or state prefix
func PrefixTable(p *Parser, tablename string, global int64) string {
	tablename = converter.Sanitize(tablename, ``)
	if global == 1 {
		return `global_` + tablename
	}
	return StateTable(p, tablename)
}

// EvalCondition gets the condition and check it
func EvalCondition(p *Parser, table, name, condfield string) error {
	conditions, err := model.Single(`SELECT `+converter.EscapeName(condfield)+` FROM `+converter.EscapeName(table)+
		` WHERE name = ?`, name).String()
	if err != nil {
		return err
	}
	if len(conditions) == 0 {
		return fmt.Errorf(`Record %s has not been found`, name)
	}
	return Eval(p, conditions)
}

// Replace replaces old substrings to new substrings
func Replace(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

// FindEcosystem checks if there is an ecosystem with the specified name
func FindEcosystem(p *Parser, country string) (int64, int64, error) {
	query := `SELECT id FROM system_states where name=?`
	cost, err := model.GetQueryTotalCost(query, country)
	if err != nil {
		return 0, 0, err
	}
	id, err := model.Single(query, country).Int64()
	if err != nil {
		return 0, 0, err
	}
	return cost, id, nil
}

// UpdateLang updates language resource
func UpdateLang(p *Parser, name, trans string) {
	language.UpdateLang(int(p.TxStateID), name, trans)
}

// Size returns the length of the string
func Size(s string) int64 {
	return int64(len(s))
}

// Substr returns the substring of the string
func Substr(s string, off int64, slen int64) string {
	ilen := int64(len(s))
	if off < 0 || slen < 0 || off > ilen {
		return ``
	}
	if off+slen > ilen {
		return s[off:]
	}
	return s[off : off+slen]
}

func IsContract(name string, state int64) bool {
	return smart.GetContract(name, uint32(state)) != nil
}

func ContractsList(value string) []interface{} {
	list := smart.ContractsList(value)
	result := make([]interface{}, len(list))
	for i := 0; i < len(list); i++ {
		result[i] = reflect.ValueOf(list[i]).Interface()
	}
	return result
}

func CompileContract(p *Parser, code string, state, id, token int64) (interface{}, error) {
	if p.TxContract.Name != `@1NewContract` && p.TxContract.Name != `@1EditContract` {
		return 0, fmt.Errorf(`CompileContract can be only called from NewContract or EditContract`)
	}
	return smart.CompileBlock(code, &script.OwnerInfo{StateID: uint32(state), WalletID: id, TokenID: token})
}

func FlushContract(p *Parser, iroot interface{}, id int64, active bool) error {
	if p.TxContract.Name != `@1NewContract` && p.TxContract.Name != `@1EditContract` {
		return fmt.Errorf(`FlushContract can be only called from NewContract or EditContract`)
	}
	root := iroot.(*script.Block)
	for i, item := range root.Children {
		if item.Type == script.ObjContract {
			root.Children[i].Info.(*script.ContractInfo).Owner.TableID = id
			root.Children[i].Info.(*script.ContractInfo).Owner.Active = active
		}
	}

	smart.FlushBlock(root)
	return nil
}

// Eval evaluates the condition
func Eval(p *Parser, condition string) error {
	if len(condition) == 0 {
		return fmt.Errorf(`The condition is empty`)
	}
	ret, err := p.EvalIf(condition)
	if err != nil {
		return err
	}
	if !ret {
		return fmt.Errorf(`Access denied`)
	}
	return nil
}

// ActivateContract sets Active status of the contract in smartVM
func ActivateContract(p *Parser, tblid int64, state int64) error {
	if p.TxContract.Name != `@1ActivateContract` {
		return fmt.Errorf(`ActivateContract can be only called from @1ActivateContract`)
	}
	smart.ActivateContract(tblid, state, true)
	return nil
}

// CreateEcosystem creates a new ecosystem
func CreateEcosystem(p *Parser, wallet int64, name string) (int64, error) {
	if p.TxContract.Name != `@1NewEcosystem` {
		return 0, fmt.Errorf(`CreateEcosystem can be only called from @1NewEcosystem`)
	}
	_, id, err := p.selectiveLoggingAndUpd([]string{`name`}, []interface{}{
		name,
	}, `system_states`, nil, nil, true)

	if err != nil {
		return 0, err
	}
	err = model.ExecSchemaEcosystem(converter.StrToInt(id), wallet, name)
	return converter.StrToInt64(id), err
}

func RollbackEcosystem(p *Parser) error {
	if p.TxContract.Name != `@1NewEcosystem` {
		return fmt.Errorf(`RollbackEcosystem can be only called from @1NewEcosystem`)
	}
	rollbackTx := &model.RollbackTx{}
	err := rollbackTx.Get(p.DbTransaction, p.TxHash, "system_states")
	if err != nil {
		return err
	}
	lastID, err := model.GetNextID(`system_states`)
	if err != nil {
		return err
	}
	lastID--
	if converter.StrToInt64(rollbackTx.TableID) != lastID {
		return fmt.Errorf(`Incorrect ecosystem id %s != %d`, rollbackTx.TableID, lastID)
	}
	for _, name := range []string{`menu`, `pages`, `languages`, `signatures`, `tables`,
		`contracts`, `parameters`, `blocks`, `history`, `keys`} {
		err = model.DropTable(p.DbTransaction, fmt.Sprintf("%s_%s", rollbackTx.TableID, name))
		if err != nil {
			return err
		}
	}
	rollbackTxToDel := &model.RollbackTx{TxHash: p.TxHash, NameTable: "system_states"}
	err = rollbackTxToDel.DeleteByHashAndTableName(p.DbTransaction)
	if err != nil {
		return err
	}
	ssToDel := &model.SystemState{ID: lastID}
	return ssToDel.Delete(p.DbTransaction)
}

func TableConditions(p *Parser, name, columns, permissions string) (err error) {
	isEdit := len(columns) == 0

	if isEdit {
		if p.TxContract.Name != `@1EditTable` {
			return fmt.Errorf(`TableConditions can be only called from @1EditTable`)
		}
	} else if p.TxContract.Name != `@1NewTable` {
		return fmt.Errorf(`TableConditions can be only called from @1NewTable`)
	}

	prefix := converter.Int64ToStr(p.TxSmart.StateID)

	t := &model.Table{}
	t.SetTablePrefix(prefix)
	exists, err := t.ExistsByName(name)
	if err != nil {
		return err
	}
	if isEdit {
		if !exists {
			return fmt.Errorf(`table %s doesn't exist`, name)
		}
	} else if exists {
		return fmt.Errorf(`table %s exists`, name)
	}

	var perm map[string]string
	err = json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		return
	}
	if len(perm) != 3 {
		return fmt.Errorf(`Permissions must contain "insert", "new_column", "update"`)
	}
	for _, v := range []string{`insert`, `update`, `new_column`} {
		if len(perm[v]) == 0 {
			return fmt.Errorf(`%v condition is empty`, v)
		}
		if err = smart.CompileEval(perm[v], uint32(p.TxSmart.StateID)); err != nil {
			return err
		}
	}
	if isEdit {
		if err = p.AccessTable(name, `update`); err != nil {
			if err = p.AccessRights(`changing_tables`, false); err != nil {
				return err
			}
		}
		return nil
	}

	var cols []map[string]string
	err = json.Unmarshal([]byte(columns), &cols)
	if err != nil {
		return
	}
	if len(cols) == 0 {
		return fmt.Errorf(`len(cols) == 0`)
	}
	if len(cols) > syspar.GetMaxColumns() {
		return fmt.Errorf(`Too many columns. Limit is %d`, syspar.GetMaxColumns())
	}
	var indexes int
	for _, data := range cols {
		if len(data[`name`]) == 0 || len(data[`type`]) == 0 {
			return fmt.Errorf(`worng column`)
		}
		itype := data[`type`]
		if itype != `varchar` && itype != `number` && itype != `datetime` && itype != `text` &&
			itype != `bytea` && itype != `double` && itype != `money` && itype != `character` {
			return fmt.Errorf(`incorrect type`)
		}
		if len(data[`conditions`]) == 0 {
			return fmt.Errorf(`Conditions is empty`)
		}
		if err = smart.CompileEval(data[`conditions`], uint32(p.TxSmart.StateID)); err != nil {
			return err
		}
		if data[`index`] == `1` {
			if itype != `varchar` && itype != `number` && itype != `datetime` {
				return fmt.Errorf(`incorrect index type`)
			}
			indexes++
		}

	}
	if indexes > syspar.GetMaxIndexes() {
		return fmt.Errorf(`Too many indexes. Limit is %d`, syspar.GetMaxIndexes())
	}

	if err := p.AccessRights("new_table", false); err != nil {
		return err
	}

	return nil
}

func CreateTable(p *Parser, name string, columns, permissions string) error {
	var err error
	if p.TxContract.Name != `@1NewTable` {
		return fmt.Errorf(`CreateTable can be only called from @1NewTable`)
	}
	prefix := converter.Int64ToStr(p.TxSmart.StateID)

	tableName := prefix + "_" + name

	var cols []map[string]string
	err = json.Unmarshal([]byte(columns), &cols)
	if err != nil {
		return err
	}
	indexes := make([]string, 0)

	colsSQL := ""
	//	colsSQL2 := ""
	colperm := make(map[string]string)
	colList := make(map[string]bool)
	for _, data := range cols {
		colname := strings.ToLower(data[`name`])
		if colList[colname] {
			return fmt.Errorf(`There are the same columns`)
		}
		colList[colname] = true
		colType := ``
		colDef := ``
		switch data[`type`] {
		case "varchar":
			colType = `varchar(102400)`
		case "character":
			colType = `character(1)`
			colDef = `NOT NULL DEFAULT '0'`
		case "number":
			colType = `bigint`
			colDef = `NOT NULL DEFAULT '0'`
		case "datetime":
			colType = `timestamp`
		case "double":
			colType = `double precision`
		case "money":
			colType = `decimal (30, 0)`
			colDef = `NOT NULL DEFAULT '0'`
		default:
			colType = data[`type`]
		}
		colsSQL += `"` + colname + `" ` + colType + " " + colDef + " ,\n"
		//colsSQL2 += `"` + data[`name`] + `": "ContractConditions(\"MainCondition\")",`
		colperm[colname] = data[`conditions`]
		if data[`index`] == "1" {
			indexes = append(indexes, colname)
		}
	}
	colout, err := json.Marshal(colperm)
	if err != nil {
		return err
	}
	//	colsSQL2 = colsSQL2[:len(colsSQL2)-1]
	err = model.CreateTable(p.DbTransaction, tableName, colsSQL)
	if err != nil {
		return err
	}

	for _, index := range indexes {
		err := model.CreateIndex(p.DbTransaction, tableName+"_"+index, tableName, index)
		if err != nil {
			return err
		}
	}
	var perm map[string]string
	permlist := make(map[string]string)
	err = json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		return err
	}
	for _, v := range []string{`insert`, `update`, `new_column`} {
		permlist[v] = perm[v]
	}
	permout, err := json.Marshal(permlist)
	if err != nil {
		return err
	}
	t := &model.Table{
		Name:        name,
		Columns:     string(colout),
		Permissions: string(permout),
		Conditions:  `ContractAccess("@1EditTable")`,
	}
	t.SetTablePrefix(prefix)
	err = t.Create(p.DbTransaction)
	if err != nil {
		return err
	}
	return nil
}

func RollbackTable(p *Parser, name string) error {
	if p.TxContract.Name != `@1NewTable` {
		return fmt.Errorf(`RollbackTable can be only called from @1NewTable`)
	}
	err := model.DropTable(p.DbTransaction, fmt.Sprintf("%d_%s", p.TxSmart.StateID, name))
	t := &model.Table{Name: name}
	err = t.Delete()
	if err != nil {
		return err
	}
	return nil
}

func PermTable(p *Parser, name, permissions string) error {
	if p.TxContract.Name != `@1EditTable` {
		return fmt.Errorf(`EditTable can be only called from @1EditTable`)
	}
	var perm map[string]string
	permlist := make(map[string]string)
	err := json.Unmarshal([]byte(permissions), &perm)
	if err != nil {
		return err
	}
	for _, v := range []string{`insert`, `update`, `new_column`} {
		permlist[v] = perm[v]
	}
	permout, err := json.Marshal(permlist)
	if err != nil {
		return err
	}
	_, _, err = p.selectiveLoggingAndUpd([]string{`permissions`}, []interface{}{string(permout)},
		fmt.Sprintf(`%d_tables`, p.TxSmart.StateID), []string{`name`}, []string{name}, true)
	return err
}

func ColumnCondition(p *Parser, tableName, name, coltype, permissions, index string) error {
	if p.TxContract.Name != `@1NewColumn` && p.TxContract.Name != `@1EditColumn` {
		return fmt.Errorf(`ColumnCondition can be only called from @1NewColumn`)
	}
	isExist := p.TxContract.Name == `@1EditColumn`
	tEx := &model.Table{}
	tEx.SetTablePrefix(converter.Int64ToStr(p.TxSmart.StateID))
	name = strings.ToLower(name)

	exists, err := tEx.IsExistsByPermissionsAndTableName(name, tableName)
	if err != nil {
		return err
	}
	if isExist {
		if !exists {
			return fmt.Errorf(`column %s doesn't exists`, name)
		}
	} else if exists {
		return fmt.Errorf(`column %s exists`, name)
	}
	if len(permissions) == 0 {
		return fmt.Errorf(`Permissions is empty`)
	}
	if err = smart.CompileEval(permissions, uint32(p.TxSmart.StateID)); err != nil {
		return err
	}
	tblName := fmt.Sprintf("%d_%s", p.TxSmart.StateID, tableName)
	if isExist {
		return p.AccessTable(tblName, `update`)
	}
	count, err := model.GetColumnCount(tblName)
	if count >= int64(syspar.GetMaxColumns()) {
		return fmt.Errorf(`Too many columns. Limit is %d`, syspar.GetMaxColumns())
	}
	if coltype != `varchar` && coltype != `number` && coltype != `datetime` && coltype != `character` &&
		coltype != `text` && coltype != `bytea` && coltype != `double` && coltype != `money` {
		return fmt.Errorf(`incorrect type`)
	}
	if index == `1` {
		count, err := model.NumIndexes(tblName)
		if err != nil {
			return err
		}
		if count >= syspar.GetMaxIndexes() {
			return fmt.Errorf(`Too many indexes. Limit is %d`, syspar.GetMaxIndexes())
		}
		if coltype != `varchar` && coltype != `number` && coltype != `datetime` {
			return fmt.Errorf(`incorrect index type`)
		}
	}

	if err := p.AccessTable(tblName, "new_column"); err != nil {
		return err
	}
	return nil
}

func CreateColumn(p *Parser, tableName, name, coltype, permissions, index string) error {
	if p.TxContract.Name != `@1NewColumn` {
		return fmt.Errorf(`CreateColumn can be only called from @1NewColumn`)
	}
	name = strings.ToLower(name)
	tblname := fmt.Sprintf(`%d_%s`, p.TxSmart.StateID, tableName)

	colType := ``
	switch coltype {
	case "varchar":
		colType = `varchar(102400)`
	case "number":
		colType = `bigint NOT NULL DEFAULT '0'`
	case "character":
		colType = `character(1) NOT NULL DEFAULT '0'`
	case "datetime":
		colType = `timestamp`
	case "double":
		colType = `double precision`
	case "money":
		colType = `decimal (30, 0) NOT NULL DEFAULT '0'`
	default:
		colType = coltype
	}
	err := model.AlterTableAddColumn(p.DbTransaction, tblname, name, colType)
	if err != nil {
		return err
	}

	if index == "1" {
		err = model.CreateIndex(p.DbTransaction, tblname+"_"+name+"_index", tblname, name)
		if err != nil {
			return err
		}
	}
	tables := fmt.Sprintf(`%d_tables`, p.TxSmart.StateID)
	type cols struct {
		Columns string
	}
	temp := &cols{}
	err = model.DBConn.Table(tables).Where("name = ?", tableName).Select("columns").Find(temp).Error
	if err != nil {
		return err
	}
	var perm map[string]string
	err = json.Unmarshal([]byte(temp.Columns), &perm)
	if err != nil {
		return err
	}
	perm[name] = permissions
	permout, err := json.Marshal(perm)
	if err != nil {
		return err
	}
	_, _, err = p.selectiveLoggingAndUpd([]string{`columns`}, []interface{}{string(permout)},
		tables, []string{`name`}, []string{tableName}, true)
	return nil
}

func RollbackColumn(p *Parser, tableName, name string) error {
	if p.TxContract.Name != `@1NewColumn` {
		return fmt.Errorf(`RollbackColumn can be only called from @1NewColumn`)
	}
	return model.AlterTableDropColumn(fmt.Sprintf(`%d_%s`, p.TxSmart.StateID, tableName), name)
}

func PermColumn(p *Parser, tableName, name, permissions string) error {
	if p.TxContract.Name != `@1EditColumn` {
		return fmt.Errorf(`EditColumn can be only called from @1EditColumn`)
	}
	name = strings.ToLower(name)
	tables := fmt.Sprintf(`%d_tables`, p.TxSmart.StateID)
	type cols struct {
		Columns string
	}
	temp := &cols{}
	err := model.DBConn.Table(tables).Where("name = ?", tableName).Select("columns").Find(temp).Error
	if err != nil {
		return err
	}
	var perm map[string]string
	err = json.Unmarshal([]byte(temp.Columns), &perm)
	if err != nil {
		return err
	}
	perm[name] = permissions
	permout, err := json.Marshal(perm)
	if err != nil {
		return err
	}
	_, _, err = p.selectiveLoggingAndUpd([]string{`columns`}, []interface{}{string(permout)},
		tables, []string{`name`}, []string{tableName}, true)
	return err
}
