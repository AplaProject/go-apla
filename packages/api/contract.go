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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
)

func getPublicKey(signID int64, ecosystemID int64, pubkey []byte, w http.ResponseWriter, logger *log.Entry) ([]byte, error) {
	var publicKey []byte
	key := &model.Key{}
	key.SetTablePrefix(ecosystemID)
	_, err := key.Get(signID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		return []byte(""), errorAPI(w, err, http.StatusInternalServerError)
	}
	if key.Deleted == 1 {
		return []byte(""), errorAPI(w, `E_DELETEDKEY`, http.StatusForbidden)
	}
	if len(key.PublicKey) == 0 {
		if len(pubkey) > 0 {
			publicKey = pubkey
			lenpub := len(publicKey)
			if lenpub > 64 {
				publicKey = publicKey[lenpub-64:]
			}
		}
		if len(publicKey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			return []byte(""), errorAPI(w, `E_EMPTYPUBLIC`, http.StatusBadRequest)
		}
	} else {
		logger.Warning("public key for wallet not found")
		publicKey = []byte("null")
	}
	return publicKey, nil
}

func getDataMultiRequestParams(fields []*script.FieldInfo, params map[string]string, w http.ResponseWriter, logger *log.Entry) ([]byte, error) {
	idata := []byte{}
	var err error
	for _, fitem := range fields {
		val := strings.TrimSpace(params[fitem.Name])
		if strings.Contains(fitem.Tags, `address`) {
			val = converter.Int64ToStr(converter.StringToAddress(val))
		}
		switch fitem.Type.String() {
		case `[]interface {}`:
			var list []string
			for key, value := range params {
				if key == fitem.Name+`[]` && len(value) > 0 {
					count := converter.StrToInt(value)
					for i := 0; i < count; i++ {
						list = append(list, params[fmt.Sprintf(`%s[%d]`, fitem.Name, i)])
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
				return idata, err
			}
			idata = append(append(idata, converter.EncodeLength(int64(len(bytes)))...), bytes...)
		}
	}
	return idata, nil
}

func getData(fields []*script.FieldInfo, req *tx.Request, w http.ResponseWriter, logger *log.Entry) ([]byte, error) {
	idata := []byte{}
	var err error
	for _, fitem := range fields {
		if fitem.ContainsTag(script.TagFile) {
			file, err := req.ReadFile(fitem.Name)
			if err != nil {
				return idata, errorAPI(w, err.Error(), http.StatusInternalServerError)
			}

			serialFile, err := msgpack.Marshal(file)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling file to msgpack")
				return idata, errorAPI(w, err, http.StatusInternalServerError)
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
			for key, value := range req.AllValues() {
				if key == fitem.Name+`[]` && len(value) > 0 {
					count := converter.StrToInt(value)
					for i := 0; i < count; i++ {
						list = append(list, req.GetValue(fmt.Sprintf(`%s[%d]`, fitem.Name, i)))
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
				return idata, err
			}
			idata = append(append(idata, converter.EncodeLength(int64(len(bytes)))...), bytes...)
		}
	}
	return idata, nil
}

type contractResult struct {
	Hash string `json:"hash"`
	// These fields are used for VDE
	Message *txstatusError `json:"errmsg,omitempty"`
	Result  string         `json:"result,omitempty"`
}

type contractMultiRequest struct {
	Pubkey         string   `json:"pubkey"`
	TokenEcosystem string   `json:"token_ecosystem"`
	MaxSum         string   `json:"max_sum"`
	Payover        string   `json:"payover"`
	SignedBy       string   `json:"signed_by"`
	Signatures     []string `json:"signatures"`
	Time           string   `json:"time"`
}

type contractMultiResult struct {
	Hashes []string `json:"hashes"`
}

func (c *contractHandlers) contractMulti(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	requestID := data.ParamString("request_id")
	var publicKey []byte
	req, ok := c.multiRequests.GetRequest(requestID)
	if !ok {
		return errorAPI(w, "E_REQUESTNOTFOUND", http.StatusNotFound, requestID)
	}
	multiRequest := contractMultiRequest{}
	if err := json.Unmarshal([]byte(r.FormValue("data")), &multiRequest); err != nil {
		return errorAPI(w, err, http.StatusBadRequest)
	}
	var signedBy int64
	signID := data.keyId
	if multiRequest.SignedBy != "" {
		signedBy = converter.StrToInt64(multiRequest.SignedBy)
		signID = signedBy
	}
	var err error
	pubkey := []byte{}
	if multiRequest.Pubkey != "" {
		pubkey, err = hex.DecodeString(multiRequest.Pubkey)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("converting signature from hex")
			return err
		}
	}
	publicKey, err = getPublicKey(signID, data.ecosystemId, pubkey, w, logger)
	if err != nil {
		return err
	}
	signatures := multiRequest.Signatures
	if len(signatures) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signatures is empty")
		return errorAPI(w, `E_EMPTYSIGN`, http.StatusBadRequest)
	}
	tokenEcosystem := converter.StrToInt64(multiRequest.TokenEcosystem)
	maxSum := multiRequest.MaxSum
	payover := multiRequest.Payover
	hashes := []string{}
	for i, c := range req.Contracts {
		contract := smart.VMGetContract(data.vm, c.Contract, uint32(data.ecosystemId))
		if contract == nil {
			return errorAPI(w, "E_CONTRACT", http.StatusBadRequest, c.Contract)
		}
		info := (*contract).Block.Info.(*script.ContractInfo)

		idata := make([]byte, 0)
		if info.Tx != nil {
			idata, err = getDataMultiRequestParams(*info.Tx, c.Params, w, logger)
			if err != nil {
				return err
			}
		}
		signatureBytes, err := hex.DecodeString(signatures[i])
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("converting signature from hex")
			return err
		}
		toSerialize := tx.SmartContract{
			Header: tx.Header{
				Type:          int(info.ID),
				Time:          converter.StrToInt64(multiRequest.Time),
				EcosystemID:   data.ecosystemId,
				KeyID:         data.keyId,
				RoleID:        data.roleId,
				PublicKey:     publicKey,
				NetworkID:     consts.NETWORK_ID,
				BinSignatures: converter.EncodeLengthPlusData(signatureBytes),
			},
			RequestID:      req.ID,
			TokenEcosystem: tokenEcosystem,
			MaxSum:         maxSum,
			PayOver:        payover,
			SignedBy:       signedBy,
			Data:           idata,
		}
		serializedData, err := msgpack.Marshal(toSerialize)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
			return errorAPI(w, err, http.StatusInternalServerError)
		}
		if hash, err := model.SendTx(int64(info.ID), data.keyId,
			append([]byte{128}, serializedData...)); err != nil {
			return errorAPI(w, err, http.StatusInternalServerError)
		} else {
			hashes = append(hashes, hex.EncodeToString(hash))
		}
	}
	data.result = &contractMultiResult{Hashes: hashes}
	return nil
}

