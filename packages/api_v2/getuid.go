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

package api_v2

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

var (
	installed bool
)

type getUIDResult struct {
	UID     string `json:"uid"`
	State   int64  `json:"state"`
	Wallet  int64  `json:"wallet"`
	Address string `json:"address"`
}

// If State == 0 then APLA has not been installed
// If Wallet == 0 then login is required

func getUID(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var result getUIDResult

	data.result = &result
	if !installed {
		if model.DBConn == nil && !config.IsExist() {
			return nil
		}
		installed = true
	}
	result.UID = converter.Int64ToStr(rand.New(rand.NewSource(time.Now().Unix())).Int63())
	sess, err := apiSess.SessionStart(w, r)
	if err != nil {
		return err
	}

	sess.Set("uid", result.UID)
	sess.SessionRelease(w)
	return nil
}
