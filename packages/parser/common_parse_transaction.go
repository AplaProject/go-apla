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
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"

	"github.com/shopspring/decimal"
)

func (p *Parser) ParseTransaction(transactionBinaryData *[]byte) ([][]byte, error) {

	var returnSlice [][]byte
	var transSlice [][]byte
	log.Debug("transactionBinaryData: %x", *transactionBinaryData)
	log.Debug("transactionBinaryData: %s", *transactionBinaryData)
	p.TxContract = nil
	p.TxPtr = nil
	p.PublicKeys = nil
	if len(*transactionBinaryData) > 0 {

		// хэш транзакции
		transSlice = append(transSlice, utils.DSha256(*transactionBinaryData))
		input := (*transactionBinaryData)[:]
		// первый байт - тип транзакции
		txType := utils.BinToDecBytesShift(transactionBinaryData, 1)
		isStruct := consts.IsStruct(int(txType))
		if txType > 127 { // транзакция с контрактом
			var err error
			p.TxPtr = &consts.TXHeader{}
			if err = lib.BinUnmarshal(&input, p.TxPtr); err != nil {
				return nil, err
			}
			isStruct = false
			p.TxStateID = uint32(p.TxPtr.(*consts.TXHeader).StateID)
			p.TxStateIDStr = utils.UInt32ToStr(p.TxStateID)
			if p.TxStateID > 0 {
				p.TxCitizenID = int64(p.TxPtr.(*consts.TXHeader).WalletID)
				p.TxWalletID = 0
			} else {
				p.TxCitizenID = 0
				p.TxWalletID = int64(p.TxPtr.(*consts.TXHeader).WalletID)
			}
			contract := smart.GetContractByID(p.TxPtr.(*consts.TXHeader).Type)
			if contract == nil {
				return nil, fmt.Errorf(`unknown contract %d`, p.TxPtr.(*consts.TXHeader).Type)
			}
			//			log.Debug(`TRANDEB %d %d NAME: %s`, int64(p.TxPtr.(*consts.TXHeader).WalletId),
			//				uint64(p.TxPtr.(*consts.TXHeader).WalletId), contract.Name)
			forsign := fmt.Sprintf("%d,%d,%d,%d,%d", p.TxPtr.(*consts.TXHeader).Type,
				p.TxPtr.(*consts.TXHeader).Time, p.TxPtr.(*consts.TXHeader).WalletID,
				p.TxPtr.(*consts.TXHeader).StateID, p.TxPtr.(*consts.TXHeader).Flags)

			p.TxContract = contract
			p.TxData = make(map[string]interface{})
			if contract.Block.Info.(*script.ContractInfo).Tx != nil {
				for _, fitem := range *contract.Block.Info.(*script.ContractInfo).Tx {
					var v interface{}
					switch fitem.Type.String() {
					case `uint64`:
						var val uint64
						lib.BinUnmarshal(&input, &val)
						v = val
					case `float64`:
						var val float64
						lib.BinUnmarshal(&input, &val)
						v = val
					case `int64`:
						v, err = lib.DecodeLenInt64(&input)
					case script.Decimal:
						var s string
						if err = lib.BinUnmarshal(&input, &s); err != nil {
							return nil, err
						}
						v, err = decimal.NewFromString(s)
					case `string`:
						var s string
						if err = lib.BinUnmarshal(&input, &s); err != nil {
							return nil, err
						}
						v = s
					case `[]uint8`:
						var b []byte
						if err = lib.BinUnmarshal(&input, &b); err != nil {
							return nil, err
						}
						v = hex.EncodeToString(b)
					}
					p.TxData[fitem.Name] = v
					if err != nil {
						return nil, err
					}
					if strings.Index(fitem.Tags, `image`) >= 0 {
						continue
					}
					forsign += fmt.Sprintf(",%v", v)
				}
			}
			p.TxData[`forsign`] = forsign
			//			fmt.Println(`Contract data`, p.TxData)
		} else if isStruct {
			p.TxPtr = consts.MakeStruct(consts.TxTypes[int(txType)])
			if err := lib.BinUnmarshal(&input, p.TxPtr); err != nil {
				return nil, err
			}
			p.TxVars = make(map[string]string)
			if int(txType) == 4 { // TXNewCitizen
				head := consts.HeaderNew(p.TxPtr)
				p.TxStateID = uint32(head.StateID)
				p.TxStateIDStr = utils.UInt32ToStr(p.TxStateID)
				if head.StateID > 0 {
					p.TxCitizenID = int64(head.WalletID)
					p.TxWalletID = 0
				} else {
					p.TxCitizenID = 0
					p.TxWalletID = int64(head.WalletID)
				}
				p.TxTime = int64(head.Time)
			} else {
				head := consts.Header(p.TxPtr)
				p.TxCitizenID = head.CitizenID
				p.TxWalletID = head.WalletID
				p.TxTime = int64(head.Time)
			}
			fmt.Println(`PARSED STRUCT %v`, p.TxPtr)
		}
		parser := GetParser(p, consts.TxTypes[int(txType)])
		err := parser.Init()
		if err != nil {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect tx"))
		}
		header := parser.Header()
		if header == nil {
			return transSlice, utils.ErrInfo(fmt.Errorf("tx header is nil"))
		}
		transSlice = append(transSlice, utils.Int64ToByte(txType))
		// следующие 4 байта - время транзакции
		transSlice = append(transSlice, utils.Int64ToByte(header.Time))
		// преобразуем бинарные данные транзакции в массив
		if txType > 127 {
			*transactionBinaryData = (*transactionBinaryData)[len(*transactionBinaryData):]
		} else if isStruct {
			t := reflect.ValueOf(p.TxPtr).Elem()

			//walletId & citizenId
			for i := 2; i < 4; i++ {
				data := lib.FieldToBytes(t.Field(0).Interface(), i)
				returnSlice = append(returnSlice, data)
			}
			for i := 1; i < t.NumField(); i++ {
				data := lib.FieldToBytes(t.Interface(), i)
				returnSlice = append(returnSlice, data)
			}
		}
		if isStruct {
			*transactionBinaryData = (*transactionBinaryData)[len(*transactionBinaryData):]
		}
		if len(*transactionBinaryData) > 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect transactionBinaryData %x", transactionBinaryData))
		}
	}
	return append(transSlice, returnSlice...), nil
}
