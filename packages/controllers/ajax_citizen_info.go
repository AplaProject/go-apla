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
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/DayLightProject/go-daylight/packages/utils"
)

const ACitizenInfo = `ajax_citizen_info`

type FieldInfo struct {
	Name     string `json:"name"`
	HtmlType string `json:"htmlType"`
	TxType   string `json:"txType"`
	Title    string `json:"title"`
}

type CitizenInfoJson struct {
	Result bool   `json:"result"`
	Error  string `json:"error"`
}

func init() {
	newPage(ACitizenInfo, `json`)
}

func (c *Controller) AjaxCitizenInfo() interface{} {
	var (
		result CitizenInfoJson
		err    error
	)
	c.w.Header().Add("Access-Control-Allow-Origin", "*")
	stateCode := utils.StrToInt64(c.r.FormValue(`stateId`))
	statePrefix, err := c.GetStatePrefix(stateCode)

	fmt.Println(`1`, statePrefix)
	field, err := c.Single(`SELECT value FROM ` + statePrefix + `_state_settings where parameter='citizen_fields'`).String()
	fmt.Println(`2`, field, err)
	if err == nil {
		var (
			fields []FieldInfo
		)
		vals := make(map[string]string)
		if err = json.Unmarshal([]byte(field), &fields); err == nil {
			fmt.Println(`3`, field, err)
			time := c.r.FormValue(`time`)
			walletId := c.r.FormValue(`walletId`)
			for _, ifield := range fields {
				vals[ifield.Name] = c.r.FormValue(ifield.Name)
			}

			data, err := c.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM dlt_wallets WHERE wallet_id = ?", walletId).String()
			if err == nil {
				fmt.Println(`4`, data)

				var PublicKeys [][]byte
				PublicKeys = append(PublicKeys, []byte(data["public_key_0"]))
				forSign := fmt.Sprintf("CitizenInfo,%d,%d", time, walletId)
				sign, err := hex.DecodeString(c.r.FormValue(`signature1`))
				fmt.Println(`5`, err)

				if err == nil {
					checkSignResult, err := utils.CheckSign(PublicKeys, forSign, sign, false)
					fmt.Println(`SIGNATURE`, checkSignResult, err)
					/*			if err != nil {
								return p.ErrInfo(err)
							}*/
				}
			}
		}
	}
	/*	if err == nil {
		request, err := c.Single(`SELECT block_id FROM `+statePrefix+`_citizenship_requests where dlt_wallet_id=?`, c.SessWalletId).Int64()
		if err == nil {
			if request > 0 {
				var state map[string]string
				state, err = c.OneRow(`select * from states where state_id=?`, stateCode).String()
				if len(state[`host`]) == 0 {
					if walletId := utils.StrToInt64(state[`delegate_wallet_id`]); walletId > 0 {
						host, _ = c.Single(`select host from dlt_wallets where wallet_id=?`, walletId).String()
					}
					if len(host) == 0 {
						if stateId := utils.StrToInt64(state[`delegate_state_id`]); stateId > 0 {
							host, err = c.Single(`select host from states where state_id=?`, stateId).String()
						}
					}
				}
				result.Time = utils.Time()
				if len(host) > 0 {
					if !strings.HasPrefix(host, `http`) {
						host = `http://` + host
					}
					if !strings.HasSuffix(host, `/`) {
						host += `/`
					}
					result.TypeName = `NewCitizen`
					result.TypeId = utils.TypeInt(result.TypeName)
				}
				result.Host = host
			}
		} else {
			result.Error = err.Error()
		}
	}*/
	fmt.Println(`Error`, err)
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
