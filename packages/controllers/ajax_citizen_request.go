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
	//	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const ACitizenRequest = `ajax_citizen_request`

type CitizenRequestJson struct {
	Host string `json:"host"`
	Time int64  `json:"time"`
	//	TypeName string `json:"type_name"`
	//	TypeId   int64  `json:"type_id"`
	Error string `json:"error"`
}

func init() {
	newPage(ACitizenRequest, `json`)
}

func (c *Controller) AjaxCitizenRequest() interface{} {
	var (
		result CitizenRequestJson
		err    error
		//		host   string
	)

	stateCode := int64(1) // utils.StrToInt64(c.r.FormValue(`state_id`))
	//	_, err = c.GetStateName(stateCode)
	//	if err == nil {
	request, err := c.Single(`SELECT block_id FROM "`+utils.Int64ToStr(stateCode)+`_citizenship_requests" where dlt_wallet_id=?`, c.SessWalletId).Int64()
	if err == nil {
		if request > 0 {
			/*				var state map[string]string
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
								//					result.TypeName = `NewCitizen`
								//					result.TypeId = utils.TypeInt(result.TypeName)
							}*/
			result.Host = `/` //host
		}
	} else {
		result.Error = err.Error()
	}
	//	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
