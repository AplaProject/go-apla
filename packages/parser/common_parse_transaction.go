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
	"reflect"

	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/script"
	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/utils"

	"github.com/shopspring/decimal"
)

func (p *Parser) ParseTransaction(transactionBinaryData *[]byte) ([][]byte, error) {

	var returnSlice [][]byte
	var transSlice [][]byte
	var merkleSlice [][]byte
	log.Debug("transactionBinaryData: %x", *transactionBinaryData)
	log.Debug("transactionBinaryData: %s", *transactionBinaryData)
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
			p.TxStateID = uint32(p.TxPtr.(*consts.TXHeader).StateId)
			p.TxStateIDStr = utils.UInt32ToStr(p.TxStateID)
			if p.TxStateID > 0 {
				p.TxCitizenID = int64(p.TxPtr.(*consts.TXHeader).WalletId)
				p.TxWalletID = 0
			} else {
				p.TxCitizenID = 0
				p.TxWalletID = int64(p.TxPtr.(*consts.TXHeader).WalletId)
			}
			contract := smart.GetContractById(p.TxPtr.(*consts.TXHeader).Type)
			if contract == nil {
				return nil, fmt.Errorf(`unknown contract %d`, p.TxPtr.(*consts.TXHeader).Type)
			}
			p.TxContract = contract
			p.TxData = make(map[string]interface{})
			for _, fitem := range *contract.Block.Info.(*script.ContractInfo).Tx {
				var v interface{}
				switch fitem.Type.String() {
				case `uint64`:
					var val uint64
					lib.BinUnmarshal(&input, &val)
					v = val
					/*				case `float64`:
									var val float64
									lib.BinUnmarshal(&input, &val)
									v = val*/
				case `int64`:
					v, err = lib.DecodeLenInt64(&input)
				case `decimal.Decimal`:
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
					v = b
				}
				p.TxData[fitem.Name] = v
				if err != nil {
					return nil, err
				}
			}
			//			fmt.Println(`Contract data`, p.TxData)
		} else if isStruct {
			p.TxPtr = consts.MakeStruct(consts.TxTypes[int(txType)])
			if err := lib.BinUnmarshal(&input, p.TxPtr); err != nil {
				return nil, err
			}
			p.TxVars = make(map[string]string)
			if int(txType) == 4 { // TXNewCitizen
				head := consts.HeaderNew(p.TxPtr)
				p.TxStateID = uint32(head.StateId)
				p.TxStateIDStr = utils.UInt32ToStr(p.TxStateID)
				if head.StateId > 0 {
					p.TxCitizenID = int64(head.WalletId)
					p.TxWalletID = 0
				} else {
					p.TxCitizenID = 0
					p.TxWalletID = int64(head.WalletId)
				}
				p.TxTime = int64(head.Time)
			} else {
				head := consts.Header(p.TxPtr)
				p.TxCitizenID = head.CitizenId
				p.TxWalletID = head.WalletId
				p.TxTime = int64(head.Time)
			}
			fmt.Println(`PARSED STRUCT %v`, p.TxPtr)
		}
		transSlice = append(transSlice, utils.Int64ToByte(txType))
		if len(*transactionBinaryData) == 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect tx"))
		}
		// следующие 4 байта - время транзакции
		transSlice = append(transSlice, utils.Int64ToByte(utils.BinToDecBytesShift(transactionBinaryData, 4)))
		if len(*transactionBinaryData) == 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect tx"))
		}
		log.Debug("%s", transSlice)
		// преобразуем бинарные данные транзакции в массив
		if txType > 127 {
			*transactionBinaryData = (*transactionBinaryData)[len(*transactionBinaryData):]
		} else if isStruct {
			t := reflect.ValueOf(p.TxPtr).Elem()

			//walletId & citizenId
			for i := 2; i < 4; i++ {
				data := lib.FieldToBytes(t.Field(0).Interface(), i)
				returnSlice = append(returnSlice, data)
				merkleSlice = append(merkleSlice, utils.DSha256(data))
			}
			for i := 1; i < t.NumField(); i++ {
				data := lib.FieldToBytes(t.Interface(), i)
				returnSlice = append(returnSlice, data)
				merkleSlice = append(merkleSlice, utils.DSha256(data))
			}
		} else {
			i := 0
			for {
				length := utils.DecodeLength(transactionBinaryData)
				i++
				if length > 0 && length < consts.MAX_TX_SIZE {
					data := utils.BytesShift(transactionBinaryData, length)
					returnSlice = append(returnSlice, data)
					merkleSlice = append(merkleSlice, utils.DSha256(data))
					log.Debug("%x", data)
					log.Debug("%s", data)
				} else if length == 0 && len(*transactionBinaryData) > 0 {
					returnSlice = append(returnSlice, []byte{})
					merkleSlice = append(merkleSlice, utils.DSha256([]byte{}))
					continue
				}
				if length == 0 || i >= 20 { // у нас нет тр-ий с более чем 20 элементами
					break
				}
			}
		}
		if isStruct {
			*transactionBinaryData = (*transactionBinaryData)[len(*transactionBinaryData):]
		}
		if len(*transactionBinaryData) > 0 {
			return transSlice, utils.ErrInfo(fmt.Errorf("incorrect transactionBinaryData %x", transactionBinaryData))
		}
	} else {
		merkleSlice = append(merkleSlice, []byte("0"))
	}
	log.Debug("merkleSlice", merkleSlice)
	if len(merkleSlice) == 0 {
		merkleSlice = append(merkleSlice, []byte("0"))
	}
	p.MerkleRoot = utils.MerkleTreeRoot(merkleSlice)
	log.Debug("MerkleRoot %s\n", p.MerkleRoot)
	return append(transSlice, returnSlice...), nil
}