func (c *contractHandlers) contract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var (
		hash, publicKey []byte
		toSerialize     interface{}
		requestID       = data.ParamString("request_id")
	)

	req, ok := c.requests.GetRequest(requestID)
	if !ok {
		return errorAPI(w, "E_REQUESTNOTFOUND", http.StatusNotFound, requestID)
	}
	contract := smart.VMGetContract(data.vm, req.Contract, uint32(data.ecosystemId))
	if contract == nil {
		return errorAPI(w, "E_CONTRACT", http.StatusBadRequest, req.Contract)
	}

	info := (*contract).Block.Info.(*script.ContractInfo)

	var signedBy int64
	signID := data.keyId
	if data.params[`signed_by`] != nil {
		signedBy = data.params[`signed_by`].(int64)
		signID = signedBy
	}

	pubkey := []byte{}
	if _, ok := data.params["public_key"]; ok {
		pubkey = data.params["public_key"].([]byte)
	}
	var err error
	publicKey, err = getPublicKey(signID, data.ecosystemId, pubkey, w, logger)
	if err != nil {
		return err
	}

	signature := data.params[`signature`].([]byte)
	if len(signature) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signature is empty")
		return errorAPI(w, `E_EMPTYSIGN`, http.StatusBadRequest)
	}
	idata := make([]byte, 0)
	if info.Tx != nil {
		idata, err = getData(*info.Tx, req, w, logger)
		if err != nil {
			return err
		}
	}
	toSerialize = tx.SmartContract{
		Header: tx.Header{
			Type:          int(info.ID),
			Time:          converter.StrToInt64(data.params[`time`].(string)),
			EcosystemID:   data.ecosystemId,
			KeyID:         data.keyId,
			RoleID:        data.roleId,
			PublicKey:     publicKey,
			NetworkID:     consts.NETWORK_ID,
			BinSignatures: converter.EncodeLengthPlusData(signature),
		},
		RequestID:      req.ID,
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

func blockchainUpdatingState(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var reason string

	switch service.NodePauseType() {
	case service.NoPause:
		return nil
	case service.PauseTypeUpdatingBlockchain:
		reason = "E_UPDATING"
		break
	case service.PauseTypeStopingNetwork:
		reason = "E_STOPPING"
		break
	}

	return errorAPI(w, reason, http.StatusServiceUnavailable)
}
