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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/EGaaS/go-egaas-mvp/packages/template"
)

// CommonPage is a structure with common information about the user and state
type CommonPage struct {
	//Lang   map[string]string
	Address   string
	WalletId  int64
	CitizenId int64
	StateId   int64
	StateName string
}

const ( // Type of pages
	// TPage - template page
	TPage = iota
	// TJson - ajax json request
	TJson
)

type pageTemplate struct {
	Template string // pattern name
	Name     string // method name
	Type     uint8  // 0 - Page, 1 - Json
}

var (
	globPages = make(map[string]*pageTemplate)
)

func newPage(name string, params ...string) {
	page := pageTemplate{Template: name}

	parts := strings.Split(name, `_`)
	for i := range parts {
		a := []rune(parts[i])
		a[0] = unicode.ToUpper(a[0])
		parts[i] = string(a)
	}
	for _, ival := range params {
		switch ival {
		case `json`:
			page.Type = TJson
		}
	}
	page.Name = strings.Join(parts, ``)
	globPages[name] = &page
}

func isPage(name string, itype uint8) bool {
	gp, ok := globPages[name]
	if ok && gp.Type != itype {
		ok = false
	}
	return ok
}

// CallPage calls the page controller
func CallPage(c *Controller, name string) string {
	html, err := CallMethod(c, globPages[name].Name)
	if err != nil {
		html = fmt.Sprintf(`{"error":%q}`, err)
		log.Error("err: %s / Controller: %s", html, name)
	}
	return html
}

// CallJSON calls the ajax controller
func CallJSON(c *Controller, name string) []byte {
	methodName := globPages[name].Name

	var (
		ptr         reflect.Value
		value       reflect.Value
		finalMethod reflect.Value
	)
	value = reflect.ValueOf(c)
	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(c))
		temp := ptr.Elem()
		temp.Set(value)
	}
	method := value.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	method = ptr.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	if finalMethod.IsValid() {
		x := finalMethod.Call([]reflect.Value{})
		jsonData, err := json.Marshal(x[0].Interface())
		if err == nil {
			return jsonData
		}
	}
	return []byte(`{"error":"system error"}`)
}

func proceedTemplate(с *Controller, html string, data interface{}) (string, error) {
	return template.ProceedTemplate(html, data)
}
