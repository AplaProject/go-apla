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
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/controllers"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

var (
	extendCost = map[string]int64{
		"DBInsert":       200,
		"DBUpdate":       100,
		"DBUpdateWhere":  100,
		"DBGetList":      300,
		"DBGetTable":     1000,
		"DBTransfer":     200,
		"DBString":       100,
		"DBInt":          100,
		"DBStringExt":    100,
		"DBIntExt":       100,
		"DBStringWhere":  100,
		"DBIntWhere":     100,
		"AddressToId":    10,
		"IdToAddress":    10,
		"DBAmount":       100,
		"IsGovAccount":   80,
		"StateValue":     80,
		"Sha256":         50,
		"PubToID":        10,
		"UpdateContract": 200,
		"UpdateParam":    200,
		"UpdateMenu":     200,
		"UpdatePage":     200,
	}
)

func init() {
	smart.Extend(&script.ExtendData{map[string]interface{}{
		//		"CallContract":   ExContract,
		"DBInsert":        DBInsert,
		"DBUpdate":        DBUpdate,
		"DBUpdateWhere":   DBUpdateWhere,
		"DBGetList":       DBGetList,
		"DBGetTable":      DBGetTable,
		"DBTransfer":      DBTransfer,
		"DBString":        DBString,
		"DBInt":           DBInt,
		"DBStringExt":     DBStringExt,
		"DBIntExt":        DBIntExt,
		"DBStringWhere":   DBStringWhere,
		"DBIntWhere":      DBIntWhere,
		"Table":           StateTable,
		"TableTx":         StateTableTx,
		"AddressToId":     AddressToID,
		"IdToAddress":     IDToAddress,
		"DBAmount":        DBAmount,
		"ContractAccess":  IsContract,
		"IsGovAccount":    IsGovAccount,
		"StateValue":      StateValue,
		"Int":             Int,
		"Str":             Str,
		"Money":           Money,
		"Float":           Float,
		"Len":             Len,
		"Sha256":          Sha256,
		"PubToID":         PubToID,
		"HexToBytes":      HexToBytes,
		"UpdateContract":  UpdateContract,
		"UpdateParam":     UpdateParam,
		"UpdateMenu":      UpdateMenu,
		"UpdatePage":      UpdatePage,
		"check_signature": CheckSignature, // system function
	}, map[string]string{
		`*parser.Parser`: `parser`,
	}})
	smart.ExtendCost(getCost)
	//	smart.Compile( embedContracts)
}

func getCost(name string) int64 {
	if val, ok := extendCost[name]; ok {
		return val
	}
	return 0
}

func (p *Parser) getExtend() *map[string]interface{} {
	head := p.TxPtr.(*consts.TXHeader) //consts.HeaderNew(contract.parser.TxPtr)
	var citizenId, walletId int64
	citizenId = int64(head.WalletId)
	walletId = int64(head.WalletId)
	// test
	block := int64(0)
	blockTime := int64(0)
	walletBlock := int64(0)
	if p.BlockData != nil {
		block = p.BlockData.BlockId
		walletBlock = p.BlockData.WalletId
		blockTime = p.BlockData.Time
	}
	extend := map[string]interface{}{`type`: head.Type, `time`: int64(head.Time), `state`: int64(head.StateId),
		`block`: block, `citizen`: citizenId, `wallet`: walletId, `wallet_block`: walletBlock,
		`parent`: ``, `txcost`: p.GetContractLimit(),
		`parser`: p, `contract`: p.TxContract, `block_time`: blockTime /*, `vars`: make(map[string]interface{})*/}
	for key, val := range p.TxData {
		extend[key] = val
	}
	/*	v := reflect.ValueOf(contract.parser.TxPtr).Elem()
		t := v.Type()
		for i := 1; i < t.NumField(); i++ {
			extend[t.Field(i).Name] = v.Field(i).Interface()
		}*/
	//	fmt.Println(`Extend`, extend)
	return &extend
}

func StackCont(p interface{}, name string) {
	cont := p.(*Parser).TxContract
	if len(name) > 0 {
		cont.StackCont = append(cont.StackCont, name)
	} else {
		cont.StackCont = cont.StackCont[:len(cont.StackCont)-1]
	}
	return
}

