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
		"DBInsert":   DBInsert,
		"DBUpdate":   DBUpdate,
		"DBTransfer": DBTransfer,
		"DBString":   DBString,
		"DBInt":      DBInt,
		"Table":      StateTable,
		"TableTx":    StateTableTx,
	}, map[string]string{
		`*parser.Parser`: `parser`,
	}})
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

	if len(p.PublicKeys) == 0 {
		data, err := p.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM dlt_wallets WHERE wallet_id = ?",
			int64(p.TxPtr.(*consts.TXHeader).WalletId)).String()
		//	fmt.Println(`HASH`, p.TxHash)
		//	fmt.Println(`TX Call DATA`, p.TxPtr.(*consts.TXHeader).WalletId, err, data)

		if err != nil {
			return err
		}
		if len(data["public_key_0"]) == 0 {
			return fmt.Errorf("unknown wallet id")
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

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.TxData[`forsign`].(string), p.TxPtr.(*consts.TXHeader).Sign, false)
	//	fmt.Println(`Forsign`, p.TxData[`forsign`], CheckSignResult, err)
	if err != nil {
		return err
	}
	if !CheckSignResult {
		return fmt.Errorf("incorrect sign")
	}

	methods := []string{`init`, `front`, `main`}
	p.TxContract.Extend = p.getExtend()
	for i := uint32(0); i < 3; i++ {
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

func StateTable(p *Parser, tblname string) string {
	return fmt.Sprintf("%d_%s", p.TxStateID, tblname)
}

func StateTableTx(p *Parser, tblname string) string {
	return fmt.Sprintf("%v_%s", p.TxData[`StateId`], tblname)
}
