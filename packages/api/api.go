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
	"net/http"
	"strings"

	"github.com/gorilla/schema"

	"github.com/AplaProject/go-apla/packages/converter"
)

const (
	multipartBuf      = 100000 // the buffer size for ParseMultipartForm
	multipartFormData = "multipart/form-data"
	contentType       = "Content-Type"
)

// Client represents data of client
type Client struct {
	KeyID         int64
	EcosystemID   int64
	EcosystemName string
	RoleID        int64
	IsMobile      bool
}

func (c *Client) Prefix() string {
	return converter.Int64ToStr(c.EcosystemID)
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
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