func (p *Parser) CallContract(flags int) (err error) {
	var public []byte
	if flags&smart.CALL_ROLLBACK == 0 {
		if p.TxPtr.(*consts.TXHeader).Flags&consts.TxfPublic > 0 {
			public = p.TxPtr.(*consts.TXHeader).Sign[len(p.TxPtr.(*consts.TXHeader).Sign)-64:]
			p.TxPtr.(*consts.TXHeader).Sign = p.TxPtr.(*consts.TXHeader).Sign[:len(p.TxPtr.(*consts.TXHeader).Sign)-64]
		}
		if len(p.PublicKeys) == 0 {
			data, err := p.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM dlt_wallets WHERE wallet_id = ?",
				int64(p.TxPtr.(*consts.TXHeader).WalletId)).String()
			if err != nil {
				return err
			}
			if len(data["public_key_0"]) == 0 {
				if len(public) > 0 {
					p.PublicKeys = append(p.PublicKeys, public)
				} else {
					return fmt.Errorf("unknown wallet id")
				}
			} else {
				p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_0"]))
				/*		if len(data["public_key_1"]) > 10 {
							p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_1"]))
						}
						if len(data["public_key_2"]) > 10 {
							p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_2"]))
						}*/
			}
		}
		/*	fmt.Printf("TXPublic=%x %d\r\n", p.PublicKeys[0], len(p.PublicKeys[0]))
			fmt.Printf("TXSign=%x %d\r\n", p.TxPtr.(*consts.TXHeader).Sign, len(p.TxPtr.(*consts.TXHeader).Sign))
			fmt.Printf("TXForSign=%s %d\r\n", p.TxData[`forsign`].(string), len(p.TxData[`forsign`].(string)))
		*/
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.TxData[`forsign`].(string), p.TxPtr.(*consts.TXHeader).Sign, false)
		//	fmt.Println(`Forsign`, p.TxData[`forsign`], CheckSignResult, err)
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
	if p.GetFuel() <= 0 {
		return fmt.Errorf(`Fuel rate must be greater than 0`)
	}
	if (flags&smart.CALL_MAIN) > 0 && !p.CheckContractLimit(price) {
		return fmt.Errorf(`there are not enough money`)
	}

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
	p.TxUsedCost = before - (*p.TxContract.Extend)[`txcost`].(int64)
	p.TxContract.TxPrice = price
	//fmt.Println(`Cost`, p.TxUsedCost)
	return
}

func DBInsert(p *Parser, tblname string, params string, val ...interface{}) (ret int64, err error) { // map[string]interface{}) {
	//	fmt.Println(`DBInsert`, tblname, params, val, len(val))
	if err = p.AccessTable(tblname, "insert"); err != nil {
		return
	}
	var lastId string
	lastId, err = p.selectiveLoggingAndUpd(strings.Split(params, `,`), val, tblname, nil, nil, true)
	if err == nil {
		ret, _ = strconv.ParseInt(lastId, 10, 64)
	}
	return
}

func DBUpdate(p *Parser, tblname string, id int64, params string, val ...interface{}) (err error) { // map[string]interface{}) {
	//	fmt.Println(`DBUpdate`, tblname, id, params, val, len(val))
	/*	if err = p.AccessTable(tblname, "general_update"); err != nil {
		return
	}*/
	columns := strings.Split(params, `,`)
	if err = p.AccessColumns(tblname, columns); err != nil {
		return
	}
	_, err = p.selectiveLoggingAndUpd(columns, val, tblname, []string{`id`}, []string{utils.Int64ToStr(id)}, true)
	return
}

func DBUpdateWhere(p *Parser, tblname string, column string, value interface{}, params string, val ...interface{}) (err error) { // map[string]interface{}) {
	var isIndex bool
	columns := strings.Split(params, `,`)
	if err = p.AccessColumns(tblname, columns); err != nil {
		return
	}
	if isIndex, err = utils.DB.IsIndex(tblname, column); err != nil {
		return
	} else if !isIndex {
		err = fmt.Errorf(`there is not index on %s`, column)
	} else {
		_, err = p.selectiveLoggingAndUpd(columns, val, tblname, []string{column}, []string{fmt.Sprint(value)}, true)
	}
	return
}

