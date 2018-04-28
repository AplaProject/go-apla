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
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	log "github.com/sirupsen/logrus"
)

const multipartFormMaxMemory = 32 << 20 // 32 MB

type prepareResult struct {
	ID      string            `json:"request_id"`
	ForSign string            `json:"forsign"`
	Signs   []TxSignJSON      `json:"signs"`
	Values  map[string]string `json:"values"`
	Time    string            `json:"time"`
}

type prepareForm struct {
	Form
	TokenEcosystem int64  `schema:"token_ecosystem"`
	MaxSum         string `schema:"max_sum"`
	PayOver        string `schema:"payover"`
	SignedBy       int64  `schema:"signed_by"`
}

type callContractHandlers struct {
	requests *tx.RequestBuffer
}

func (h *callContractHandlers) PrepareHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(multipartFormMaxMemory)

	form := &prepareForm{}
	if ok := ParseForm(w, r, form); !ok {
		return
	}

	params := mux.Vars(r)
	client := getClient(r)

	var (
		result  prepareResult
		smartTx tx.SmartContract
	)

	contract, parerr, err := validateSmartContract(r, params["name"], &result)
	if err != nil {
		errorResponse(w, err, http.StatusBadRequest, parerr)
		return
	}

	info := (*contract).Block.Info.(*script.ContractInfo)
	smartTx.TokenEcosystem = form.TokenEcosystem
	smartTx.MaxSum = form.MaxSum
	smartTx.PayOver = form.PayOver
	if form.SignedBy != 0 {
		smartTx.SignedBy = form.SignedBy
	}

	req := h.requests.NewRequest(contract.Name)

	smartTx.RequestID = req.ID
	smartTx.Header = tx.Header{
		Type:        int(info.ID),
		Time:        req.Time.Unix(),
		EcosystemID: client.EcosystemID,
		KeyID:       client.KeyID,
		RoleID:      client.RoleID,
		NetworkID:   consts.NETWORK_ID,
	}

	forsign := []string{smartTx.ForSign()}
	if info.Tx != nil {
		f, ok := forsignFormData(w, r, req, *info.Tx)
		if !ok {
			return
		}
		forsign = append(forsign, f...)
	}

	result.ID = req.ID
	result.ForSign = strings.Join(forsign, ",")
	result.Time = converter.Int64ToStr(req.Time.Unix())
	result.Values = make(map[string]string)

	jsonResponse(w, result)
}

func forsignFormData(w http.ResponseWriter, r *http.Request, req *tx.Request, fields []*script.FieldInfo) ([]string, bool) {
	logger := getLogger(r)

	forsign := []string{}
	for _, fitem := range fields {
		if fitem.ContainsTag(script.TagSignature) {
			continue
		}
		var val string
		if fitem.ContainsTag(script.TagFile) {
			file, header, err := r.FormFile(fitem.Name)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("getting multipart file")
				errorResponse(w, err, http.StatusBadRequest)
				return nil, false
			}
			fileHeader, err := req.WriteFile(fitem.Name, header.Header.Get(`Content-Type`), file)
			file.Close()
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing file")
				errorResponse(w, err, http.StatusInternalServerError)
				return nil, false
			}
			forsign = append(forsign, fileHeader.MimeType, fileHeader.Hash)
			continue
		} else if fitem.Type.String() == `[]interface {}` {
			for key, values := range r.Form {
				if key == fitem.Name+`[]` && len(values) > 0 {
					count := converter.StrToInt(values[0])
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
		} else {
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

	return forsign, true
}
