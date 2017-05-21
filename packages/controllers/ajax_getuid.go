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

package controllers

import (
	"math/rand"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const aGetUID = `ajax_get_uid`

// GetUIDJSON is a structure for the answer of ajax_get_uid ajax request
type GetUIDJSON struct {
	UID   string `json:"uid"`
	Error string `json:"error"`
}

func init() {
	newPage(aGetUID, `json`)
}

// AjaxGetUid is a controller of ajax_get_uid request
func (c *Controller) AjaxGetUid() interface{} {
	var result GetUIDJSON

	r := rand.New(rand.NewSource(time.Now().Unix()))
	result.UID = utils.Int64ToStr(r.Int63())
	c.sess.Set("uid", result.UID)
	return result
}
