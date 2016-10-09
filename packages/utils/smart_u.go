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

package utils

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/DayLightProject/go-daylight/packages/script"
	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/static"
)

type FieldInfo struct {
	Name     string `json:"name"`
	HtmlType string `json:"htmlType"`
	TxType   string `json:"txType"`
	Title    string `json:"title"`
	Value    string `json:"value"`
}

type FormCommon struct {
	//Lang   map[string]string
	/*	Address      string
		WalletId     int64
		CitizenId    int64
		StateId      int64
		StateName    string*/
	CountSignArr []byte
}

type FormInfo struct {
	TxName string
	Fields []FieldInfo
	Data   FormCommon
}

func init() {
	smart.Extend(&script.ExtendData{map[string]interface{}{
		"Balance":    Balance,
		"StateParam": StateParam,
		/*		"DBInsert":   DBInsert,
		 */
	}, map[string]string{
	//		`*parser.Parser`: `parser`,
	}})
}

// Reading and compiling contracts from smart_contracts tables
func LoadContracts() (err error) {
	var states []map[string]string
	prefix := []string{`global`}
	states, err = DB.GetAll(`select id from system_states order by id`, -1)
	if err != nil {
		return err
	}
	for _, istate := range states {
		prefix = append(prefix, istate[`id`])
	}
	for _, ipref := range prefix {
		if err = LoadContract(ipref); err != nil {
			return err
		}
	}
	return
}

// Reading and compiling contract of new state
func LoadContract(prefix string) (err error) {
	var contracts []map[string]string
	contracts, err = DB.GetAll(`select * from "`+prefix+`_smart_contracts" order by id`, -1)
	if err != nil {
		return err
	}
	for _, item := range contracts {
		if err = smart.Compile(item[`value`]); err != nil {
			return
		}
	}
	return
}

func Balance(wallet_id int64) (float64, error) {
	return DB.Single("SELECT amount FROM dlt_wallets WHERE wallet_id = ?", wallet_id).Float64()
}

func StateParam(idstate int64, name string) (string, error) {
	return DB.Single(`SELECT value FROM "`+Int64ToStr(idstate)+`_state_parameters" WHERE name = ?`, name).String()
}

func TxForm(name string) string {
	contract := smart.GetContract(name)
	if contract == nil || contract.Block.Info.(*script.ContractInfo).Tx == nil {
		return fmt.Sprintf(`there is not %s contract or parameters`, name)
	}
	funcMap := template.FuncMap{
		"sum": func(a, b interface{}) float64 {
			return InterfaceToFloat64(a) + InterfaceToFloat64(b)
		},
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	data, err := static.Asset("static/tx_form.html")

	sign, err := static.Asset("static/signatures_new.html")
	if err != nil {
		return fmt.Sprint(err.Error())
	}

	t := template.New("template").Funcs(funcMap)
	t, err = t.Parse(string(data))
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	t = template.Must(t.Parse(string(sign)))

	b := new(bytes.Buffer)
	finfo := FormInfo{TxName: name, Fields: make([]FieldInfo, 0), Data: FormCommon{
		CountSignArr: []byte{1}}}
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		if fitem.Type.String() == `string` {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "textinput",
				TxType: fitem.Type.String(), Title: fitem.Name})
		}
	}

	if err = t.Execute(b, finfo); err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return b.String()
}
