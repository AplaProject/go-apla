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
	"math/rand"
	"net/http"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
)

type getUIDResult struct {
	UID string `json:"uid"`
}

func getUID(w http.ResponseWriter, r *http.Request, data *apiData) error {
	uid := converter.Int64ToStr(rand.New(rand.NewSource(time.Now().Unix())).Int63())
	sess, err := apiSess.SessionStart(w, r)
	if err != nil {
		return err
	}
	sess.Set("uid", uid)
	sess.SessionRelease(w)
	data.result = &getUIDResult{UID: uid}
	return nil
}
