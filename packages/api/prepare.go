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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type prepareResult struct {
	ID         string            `json:"request_id"`
	ForSign    string            `json:"forsign"`
	Signs      []TxSignJSON      `json:"signs"`
	Values     map[string]string `json:"values"`
	Time       string            `json:"time"`
	Expiration string            `json:"expiration"`
}

type multiPrepareResult struct {
	ID       string   `json:"request_id"`
	ForSigns []string `json:"forsign"`
	Time     string   `json:"time"`
}

type multiPrepareRequest struct {
	TokenEcosystem string `json:"token_ecosystem"`
	MaxSum         string `json:"max_sum"`
	Payover        string `json:"payover"`
	SignedBy       string `json:"signed_by"`

	Contracts []multiPrepareRequestItem `json:"contracts"`
}

type multiPrepareRequestItem struct {
	Contract string            `json:"contract"`
	Params   map[string]string `json:"params"`
}

type contractHandlers struct {
	requests      *tx.RequestBuffer
	multiRequests *tx.MultiRequestBuffer
}

func (h *contractHandlers) prepareMultipleContract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	requests := multiPrepareRequest{}
	if err := json.Unmarshal([]byte(r.FormValue("data")), &requests); err != nil {
		return errorAPI(w, err, http.StatusBadRequest)
	}

	tokenEcosystem := converter.StrToInt64(requests.TokenEcosystem)
	maxSum := requests.MaxSum
	payOver := requests.Payover
	var signedBy int64
	if requests.SignedBy != "" {
		signedBy = converter.StrToInt64(requests.SignedBy)
	}

	req := h.multiRequests.NewMultiRequest()
	forSigns := []string{}

	for _, c := range requests.Contracts {
		var smartTx tx.SmartContract
		contract, parerr, err := validateSmartContractJSON(r, data, c.Contract, c.Params)
		if err != nil {
			if strings.HasPrefix(err.Error(), `E_`) {
				return errorAPI(w, err.Error(), http.StatusBadRequest, parerr)
			}
			return errorAPI(w, err, http.StatusBadRequest)
		}
		info := (*contract).Block.Info.(*script.ContractInfo)
		smartTx.TokenEcosystem = tokenEcosystem
		smartTx.MaxSum = maxSum
		smartTx.PayOver = payOver
		if signedBy != 0 {
			smartTx.SignedBy = signedBy
		}

		smartTx.RequestID = req.ID
		smartTx.Header = tx.Header{
			Type:        int(info.ID),
			Time:        req.Time.Unix(),
			EcosystemID: data.ecosystemId,
			KeyID:       data.keyId,
			RoleID:      data.roleId,
			NetworkID:   consts.NETWORK_ID,
		}
		forsign := []string{smartTx.ForSign()}
		if info.Tx != nil {
			f, requestParams, err := forsignJSONData(w, c.Params, logger, *info.Tx)
			if err != nil {
				return err
			}
			forsign = append(forsign, f...)
			req.AddContract(c.Contract, requestParams)
		} else {
			req.AddContract(c.Contract, c.Params)
		}
		forSigns = append(forSigns, strings.Join(forsign, ","))
	}
	h.multiRequests.AddRequest(req)

	result := multiPrepareResult{
		ID:       req.ID,
		ForSigns: forSigns,
		Time:     converter.Int64ToStr(req.Time.Unix()),
	}
	data.result = result
	return nil
}

