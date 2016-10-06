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
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) TxStatus() (string, error) {

	hash := c.r.FormValue("hash")

	tx, err := c.OneRow(`SELECT block_id, error FROM transactions_status WHERE hash = [hex]`, hash).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if tx["block_id"] != "0" && tx["block_id"] != "" {
		return `{"success":"` + tx["block_id"] + `"}`, nil
	} else if len(tx["error"]) > 0 {
		return "", utils.ErrInfo(tx["error"])
	}
	return `{"wait":"1"}`, nil
}
