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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
)

const aSendTx = `ajax_send_tx`

// SendTxJSON is a structure for the answer of ajax_send_tx ajax request
type SendTxJSON struct {
	Error string `json:"error"`
	Hash  string `json:"hash"`
}

func init() {
	newPage(aSendTx, `json`)
}

// AjaxSendTx is a controller of ajax_send_tx request
func (c *Controller) AjaxSendTx() interface{} {
	var (
		result SendTxJSON
		flags  uint8
	)
	contract, err := c.checkTx(nil)
	if err == nil {
		//		info := (*contract).Block.Info.(*script.ContractInfo)
		userID := uint64(c.SessWalletID)
		sign := make([]byte, 0)
		signature, err := crypto.JSSignToBytes(c.r.FormValue("signature1"))
		if err != nil {
			result.Error = err.Error()
		} else if len(signature) > 0 {
			converter.EncodeLenByte(&sign, signature)
		}
		var isPublic []byte
		isPublic, err = c.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, c.SessWalletID).Bytes()
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
				Type:     int32(contract.Block.Info.(*script.ContractInfo).ID), /* + smart.CNTOFF*/
				Time:     uint32(converter.StrToInt64(c.r.FormValue(`time`))),
				WalletID: userID,
				StateID:  int32(c.SessStateID),
				Flags:    flags,
				Sign:     sign,
			}
			//fmt.Println(`SEND TX`, contract.Block.Info.(*script.ContractInfo))
			//			fmt.Println(`Header`, header)
			_, err = converter.BinMarshal(&data, &header)
			if err == nil {
				if contract.Block.Info.(*script.ContractInfo).Tx != nil {
				fields:
					for _, fitem := range *contract.Block.Info.(*script.ContractInfo).Tx {
						val := strings.TrimSpace(c.r.FormValue(fitem.Name))
						if strings.Index(fitem.Tags, `address`) >= 0 {
							val = converter.Int64ToStr(converter.StringToAddress(val))
						}
						switch fitem.Type.String() {
						case `[]interface {}`:
							var list []string
							for key, values := range c.r.Form {
								if key == fitem.Name+`[]` {
									for _, value := range values {
										list = append(list, value)
									}
								}
							}
							data = append(data, converter.EncodeLength(int64(len(list)))...)
							for _, ilist := range list {
								blist := []byte(ilist)
								data = append(append(data, converter.EncodeLength(int64(len(blist)))...), blist...)
							}
						case `uint64`:
							converter.BinMarshal(&data, converter.StrToUint64(val))
							//					case `float64`:
							//						lib.BinMarshal(&data, utils.StrToFloat64(val))
						case `int64`:
							converter.EncodeLenInt64(&data, converter.StrToInt64(val))
						case `float64`:
							converter.BinMarshal(&data, converter.StrToFloat64(val))
						case `string`, script.Decimal:
							data = append(append(data, converter.EncodeLength(int64(len(val)))...), []byte(val)...)
						case `[]uint8`:
							var bytes []byte
							bytes, err = hex.DecodeString(val)
							if err != nil {
								break fields
							}
							data = append(append(data, converter.EncodeLength(int64(len(bytes)))...), bytes...)
						}
					}
				}
				if err == nil {
					hash, err := crypto.Hash(data)
					if err != nil {
						log.Fatal(err)
					}
					hash = converter.BinToHex(hash)
					err = c.ExecSQL(`INSERT INTO transactions_status (
						hash, time,	type, wallet_id, citizen_id	) VALUES (
						[hex], ?, ?, ?, ? )`, hash, time.Now().Unix(), header.Type, int64(userID), int64(userID)) //c.SessStateID)
					if err == nil {
						log.Debug("INSERT INTO queue_tx (hash, data) VALUES (%s, %s)", hash, hex.EncodeToString(data))
						err = c.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", hash, hex.EncodeToString(data))
						if err == nil {
							result.Hash = string(hash)
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
