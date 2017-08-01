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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
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
		result       SendTxJSON
		public, hash []byte
	)
	contract, err := c.checkTx(nil)
	if err == nil {
		signature, err := crypto.JSSignToBytes(c.r.FormValue("signature1"))
		if err != nil {
			result.Error = err.Error()
		}
		var isPublic []byte
		wallet := &model.DltWallet{}
		err = wallet.GetWallet(c.SessWalletID)
		isPublic = wallet.PublicKey
		if err == nil && len(signature) > 0 && len(isPublic) == 0 {
			public, _ := hex.DecodeString(c.r.FormValue(`public`))
			if len(public) == 0 {
				err = fmt.Errorf(`empty public key`)
			} else {
				signature = append(signature, public[1:]...)
			}
		}
		if len(signature) == 0 {
			result.Error = `signature is empty`
		} else if err == nil {
			data := make([]byte, 0)
			info := contract.Block.Info.(*script.ContractInfo)
			if info.Tx != nil {
			fields:
				for _, fitem := range *info.Tx {
					val := strings.TrimSpace(c.r.FormValue(fitem.Name))
					if strings.Contains(fitem.Tags, `address`) {
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
				toSerialize := tx.SmartContract{
					Header: tx.Header{Type: int(info.ID), Time: converter.StrToInt64(c.r.FormValue(`time`)),
						UserID: c.SessWalletID, StateID: c.SessStateID, PublicKey: public,
						BinSignatures: converter.EncodeLengthPlusData(signature)},
					Data: data,
				}
				serializedData, err := msgpack.Marshal(toSerialize)
				transactionStatus := &model.TransactionStatus{Hash: hash, Time: int32(time.Now().Unix()), Type: int32(info.ID),
					WalletID: c.SessWalletID, CitizenID: c.SessWalletID}
				err = transactionStatus.Create()
				queueTx := &model.QueueTx{Hash: hash, Data: data}
				err = queueTx.Create()
				if err == nil {
					hash, err = sql.DB.SendTx(int64(info.ID), c.SessWalletID,
						append([]byte{128}, serializedData...))
					result.Hash = string(hash)
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
