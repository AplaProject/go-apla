// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/schema"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/types"

	log "github.com/sirupsen/logrus"
)

const (
	multipartBuf      = 100000 // the buffer size for ParseMultipartForm
	multipartFormData = "multipart/form-data"
	contentType       = "Content-Type"
)

type Mode struct {
	EcosysIDValidator  types.EcosystemIDValidator
	EcosysNameGetter   types.EcosystemNameGetter
	EcosysLookupGetter types.EcosystemLookupGetter
	ContractRunner     types.SmartContractRunner
	ClientTxProcessor  types.ClientTxPreprocessor
}

// Client represents data of client
type Client struct {
	KeyID         int64
	AccountID     string
	EcosystemID   int64
	EcosystemName string
	RoleID        int64
	IsMobile      bool
}

func (c *Client) Prefix() string {
	return converter.Int64ToStr(c.EcosystemID)
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	jsonResult, err := json.Marshal(v)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marhsalling http response to json")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonResult)
}

func errorResponse(w http.ResponseWriter, err error, code ...int) {
	et, ok := err.(errType)
	if !ok {
		et = errServer
		et.Message = err.Error()
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	if len(code) == 0 {
		w.WriteHeader(et.Status)
	} else {
		w.WriteHeader(code[0])
	}

	jsonResponse(w, et)
}

type formValidator interface {
	Validate(r *http.Request) error
}

type nopeValidator struct{}

func (np nopeValidator) Validate(r *http.Request) error {
	return nil
}

func parseForm(r *http.Request, f formValidator) (err error) {
	if isMultipartForm(r) {
		err = r.ParseMultipartForm(multipartBuf)
	} else {
		err = r.ParseForm()
	}
	if err != nil {
		return
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(f, r.Form); err != nil {
		return err
	}
	return f.Validate(r)
}

func isMultipartForm(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get(contentType), multipartFormData)
}

type hexValue struct {
	value []byte
}

func (hv hexValue) Bytes() []byte {
	return hv.value
}

func (hv *hexValue) UnmarshalText(v []byte) (err error) {
	hv.value, err = hex.DecodeString(string(v))
	return
}
