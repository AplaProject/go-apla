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

package controllers

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const ASendTx = `ajax_send_tx`

type SendTxJson struct {
	Error string `json:"error"`
	Hash  string `json:"hash"`
}

func init() {
	newPage(ASendTx, `json`)
}

func (c *Controller) AjaxSendTx() interface{} {
	var (
		result SendTxJson
		flags  uint8
	)
	contract, err := c.checkTx()
	if err == nil {
		//		info := (*contract).Block.Info.(*script.ContractInfo)
		userId := uint64(c.SessWalletId)
		sign := make([]byte, 0)
		for i := 1; i <= 3; i++ {
			signature := utils.ConvertJSSign(c.r.FormValue(fmt.Sprintf("signature%d", i)))
			if i == 1 || len(signature) > 0 {
				bsign, _ := hex.DecodeString(signature)
				//				sign = append(sign, utils.EncodeLengthPlusData(bsign)...)
				lib.EncodeLenByte(&sign, bsign)
			}
		}
		var isPublic []byte
		isPublic, err = c.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, c.SessWalletId).Bytes()
		if err == nil && len(sign) > 0 && len(isPublic) == 0 {
			flags |= consts.TxfPublic
			public, _ := hex.DecodeString(c.r.FormValue(`public`))
			if len(public) == 0 {
				err = fmt.Errorf(`empty public key`)
			} else {
				sign = append(sign, public[1:]...)
			}
		}
		if len(sign) == 0 {
			result.Error = `signature is empty`
		} else if err == nil {
			//			var (
			data := make([]byte, 0)
			//			)
			header := consts.TXHeader{
				Type:     int32(contract.Block.Info.(*script.ContractInfo).Id), /* + smart.CNTOFF*/
				Time:     uint32(utils.StrToInt64(c.r.FormValue(`time`))),
				WalletId: userId,
				StateId:  int32(c.SessStateId),
				Flags:    flags,
				Sign:     sign,
			}
			//fmt.Println(`SEND TX`, contract.Block.Info.(*script.ContractInfo))
			//			fmt.Println(`Header`, header)
			_, err = lib.BinMarshal(&data, &header)
			if err == nil {
			fields:
				for _, fitem := range *contract.Block.Info.(*script.ContractInfo).Tx {
					val := strings.TrimSpace(c.r.FormValue(fitem.Name))
					if strings.Index(fitem.Tags, `address`) >= 0 {
						val = utils.Int64ToStr(lib.StringToAddress(val))
					}
					switch fitem.Type.String() {
					case `uint64`:
						lib.BinMarshal(&data, utils.StrToUint64(val))
						//					case `float64`:
						//						lib.BinMarshal(&data, utils.StrToFloat64(val))
					case `int64`:
						lib.EncodeLenInt64(&data, utils.StrToInt64(val))
					case `float64`:
						lib.BinMarshal(&data, utils.StrToFloat64(val))
					case `string`, `decimal.Decimal`:
						data = append(append(data, lib.EncodeLength(int64(len(val)))...), []byte(val)...)
					case `[]uint8`:
						var bytes []byte
						bytes, err = hex.DecodeString(val)
						if err != nil {
							break fields
						}
						data = append(append(data, lib.EncodeLength(int64(len(bytes)))...), bytes...)
					}
				}
				if err == nil {
					md5 := utils.Md5(data)
					err = c.ExecSql(`INSERT INTO transactions_status (
						hash, time,	type, wallet_id, citizen_id	) VALUES (
						[hex], ?, ?, ?, ? )`, md5, time.Now().Unix(), header.Type, int64(userId), int64(userId)) //c.SessStateId)
					if err == nil {
						log.Debug("INSERT INTO queue_tx (hash, data) VALUES (%s, %s)", md5, hex.EncodeToString(data))
						err = c.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, hex.EncodeToString(data))
						if err == nil {
							result.Hash = string(md5)
						}
					}
				}
			}
			fmt.Printf("Data error: %v lendata: %d hash: %s", err, len(data), result.Hash)
		}
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
