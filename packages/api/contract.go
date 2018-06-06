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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func getPublicKey(w http.ResponseWriter, r *http.Request, signID int64, ecosystemID int64, pubkey []byte) ([]byte, bool) {
	logger := getLogger(r)

	var publicKey []byte
	key := &model.Key{}
	key.SetTablePrefix(ecosystemID)
	_, err := key.Get(signID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		errorResponse(w, err, http.StatusInternalServerError)
		return nil, false
	}
	if key.Deleted == 1 {
		errorResponse(w, errDeletedKey, http.StatusForbidden)
		return nil, false
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
			errorResponse(w, errEmptyPublic, http.StatusBadRequest)
			return nil, false
		}
	} else {
		logger.Warning("public key for wallet not found")
		publicKey = []byte("null")
	}
	return publicKey, true
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

func getData(w http.ResponseWriter, r *http.Request, fields []*script.FieldInfo, req *tx.Request) ([]byte, bool) {
	logger := getLogger(r)

	idata := []byte{}
	var err error
	for _, fitem := range fields {
		if fitem.ContainsTag(script.TagFile) {
			file, err := req.ReadFile(fitem.Name)
			if err != nil {
				errorResponse(w, err, http.StatusInternalServerError)
				return nil, false
			}

			serialFile, err := msgpack.Marshal(file)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling file to msgpack")
				errorResponse(w, err, http.StatusInternalServerError)
				return nil, false
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
				errorResponse(w, err, http.StatusBadRequest)
				return nil, false
			}
			idata = append(append(idata, converter.EncodeLength(int64(len(bytes)))...), bytes...)
		}
	}
	return idata, true
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

func (c *contractHandlers) ContractMultiHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	requestID := params["request_id"]
	var publicKey []byte
	req, ok := c.multiRequests.GetRequest(requestID)
	if !ok {
		errorResponse(w, errRequestNotFound, http.StatusNotFound, requestID)
		return
	}
	multiRequest := contractMultiRequest{}
	if err := json.Unmarshal([]byte(r.FormValue("data")), &multiRequest); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}
	var signedBy int64
	signID := client.KeyID
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
			errorResponse(w, err, http.StatusBadRequest)
			return
		}
	}
	publicKey, ok = getPublicKey(w, r, signID, client.EcosystemID, pubkey)
	if !ok {
		return
	}
	signatures := multiRequest.Signatures
	if len(signatures) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signatures is empty")
		errorResponse(w, errEmptySign, http.StatusBadRequest)
		return
	}
	tokenEcosystem := converter.StrToInt64(multiRequest.TokenEcosystem)
	maxSum := multiRequest.MaxSum
	payover := multiRequest.Payover
	hashes := []string{}
	for i, c := range req.Contracts {
		contract := getContract(r, c.Contract)
		if contract == nil {
			errorResponse(w, errContract, http.StatusBadRequest, c.Contract)
			return
		}
		info := (*contract).Block.Info.(*script.ContractInfo)

		idata := make([]byte, 0)
		if info.Tx != nil {
			idata, err = getDataMultiRequestParams(*info.Tx, c.Params, w, logger)
			if err != nil {
				errorResponse(w, err, http.StatusBadRequest)
				return
			}
		}
		signatureBytes, err := hex.DecodeString(signatures[i])
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("converting signature from hex")
			errorResponse(w, err, http.StatusInternalServerError)
			return
		}
		toSerialize := tx.SmartContract{
			Header: tx.Header{
				Type:          int(info.ID),
				Time:          converter.StrToInt64(multiRequest.Time),
				EcosystemID:   client.EcosystemID,
				KeyID:         client.KeyID,
				RoleID:        client.RoleID,
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
			errorResponse(w, err, http.StatusInternalServerError)
			return
		}
		if hash, err := model.SendTx(int64(info.ID), client.KeyID,
			append([]byte{128}, serializedData...)); err != nil {
			errorResponse(w, err, http.StatusInternalServerError)
			return
		} else {
			hashes = append(hashes, hex.EncodeToString(hash))
		}
	}

	jsonResponse(w, &contractMultiResult{Hashes: hashes})
}

func (c *contractHandlers) ContractHandler(w http.ResponseWriter, r *http.Request) {
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

	req, ok := c.requests.GetRequest(requestID)
	if !ok {
		errorResponse(w, errRequestNotFound, http.StatusNotFound, requestID)
		return
	}

	contract := getContract(r, req.Contract)
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

	if len(form.PublicKey.Value()) > 0 {
		publicKey = form.PublicKey.Value()
	}
	if publicKey, ok = getPublicKey(w, r, signID, client.EcosystemID, publicKey); !ok {
		return
	}

	signature := form.Signature.Value()
	if len(signature) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("signature is empty")
		errorResponse(w, errEmptySign, http.StatusBadRequest)
		return
	}

	idata := make([]byte, 0)
	if info.Tx != nil {
		if idata, ok = getData(w, r, *info.Tx, req); !ok {
			return
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

	// TODO: vde support
	// if data.vde {
	// 	ret, err := VDEContract(serializedData, data)
	// 	if err != nil {
	// 		return errorAPI(w, err, http.StatusInternalServerError)
	// 	}
	// 	data.result = ret
	// 	return nil
	// }
	if hash, err = model.SendTx(int64(info.ID), client.KeyID,
		append([]byte{128}, serializedData...)); err != nil {
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, &contractResult{Hash: hex.EncodeToString(hash)})
}