func (h *contractHandlers) prepareContract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var (
		result  prepareResult
		smartTx tx.SmartContract
	)

	contract, parerr, err := validateSmartContract(r, data, &result, data.params["name"].(string))
	if err != nil {
		if strings.HasPrefix(err.Error(), `E_`) {
			return errorAPI(w, err.Error(), http.StatusBadRequest, parerr)
		}
		return errorAPI(w, err, http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	smartTx.TokenEcosystem = data.params[`token_ecosystem`].(int64)
	smartTx.MaxSum = data.params[`max_sum`].(string)
	smartTx.PayOver = data.params[`payover`].(string)
	if data.params[`signed_by`] != nil {
		smartTx.SignedBy = data.params[`signed_by`].(int64)
	}

	req := h.requests.NewRequest(contract.Name)

	smartTx.RequestID = req.ID
	smartTx.Header = tx.Header{
		Type:        int(info.ID),
		Time:        req.Time.Unix(),
		EcosystemID: data.ecosystemId,
		KeyID:       data.keyId,
		RoleID:      data.roleId,
		NetworkID:   consts.NETWORK_ID,
	}

	forsign := []string{smartTx.ForSign()}
	if info.Tx != nil {
		f, err := forsignFormData(w, r, data, logger, req, *info.Tx)
		if err != nil {
			return err
		}
		forsign = append(forsign, f...)
	}

	result.ID = req.ID
	result.ForSign = strings.Join(forsign, ",")
	result.Time = converter.Int64ToStr(req.Time.Unix())
	result.Expiration = converter.Int64ToStr(req.Time.Add(h.requests.ExpireDuration()).Unix())
	data.result = result
	return nil
}

func forsignJSONData(w http.ResponseWriter, params map[string]string, logger *log.Entry, fields []*script.FieldInfo) ([]string, map[string]string, error) {
	forsign := []string{}
	requestParams := map[string]string{}
	for _, fitem := range fields {
		if fitem.ContainsTag(`signature`) || fitem.ContainsTag(script.TagFile) {
			continue
		}
		var val string
		if fitem.Type.String() == `[]interface {}` {
			for key, values := range params {
				if key == fitem.Name+`[]` && len(values) > 0 {
					count := converter.StrToInt(string(values[0]))
					requestParams[key] = string(values[0])
					var list []string
					for i := 0; i < count; i++ {
						k := fmt.Sprintf(`%s[%d]`, fitem.Name, i)
						v := params[k]
						list = append(list, v)
						requestParams[k] = v
					}
					val = strings.Join(list, `,`)
				}
			}
			if len(val) == 0 {
				val = params[fitem.Name]
				requestParams[fitem.Name] = val
			}
		} else {
			val = strings.TrimSpace(params[fitem.Name])
			requestParams[fitem.Name] = val
			if fitem.ContainsTag(`address`) {
				val = converter.Int64ToStr(converter.StringToAddress(val))
			} else if fitem.Type.String() == script.Decimal {
				val = strings.TrimLeft(val, `0`)
			} else if fitem.Type.String() == `int64` && len(val) == 0 {
				val = `0`
			}
		}
		forsign = append(forsign, val)
	}

	return forsign, requestParams, nil
}

func forsignFormData(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry, req *tx.Request, fields []*script.FieldInfo) ([]string, error) {
	forsign := []string{}
	for _, fitem := range fields {
		if strings.Contains(fitem.Tags, `signature`) {
			continue
		}
		var val string
		if fitem.ContainsTag(script.TagFile) {
			file, header, err := r.FormFile(fitem.Name)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("getting multipart file")
				return nil, errorAPI(w, err.Error(), http.StatusBadRequest)
			}
			fileHeader, err := req.WriteFile(fitem.Name, header.Header.Get(`Content-Type`), file)
			file.Close()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing file")
				return nil, errorAPI(w, err.Error(), http.StatusInternalServerError)
			}
			forsign = append(forsign, fileHeader.MimeType, fileHeader.Hash)
			continue
		}

		switch fitem.Type.String() {
		case `[]interface {}`:
			for key, values := range r.Form {
				if key == fitem.Name+`[]` && len(values) > 0 {
					count := converter.StrToInt(values[0])
					req.SetValue(key, values[0])
					var list []string
					for i := 0; i < count; i++ {
						k := fmt.Sprintf(`%s[%d]`, fitem.Name, i)
						v := r.FormValue(k)
						list = append(list, v)
						req.SetValue(k, v)
					}
					val = strings.Join(list, `,`)
				}
			}
			if len(val) == 0 {
				val = r.FormValue(fitem.Name)
				req.SetValue(fitem.Name, val)
			}

		case script.Decimal:
			d, err := decimal.NewFromString(r.FormValue(fitem.Name))
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("converting to decimal")
				return nil, errorAPI(w, err, http.StatusBadRequest)
			}

			sp := &model.StateParameter{}
			sp.SetTablePrefix(getPrefix(data))
			if _, err = sp.Get(nil, model.ParamMoneyDigit); err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting value from db")
				return nil, errorAPI(w, err, http.StatusInternalServerError)
			}
			exp := int32(converter.StrToInt(sp.Value))

			val = d.Mul(decimal.New(1, exp)).StringFixed(0)
			req.SetValue(fitem.Name, val)

		default:
			val = strings.TrimSpace(r.FormValue(fitem.Name))
			req.SetValue(fitem.Name, val)
			if strings.Contains(fitem.Tags, `address`) {
				val = converter.Int64ToStr(converter.StringToAddress(val))
			} else if fitem.Type.String() == script.Decimal {
				val = strings.TrimLeft(val, `0`)
			} else if fitem.Type.String() == `int64` && len(val) == 0 {
				val = `0`
			}
		}

		forsign = append(forsign, val)
	}

	return forsign, nil
}
