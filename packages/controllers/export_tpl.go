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
	"fmt"
	//	"strconv"

	//	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const NExportTpl = `export_tpl`

type exportInfo struct {
	//	Id   int    `json:"id"`
	Name string `json:"name"`
}

type exportTplPage struct {
	Data      *CommonPage
	Contracts *[]exportInfo
	Pages     *[]exportInfo
	Tables    *[]exportInfo
}

func init() {
	newPage(NExportTpl)
}

func (c *Controller) getList(table string) (*[]exportInfo, error) {
	ret := make([]exportInfo, 0)
	contracts, err := c.GetAll(fmt.Sprintf(`select name from "%d_%s" order by name`, c.SessStateId, table), -1)
	if err != nil {
		return nil, err
	}
	for _, ival := range contracts {
		//		id, _ := strconv.ParseInt(ival[`id`], 10, 32)
		ret = append(ret, exportInfo{ival["name"]})
	}
	return &ret, nil
}

func (c *Controller) ExportTpl() (string, error) {
	contracts, err := c.getList(`smart_contracts`)
	if err != nil {
		return ``, err
	}
	pages, err := c.getList(`pages`)
	if err != nil {
		return ``, err
	}
	tables, err := c.getList(`tables`)
	if err != nil {
		return ``, err
	}
	fmt.Println(`Export`, contracts, pages, tables)
	pageData := exportTplPage{Data: c.Data, Contracts: contracts, Pages: pages, Tables: tables}
	return proceedTemplate(c, NExportTpl, &pageData)
}
