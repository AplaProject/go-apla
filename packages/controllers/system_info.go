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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const nSystemInfo = `system_info`

type systemInfoPage struct {
	Data             *CommonPage
	List             []map[string]string
	Latest           int64
	BlockID          int64
	UpdFullNodes     []map[string]string
	MainLock         []map[string]string
	Rollback         []map[string]string
	FullNodes        []map[string]string
	Votes            []map[string]string
	SystemParameters []map[string]string
}

func init() {
	newPage(nSystemInfo)
}

// SystemInfo shows the system information about the blockchain
func (c *Controller) SystemInfo() (string, error) {
	var err error
	pageData := systemInfoPage{Data: c.Data}

	pageData.UpdFullNodes, err = c.GetAll(`SELECT * FROM upd_full_nodes`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	pageData.MainLock, err = c.GetAll(`SELECT * FROM main_lock`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	pageData.Rollback, err = c.GetAll(`SELECT * FROM rollback ORDER BY rb_id DESC LIMIT 100`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	pageData.FullNodes, err = c.GetAll(`SELECT * FROM full_nodes`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	pageData.SystemParameters, err = c.GetAll(`SELECT * FROM system_parameters`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	pageData.Votes, err = c.GetAll(`SELECT address_vote, sum(amount) as sum FROM dlt_wallets WHERE address_vote !='' GROUP BY address_vote ORDER BY sum(amount) DESC LIMIT 10`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return proceedTemplate(c, nSystemInfo, &pageData)
}
