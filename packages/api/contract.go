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
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type contractResult struct {
	Hash string `json:"hash"`
	// These fields are used for VDE
	Message *txstatusError `json:"errmsg,omitempty"`
	Result  string         `json:"result,omitempty"`
}

type contractForm struct {
	Form
	PublicKey      hexValue `schema:"pubkey"`
	Signature      hexValue `schema:"signature"`
	Time           int64    `schema:"time"`
	TokenEcosystem int64    `schema:"token_ecosystem"`
	MaxSum         string   `schema:"max_sum"`
	PayOver        string   `schema:"payover"`
	SignedBy       int64    `schema:"signed_by"`
}

func (h *callContractHandlers) ContractHandler(w http.ResponseWriter, r *http.Request) {
	form := &contractForm{}
	if ok := ParseForm(w, r, form); !ok {
		return
	}

	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	var (
		hash, publicKey []byte
		toSerialize     interface{}
		requestID       = params["request_id"]
	)

	req, ok := h.requests.GetRequest(requestID)
	if !ok {
		errorResponse(w, errRequestNotFound, http.StatusNotFound, requestID)
		return
	}

	// TODO: вынести в отдельную функцию
	vm := smart.GetVM(false, client.EcosystemID)
	contract := smart.VMGetContract(vm, req.Contract, uint32(client.EcosystemID))
	if contract == nil {
		errorResponse(w, errContract, http.StatusBadRequest, req.Contract)
		return
	}

	info := (*contract).Block.Info.(*script.ContractInfo)

	var signedBy int64
	signID := client.KeyID
	if form.SignedBy != 0 {
		signedBy = form.SignedBy
		signID = form.SignedBy
	}

	key := &model.Key{}
	key.SetTablePrefix(client.EcosystemID)
	_, err := key.Get(signID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if key.Deleted == 1 {
		errorResponse(w, errDeletedKey, http.StatusForbidden)
		return
	}
	if len(key.PublicKey) == 0 {
		if len(form.PublicKey.Value()) > 0 {
			publicKey = form.PublicKey.Value()
			lenpub := len(publicKey)
			if lenpub > 64 {
				publicKey = publicKey[lenpub-64:]
			}
		}
		if len(publicKey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			errorResponse(w, errEmptyPublic, http.StatusBadRequest)
			return
		}
	} else {
		logger.Warning("public key for wallet not found")
		publicKey = []byte("null")
	}
	signature := form.Signature.Value()
	if len(signature) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signature is empty")
		errorResponse(w, errEmptySign, http.StatusBadRequest)
		return
	}
	idata := make([]byte, 0)
	if info.Tx != nil {
	fields:
		for _, fitem := range *info.Tx {
			if fitem.ContainsTag(script.TagFile) {
				file, err := req.ReadFile(fitem.Name)
				if err != nil {
					errorResponse(w, err, http.StatusInternalServerError)
					return
				}

				serialFile, err := msgpack.Marshal(file)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling file to msgpack")
					errorResponse(w, err, http.StatusInternalServerError)
					return
				}

				idata = append(append(idata, converter.EncodeLength(int64(len(serialFile)))...), serialFile...)
				continue
			}

			val := strings.TrimSpace(req.GetValue(fitem.Name))
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
		Header: tx.Header{
			Type:          int(info.ID),
			Time:          form.Time,
			EcosystemID:   client.EcosystemID,
			KeyID:         client.KeyID,
			RoleID:        client.RoleID,
			PublicKey:     publicKey,
			NetworkID:     consts.NETWORK_ID,
			BinSignatures: converter.EncodeLengthPlusData(signature),
		},
		RequestID:      req.ID,
		TokenEcosystem: form.TokenEcosystem,
		MaxSum:         form.MaxSum,
		PayOver:        form.PayOver,
		SignedBy:       signedBy,
		Data:           idata,
	}
	serializedData, err := msgpack.Marshal(toSerialize)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if client.IsVDE {
		// TODO: сделать поддержку vde
		// ret, err := VDEContract(serializedData, data)
		// if err != nil {
		// 	return errorAPI(w, err, http.StatusInternalServerError)
		// }
		// data.result = ret
		return
	}
	if hash, err = model.SendTx(int64(info.ID), client.KeyID,
		append([]byte{128}, serializedData...)); err != nil {
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, &contractResult{Hash: hex.EncodeToString(hash)})
}

func blockchainUpdatingState(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var reason string

	switch service.NodePauseType() {
	case service.NoPause:
		return nil
	case service.PauseTypeUpdatingBlockchain:
		reason = "Node is updating blockchain"
		break
	case service.PauseTypeStopingNetwork:
		reason = "Network is stopping"
		break
	default:
		reason = "Node is paused"
	}

	return errorAPI(w, errors.New(reason), http.StatusServiceUnavailable)
}
