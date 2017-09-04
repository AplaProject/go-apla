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
	"github.com/EGaaS/go-egaas-mvp/packages/model"
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
	pageData := systemInfoPage{Data: c.Data}

	ufn := &model.UpdFullNode{}
	ufnList, err := ufn.GetAll()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, node := range ufnList {
		pageData.UpdFullNodes = append(pageData.UpdFullNodes, node.ToMap())
	}

	rollback := &model.Rollback{}
	rollbacks, err := rollback.GetRollbacks(100)

	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, rb := range rollbacks {
		pageData.Rollback = append(pageData.Rollback, rb.ToMap())
	}

	fullNode := &model.FullNode{}
	nodes, err := fullNode.GetAll()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, node := range *nodes {
		pageData.FullNodes = append(pageData.FullNodes, node.ToMap())
	}

	systemParameters, err := model.GetAllSystemParameters()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, param := range systemParameters {
		pageData.SystemParameters = append(pageData.SystemParameters, param.ToMap())
	}

	wallet := &model.DltWallet{}
	pageData.Votes, err = wallet.GetVotes(10)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return proceedTemplate(c, nSystemInfo, &pageData)
}
