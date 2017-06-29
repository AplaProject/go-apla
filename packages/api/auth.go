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
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
)

type authResult struct {
	Address string `json:"address"`
}

func auth(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var msg string
	switch uid := data.sess.Get(`uid`).(type) {
	case string:
		msg = uid
	default:
		return errConflict(w, "unknown uid")
	}
	pubkey := data.params[`pubkey`].([]byte)
	verify, err := crypto.CheckSign(pubkey, msg, data.params[`signature`].([]byte))
	if err != nil {
		return errConflict(w, err.Error())
	}
	if !verify {
		return errConflict(w, `signature is incorrect`)
	}

	data.result = &authResult{Address: crypto.KeyToAddress(pubkey)}
	return nil
}