func DBTransfer(p *Parser, tblname, columns string, idFrom, idTo int64, amount decimal.Decimal) (err error) { // map[string]interface{}) {
	cols := strings.Split(columns, `,`)
	idname := `id`
	if len(cols) == 2 {
		idname = cols[1]
	}
	column := cols[0]
	if err = p.AccessColumns(tblname, []string{column}); err != nil {
		return
	}
	value := amount.String()

	if _, err = p.selectiveLoggingAndUpd([]string{`-` + column}, []interface{}{value}, tblname, []string{idname},
		[]string{utils.Int64ToStr(idFrom)}, true); err != nil {
		return
	}
	if _, err = p.selectiveLoggingAndUpd([]string{`+` + column}, []interface{}{value}, tblname, []string{idname},
		[]string{utils.Int64ToStr(idTo)}, true); err != nil {
		return
	}
	return
}

func DBString(tblname string, name string, id int64) (string, error) {
	return utils.DB.Single(`select `+lib.EscapeName(name)+` from `+lib.EscapeName(tblname)+` where id=?`, id).String()
}

func Sha256(text string) string {
	return string(utils.Sha256(text))
}

func PubToID(hexkey string) int64 {
	pubkey, err := hex.DecodeString(hexkey)
	if err != nil {
		return 0
	}
	return int64(lib.Address(pubkey))
}

func HexToBytes(hexdata string) ([]byte, error) {
	return hex.DecodeString(hexdata)
}

func DBInt(tblname string, name string, id int64) (int64, error) {
	return utils.DB.Single(`select `+lib.EscapeName(name)+` from `+lib.EscapeName(tblname)+` where id=?`, id).Int64()
}

func DBStringExt(tblname string, name string, id interface{}, idname string) (string, error) {
	if isIndex, err := utils.DB.IsIndex(tblname, idname); err != nil {
		return ``, err
	} else if !isIndex {
		return ``, fmt.Errorf(`there is not index on %s`, idname)
	}
	return utils.DB.Single(`select `+lib.EscapeName(name)+` from `+lib.EscapeName(tblname)+` where `+
		lib.EscapeName(idname)+`=?`, id).String()
}

func DBIntExt(tblname string, name string, id interface{}, idname string) (ret int64, err error) {
	var val string
	val, err = DBStringExt(tblname, name, id, idname)
	if err != nil {
		return 0, err
	}
	if len(val) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(val, 10, 64)
}

func DBStringWhere(tblname string, name string, where string, params ...interface{}) (string, error) {
	re := regexp.MustCompile(`([a-z]+[\w_]*)\"?\s*[><=]`)
	ret := re.FindAllStringSubmatch(where, -1)
	for _, iret := range ret {
		if len(iret) != 2 {
			continue
		}
		if isIndex, err := utils.DB.IsIndex(tblname, iret[1]); err != nil {
			return ``, err
		} else if !isIndex {
			return ``, fmt.Errorf(`there is not index on %s`, iret[1])
		}
	}
	return utils.DB.Single(`select `+lib.EscapeName(name)+` from `+lib.EscapeName(tblname)+` where `+
		strings.Replace(lib.Escape(where), `$`, `?`, -1), params...).String()
}

func DBIntWhere(tblname string, name string, where string, params ...interface{}) (ret int64, err error) {
	var val string
	val, err = DBStringWhere(tblname, name, where, params...)
	if err != nil {
		return 0, err
	}
	if len(val) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(val, 10, 64)
}

func StateTable(p *Parser, tblname string) string {
	return fmt.Sprintf("%d_%s", p.TxStateID, tblname)
}

func StateTableTx(p *Parser, tblname string) string {
	return fmt.Sprintf("%v_%s", p.TxData[`StateId`], tblname)
}

func IsContract(p *Parser, name string) bool {
	if p.TxContract != nil && len(name) > 0 {
		if name[0] != '@' {
			name = fmt.Sprintf(`@%d`, p.TxStateID) + name
		}
		//		return p.TxContract.Name == name
		return p.TxContract.StackCont[len(p.TxContract.StackCont)-1] == name
	} else if len(p.TxSlice) > 1 {
		return consts.TxTypes[utils.BytesToInt(p.TxSlice[1])] == name
	}
	return false
}

func IsGovAccount(p *Parser, citizen int64) bool {
	return utils.StrToInt64(StateValue(p, `gov_account`)) == citizen
}

