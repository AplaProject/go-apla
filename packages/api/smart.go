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
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
)

type smartField struct {
	Name string `json:"name"`
	HTML string `json:"htmltype"`
	Type string `json:"txtype"`
	Tags string `json:"tags"`
}

type smartFieldsResult struct {
	Fields []smartField `json:"fields"`
	Name   string       `json:"name"`
	Active bool         `json:"active"`
}

//SignRes contains the data of the signature
type SignRes struct {
	Param string `json:"name"`
	Text  string `json:"text"`
}

// TxSignJSON is a structure for additional signs of transaction
type TxSignJSON struct {
	ForSign string    `json:"forsign"`
	Field   string    `json:"field"`
	Title   string    `json:"title"`
	Params  []SignRes `json:"params"`
}

// PrepareTxJSON is a structure for the answer of ajax_prepare_tx ajax request
type PrepareTxJSON struct {
	ForSign string            `json:"forsign"`
	Signs   []TxSignJSON      `json:"signs"`
	Values  map[string]string `json:"values"`
	Time    uint32            `json:"time"`
	//	Error   string            `json:"error"`
}

func getSmartContract(w http.ResponseWriter, r *http.Request, data *apiData) error {

	cntname := data.params[`name`].(string)
	contract := smart.GetContract(cntname, uint32(data.sess.Get(`state`).(int64)))
	if contract == nil {
		return errorAPI(w, fmt.Sprintf(`there is not %s contract`, cntname), http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	fields := make([]smartField, 0)
	result := smartFieldsResult{Fields: fields, Name: info.Name, Active: info.Active}

	if info.Tx != nil {
		for _, fitem := range *info.Tx {
			field := smartField{Name: fitem.Name, Type: fitem.Type.String(), Tags: fitem.Tags}

			if strings.Index(fitem.Tags, `hidden`) >= 0 || strings.Index(fitem.Tags, `signature`) >= 0 {
				field.HTML = `hidden`
			} else {
				for _, tag := range []string{`date`, `polymap`, `map`, `image`, `text`, `address`} {
					if strings.Index(fitem.Tags, tag) >= 0 {
						field.HTML = tag
						break
					}
				}
				if len(field.HTML) == 0 {
					if fitem.Type.String() == script.Decimal {
						field.HTML = `money`
					} else if fitem.Type.String() == `string` || fitem.Type.String() == `int64` || fitem.Type.String() == `float64` {
						field.HTML = `textinput`
					}
				}
			}
			fields = append(fields, field)
		}
	}
	data.result = result
	return nil
}

func validateSmartContract(data *apiData, result *PrepareTxJSON) (contract *smart.Contract, err error) {
	cntname := data.params[`name`].(string)
	state := data.sess.Get(`state`).(int64)
	contract = smart.GetContract(cntname, uint32(state))
	if contract == nil {
		return nil, fmt.Errorf(`there is not %s contract`, cntname)
	}

	if contract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
			if strings.Index(fitem.Tags, `image`) >= 0 || strings.Index(fitem.Tags, `crypt`) >= 0 {
				continue
			}
			if strings.Index(fitem.Tags, `signature`) >= 0 && result != nil {
				if ret := regexp.MustCompile(`(?is)signature:([\w_\d]+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
					pref := getPrefix(data)
					var value string
					value, err = sql.DB.Single(fmt.Sprintf(`select value from "%s_signatures" where name=?`, pref), ret[1]).String()
					if err != nil {
						break
					}
					if len(value) == 0 {
						err = fmt.Errorf(`%s is unknown signature`, ret[1])
						break
					}
					var sign TxSignJSON
					err = json.Unmarshal([]byte(value), &sign)
					if err != nil {
						break
					}
					sign.ForSign = fmt.Sprintf(`%d,%d`, (*result).Time, uint64(data.sess.Get(`wallet`).(int64)))
					for _, isign := range sign.Params {
						var val string

						if _, ok := data.params[isign.Param]; ok {
							val = strings.TrimSpace(data.params[isign.Param].(string))
						}
						sign.ForSign += fmt.Sprintf(`,%v`, val)
					}
					sign.Field = fitem.Name
					(*result).Signs = append((*result).Signs, sign)
				}
			} else {
				var val string

				if _, ok := data.params[fitem.Name]; ok {
					val = strings.TrimSpace(data.params[fitem.Name].(string))
				}
				if len(val) == 0 && !strings.Contains(fitem.Tags, `optional`) {
					err = fmt.Errorf(`%s is empty`, fitem.Name)
					break
				}
				if strings.Index(fitem.Tags, `address`) >= 0 {
					addr := converter.StringToAddress(val)
					if addr == 0 {
						err = fmt.Errorf(`Address %s is not valid`, val)
						break
					}
				}
				if fitem.Type.String() == script.Decimal {
					re := regexp.MustCompile(`^\d+$`) //`^\d+\.?\d+?$`
					if !re.Match([]byte(val)) {
						err = fmt.Errorf(`The value of money %s is not valid`, val)
						break
					}
				}
			}
		}
	}
	return
}

func txPreSmartContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var (
		result PrepareTxJSON
	)
	result.Time = uint32(time.Now().Unix())
	result.Values = make(map[string]string)
	contract, err := validateSmartContract(data, &result)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	fmt.Println(`preSmart`, contract)
	return nil
}

func txSmartContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	return nil
}
