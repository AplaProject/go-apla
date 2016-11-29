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
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"regexp"
	"sort"
	"strings"
)

const NAppCatalog = `app_catalog`

type AppInfo struct {
	Name  string
	Title string
	Desc  string
}

type AppsList []AppInfo

type appCatalogPage struct {
	List *AppsList
	Data *CommonPage
}

func init() {
	newPage(NAppCatalog)
}

func getPar(data string, name string) string {
	re := regexp.MustCompile(fmt.Sprintf("%s:\\s*\"(.*)\"", name))
	ret := re.FindStringSubmatch(data)
	if len(ret) > 1 {
		return ret[1]
	}
	return ``
}

func (a AppsList) Len() int           { return len(a) }
func (a AppsList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AppsList) Less(i, j int) bool { return strings.Compare(a[i].Title, a[j].Title) < 1 }

func (c *Controller) AppCatalog() (string, error) {

	files, err := static.AssetDir(`static`)
	if err != nil {
		return ``, err
	}
	list := make(AppsList, 0)
	for _, item := range files {
		if strings.HasSuffix(item, `.tpl`) {
			data, err := static.Asset(`static/` + item)
			if err != nil {
				return ``, err
			}
			var app AppInfo
			app.Name = item[:len(item)-4]
			app.Title = getPar(string(data), `Head`)
			app.Desc = getPar(string(data), `Desc`)
			list = append(list, app)
		}
	}
	sort.Sort(AppsList(list))
	pageData := appCatalogPage{Data: c.Data, List: &list}
	return proceedTemplate(c, NAppCatalog, &pageData)
}