func AddressToID(input string) (addr int64) {
	input = strings.TrimSpace(input)
	if len(input) < 2 {
		return 0
	}
	if input[0] == '-' {
		addr, _ = strconv.ParseInt(input, 10, 64)
	} else if strings.Count(input, `-`) == 4 {
		addr = lib.StringToAddress(input)
	} else {
		uaddr, _ := strconv.ParseUint(input, 10, 64)
		addr = int64(uaddr)
	}
	if !lib.IsValidAddress(lib.AddressToString(uint64(addr))) {
		return 0
	}
	return
}

func IDToAddress(id int64) (out string) {
	out = lib.AddressToString(uint64(id))
	if !lib.IsValidAddress(out) {
		out = `invalid`
	}
	return
}

func DBAmount(tblname, column string, id int64) decimal.Decimal {
	balance, err := utils.DB.Single("SELECT amount FROM "+lib.EscapeName(tblname)+" WHERE "+lib.EscapeName(column)+" = ?", id).String()
	if err != nil {
		return decimal.New(0, 0)
	}
	val, _ := decimal.NewFromString(balance)
	return val
}

func (p *Parser) EvalIf(conditions string) (bool, error) {
	time := int64(0)
	if p.TxPtr != nil {
		time = int64(p.TxPtr.(*consts.TXHeader).Time)
	}
	blockTime := int64(0)
	if p.BlockData != nil {
		blockTime = p.BlockData.Time
	}

	return smart.EvalIf(conditions, utils.Int64ToStr(int64(p.TxStateID)), &map[string]interface{}{`state`: p.TxStateID,
		`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID, `parser`: p,
		`block_time`: blockTime, `time`: time})
}

func StateValue(p *Parser, name string) string {
	val, _ := utils.StateParam(int64(p.TxStateID), name)
	return val
}

func Int(val string) int64 {
	return utils.StrToInt64(val)
}

func Str(v interface{}) (ret string) {
	switch val := v.(type) {
	case float64:
		ret = fmt.Sprintf(`%f`, val)
	default:
		ret = fmt.Sprintf(`%v`, val)
	}
	return
}

func Money(v interface{}) (ret decimal.Decimal) {
	return script.ValueToDecimal(v)
}

func Float(v interface{}) (ret float64) {
	return script.ValueToFloat(v)
}

func UpdateContract(p *Parser, name, value, conditions string) error {
	var (
		fields []string
		values []interface{}
	)
	prefix := utils.Int64ToStr(int64(p.TxStateID))
	cnt, err := p.OneRow(`SELECT id,conditions FROM "`+prefix+`_smart_contracts" WHERE name = ?`, name).String()
	if err != nil {
		return err
	}
	if len(cnt) == 0 {
		return fmt.Errorf(`unknown contract %s`, name)
	}
	cond := cnt[`conditions`]
	if len(cond) > 0 {
		ret, err := p.EvalIf(cond)
		if err != nil {
			return err
		}
		if !ret {
			if err = p.AccessRights(`changing_smart_contracts`, false); err != nil {
				return err
			}
		}
	}
	if len(value) > 0 {
		fields = append(fields, "value")
		values = append(values, value)
	}
	if len(conditions) > 0 {
		if err := smart.CompileEval(conditions, p.TxStateID); err != nil {
			return err
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	if len(fields) == 0 {
		return fmt.Errorf(`empty value and condition`)
	}
	root, err := smart.CompileBlock(value, prefix)
	if err != nil {
		return err
	}
	_, err = p.selectiveLoggingAndUpd(fields, values,
		prefix+"_smart_contracts", []string{"id"}, []string{cnt["id"]}, true)
	if err != nil {
		return err
	}
	smart.FlushBlock(root)

	return nil
}

func UpdateParam(p *Parser, name, value, conditions string) error {
	var (
		fields []string
		values []interface{}
	)

	if err := p.AccessRights(name, true); err != nil {
		return err
	}
	if len(value) > 0 {
		fields = append(fields, "value")
		values = append(values, value)
	}
	if len(conditions) > 0 {
		if err := smart.CompileEval(conditions, uint32(p.TxStateID)); err != nil {
			return err
		}
		fields = append(fields, "conditions")
		values = append(values, conditions)
	}
	if len(fields) == 0 {
		return fmt.Errorf(`empty value and condition`)
	}
	_, err := p.selectiveLoggingAndUpd(fields, values,
		utils.Int64ToStr(int64(p.TxStateID))+"_state_parameters", []string{"name"}, []string{name}, true)
	if err != nil {
		return err
	}
	return nil
}

func UpdateMenu(p *Parser, name, value, conditions string) error {
	if err := p.AccessChange(`menu`, name); err != nil {
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
	_, err := p.selectiveLoggingAndUpd(fields, values, utils.Int64ToStr(int64(p.TxStateID))+"_menu",
		[]string{"name"}, []string{name}, true)
	if err != nil {
		return err
	}
	return nil
}

func CheckSignature(i *map[string]interface{}, name string) error {
	state, name := script.ParseContract(name)
	pref := utils.Int64ToStr(int64(state))
	if state == 0 {
		pref = `global`
	}
	//	fmt.Println(`CheckSignature`, i, state, name)
	p := (*i)[`parser`].(*Parser)
	value, err := p.Single(`select value from "`+pref+`_signatures" where name=?`, name).String()
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

	var sign controllers.TxSignJson
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
	fmt.Println(`Check signature`, CheckSignResult, forsign)
	if !CheckSignResult {
		return fmt.Errorf(`incorrect signature ` + forsign)
	}
	return nil
}

func UpdatePage(p *Parser, name, value, menu, conditions string) error {
	if err := p.AccessChange(`pages`, name); err != nil {
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
	_, err := p.selectiveLoggingAndUpd(fields, values, utils.Int64ToStr(int64(p.TxStateID))+"_pages",
		[]string{"name"}, []string{name}, true)
	if err != nil {
		return err
	}

	return nil
}

func Len(in []interface{}) int64 {
	return int64(len(in))
}

func checkWhere(tblname string, where string, order string) (string, string, error) {
	re := regexp.MustCompile(`([a-z]+[\w_]*)\"?\s*[><=]`)
	ret := re.FindAllStringSubmatch(where, -1)

	for _, iret := range ret {
		if len(iret) != 2 {
			continue
		}
		if isIndex, err := utils.DB.IsIndex(tblname, iret[1]); err != nil {
			return ``, ``, err
		} else if !isIndex {
			return ``, ``, fmt.Errorf(`there is not index on %s`, iret[1])
		}
	}
	if len(order) > 0 {
		order = ` order by ` + lib.EscapeName(order)
	}
	return strings.Replace(lib.Escape(where), `$`, `?`, -1), order, nil
}

func DBGetList(tblname string, name string, offset, limit int64, order string,
	where string, params ...interface{}) ([]interface{}, error) {
	re := regexp.MustCompile(`([a-z]+[\w_]*)\"?\s*[><=]`)
	ret := re.FindAllStringSubmatch(where, -1)

	for _, iret := range ret {
		if len(iret) != 2 {
			continue
		}
		if isIndex, err := utils.DB.IsIndex(tblname, iret[1]); err != nil {
			return nil, err
		} else if !isIndex {
			return nil, fmt.Errorf(`there is not index on %s`, iret[1])
		}
	}
	if len(order) > 0 {
		order = ` order by ` + lib.EscapeName(order)
	}
	if limit <= 0 {
		limit = -1
	}
	list, err := utils.DB.GetAll(`select `+lib.Escape(name)+` from `+lib.EscapeName(tblname)+` where `+
		strings.Replace(lib.Escape(where), `$`, `?`, -1)+order+fmt.Sprintf(` offset %d `, offset), int(limit), params...)
	result := make([]interface{}, len(list))
	for i := 0; i < len(list); i++ {
		result[i] = reflect.ValueOf(list[i]).Interface()
	}
	return result, err
}

func DBGetTable(tblname string, columns string, offset, limit int64, order string,
	where string, params ...interface{}) ([]interface{}, error) {
	var err error

	where, order, err = checkWhere(tblname, where, order)
	if limit <= 0 {
		limit = -1
	}
	cols := strings.Split(lib.Escape(columns), `,`)
	list, err := utils.DB.GetAll(`select `+strings.Join(cols, `,`)+` from `+lib.EscapeName(tblname)+` where `+
		where+order+fmt.Sprintf(` offset %d `, offset), int(limit), params...)
	result := make([]interface{}, len(list))
	for i := 0; i < len(list); i++ {
		//result[i] = make(map[string]interface{})
		result[i] = reflect.ValueOf(list[i]).Interface()
		/*		for _, key := range cols {
				result[i][key] = reflect.ValueOf(list[i][key]).Interface()
			}*/
	}
	return result, err
}
