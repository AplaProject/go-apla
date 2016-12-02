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
	"fmt"
	"strconv"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

func init() {
	smart.Extend(&script.ExtendData{map[string]interface{}{
		"DBInsert":    DBInsert,
		"DBUpdate":    DBUpdate,
		"DBTransfer":  DBTransfer,
		"DBString":    DBString,
		"DBInt":       DBInt,
		"DBStringExt": DBStringExt,
		"DBIntExt":    DBIntExt,
		"Table":       StateTable,
		"TableTx":     StateTableTx,
		"AddressToId": AddressToID,
		"DBAmount":    DBAmount,
		"IsContract":  IsContract,
		"StateValue":  StateValue,
		"Int":         Int,
	}, map[string]string{
		`*parser.Parser`: `parser`,
	}})
	//	smart.Compile( embedContracts)
}

func (p *Parser) getExtend() *map[string]interface{} {
	head := p.TxPtr.(*consts.TXHeader) //consts.HeaderNew(contract.parser.TxPtr)
	var citizenId, walletId int64
	citizenId = int64(head.WalletId)
	walletId = int64(head.WalletId)
	// test
	block := int64(0)
	walletBlock := int64(0)
	if p.BlockData != nil {
		block = p.BlockData.BlockId
		walletBlock = p.BlockData.WalletId
	}

	extend := map[string]interface{}{`type`: head.Type, `time`: head.Type, `state`: int64(head.StateId),
		`block`: block, `citizen`: citizenId, `wallet`: walletId, `wallet_block`: walletBlock,
		`parser`: p, `contract`: p.TxContract}
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

	methods := []string{`init`, `front`, `main`, `rollback`}
	p.TxContract.Extend = p.getExtend()
	for i := uint32(0); i < 4; i++ {
		if (flags & (1 << i)) > 0 {
			cfunc := p.TxContract.GetFunc(methods[i])
			if cfunc == nil {
				continue
			}
			p.TxContract.Called = 1 << i
			_, err = smart.Run(cfunc, nil, p.TxContract.Extend)
			if err != nil {
				//			fmt.Println(`Contract Error`, err)
				return
			}
		}
	}
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

func DBInt(tblname string, name string, id int64) (int64, error) {
	return utils.DB.Single(`select `+lib.EscapeName(name)+` from `+lib.EscapeName(tblname)+` where id=?`, id).Int64()
}

func DBStringExt(tblname string, name string, id int64, idname string) (string, error) {
	if isIndex, err := utils.DB.IsIndex(tblname, idname); err != nil {
		return ``, err
	} else if !isIndex {
		return ``, fmt.Errorf(`there is not index on %s`, idname)
	}
	return utils.DB.Single(`select `+lib.EscapeName(name)+` from `+lib.EscapeName(tblname)+` where `+
		lib.EscapeName(idname)+`=?`, id).String()
}

func DBIntExt(tblname string, name string, id int64, idname string) (ret int64, err error) {
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

func StateTable(p *Parser, tblname string) string {
	return fmt.Sprintf("%d_%s", p.TxStateID, tblname)
}

func StateTableTx(p *Parser, tblname string) string {
	return fmt.Sprintf("%v_%s", p.TxData[`StateId`], tblname)
}

func IsContract(p *Parser, name string) bool {
	if p.TxContract != nil {
		return p.TxContract.Name == name
	} else if len(p.TxSlice) > 1 {
		return consts.TxTypes[utils.BytesToInt(p.TxSlice[1])] == name
	}
	return false
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

func DBAmount(tblname, column string, id int64) decimal.Decimal {
	balance, err := utils.DB.Single("SELECT amount FROM "+lib.EscapeName(tblname)+" WHERE "+lib.EscapeName(column)+" = ?", id).String()
	if err != nil {
		return decimal.New(0, 0)
	}
	val, _ := decimal.NewFromString(balance)
	return val
}

func (p *Parser) EvalIf(conditions string) (bool, error) {
	return smart.EvalIf(conditions, &map[string]interface{}{`state`: p.TxStateID,
		`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID, `parser`: p})
}

func StateValue(p *Parser, name string) string {
	val, _ := utils.StateParam(int64(p.TxStateID), name)
	return val
}

func Int(val string) int64 {
	return utils.StrToInt64(val)
}
