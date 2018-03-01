// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package api

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type contractResult struct {
	Hash string `json:"hash"`
	// These fields are used for VDE
	Message *txstatusError `json:"errmsg,omitempty"`
	Result  string         `json:"result,omitempty"`
}

func contract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var (
		hash, publicKey []byte
		toSerialize     interface{}
	)
	contract, parerr, err := validateSmartContract(r, data, nil)
	if err != nil {
		if strings.HasPrefix(err.Error(), `E_`) {
			return errorAPI(w, err.Error(), http.StatusBadRequest, parerr)
		}
		return errorAPI(w, err, http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)

	var signedBy int64
	signID := data.keyId
	if data.params[`signed_by`] != nil {
		signedBy = data.params[`signed_by`].(int64)
		signID = signedBy
	}

	key := &model.Key{}
	key.SetTablePrefix(data.ecosystemId)
	_, err = key.Get(signID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if len(key.PublicKey) == 0 {
		if _, ok := data.params[`pubkey`]; ok && len(data.params[`pubkey`].([]byte)) > 0 {
			publicKey = data.params[`pubkey`].([]byte)
			lenpub := len(publicKey)
			if lenpub > 64 {
				publicKey = publicKey[lenpub-64:]
			}
		}
		if len(publicKey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			return errorAPI(w, `E_EMPTYPUBLIC`, http.StatusBadRequest)
		}
	} else {
		logger.Warning("public key for wallet not found")
		publicKey = []byte("null")
	}
	signature := data.params[`signature`].([]byte)
	if len(signature) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signature is empty")
		return errorAPI(w, `E_EMPTYSIGN`, http.StatusBadRequest)
	}
	idata := make([]byte, 0)
	if info.Tx != nil {
	fields:
		for _, fitem := range *info.Tx {
			val := strings.TrimSpace(r.FormValue(fitem.Name))
			if strings.Contains(fitem.Tags, `address`) {
				val = converter.Int64ToStr(converter.StringToAddress(val))
			}
			switch fitem.Type.String() {
			case `[]interface {}`:
				var list []string
				for key, values := range r.Form {
					if key == fitem.Name+`[]` && len(values) > 0 {
						count := converter.StrToInt(values[0])
						for i := 0; i < count; i++ {
							list = append(list, r.FormValue(fmt.Sprintf(`%s[%d]`, fitem.Name, i)))
						}
					}
				}
				if len(list) == 0 && len(val) > 0 {
					list = append(list, val)
				}
				idata = append(idata, converter.EncodeLength(int64(len(list)))...)
				for _, ilist := range list {
					blist := []byte(ilist)
					idata = append(append(idata, converter.EncodeLength(int64(len(blist)))...), blist...)
				}
			case `uint64`:
				converter.BinMarshal(&idata, converter.StrToUint64(val))
			case `int64`:
				converter.EncodeLenInt64(&idata, converter.StrToInt64(val))
			case `float64`:
				converter.BinMarshal(&idata, converter.StrToFloat64(val))
			case `string`, script.Decimal:
				idata = append(append(idata, converter.EncodeLength(int64(len(val)))...), []byte(val)...)
			case `[]uint8`:
				var bytes []byte
				bytes, err = hex.DecodeString(val)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": val}).Error("decoding value from hex")
					break fields
				}
				idata = append(append(idata, converter.EncodeLength(int64(len(bytes)))...), bytes...)
			}
		}
	}
	toSerialize = tx.SmartContract{
		Header: tx.Header{Type: int(info.ID), Time: converter.StrToInt64(data.params[`time`].(string)),
			EcosystemID: data.ecosystemId, KeyID: data.keyId, PublicKey: publicKey,
			BinSignatures: converter.EncodeLengthPlusData(signature)},
		TokenEcosystem: data.params[`token_ecosystem`].(int64),
		MaxSum:         data.params[`max_sum`].(string),
		PayOver:        data.params[`payover`].(string),
		SignedBy:       signedBy,
		Data:           idata,
	}
	serializedData, err := msgpack.Marshal(toSerialize)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if data.vde {
		ret, err := VDEContract(serializedData, data)
		if err != nil {
			return errorAPI(w, err, http.StatusInternalServerError)
		}
		data.result = ret
		return nil
	}
	if hash, err = model.SendTx(int64(info.ID), data.keyId,
		append([]byte{128}, serializedData...)); err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	data.result = &contractResult{Hash: hex.EncodeToString(hash)}
	return nil
}
