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

package exchangeapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/shopspring/decimal"
)

type histOper struct {
	BlockID string `json:"block_id"`
	Dif     string `json:"dif"`
	Amount  string `json:"amount"`
	EGS     string `json:"egs"`
	Time    string `json:"time"`
	//	Wallet  string `json:"wallet"`
}

// History is an answer structure for history request
type History struct {
	Error string     `json:"error"`
	Items []histOper `json:"history"`
}

func history(r *http.Request) interface{} {
	var (
		result History
	)

	wallet := converter.StringToAddress(r.FormValue(`wallet`))
	if wallet == 0 {
		result.Error = `Wallet is invalid`
		return result
	}
	c, err := strconv.ParseInt(r.FormValue("count"), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, r.FormValue("count"))
	}
	count := int(c)
	if count == 0 {
		count = 50
	}
	if count > 200 {
		count = 200
	}
	list := make([]histOper, 0)
	current, err := model.GetOneRow(`select amount, rb_id from dlt_wallets where wallet_id=?`, wallet).String()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	rb, err := strconv.ParseInt(current[`rb_id`], 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, current[`rb_id`])
	}
	if len(current) > 0 && rb != 0 {
		balance, _ := decimal.NewFromString(current[`amount`])
		for len(list) < count && rb > 0 {
			var data map[string]string
			prev, err := model.GetOneRow(`select r.*, b.time from rollback as r
			left join block_chain as b on b.id=r.block_id
			where r.rb_id=?`, rb).String()
			if err != nil {
				result.Error = err.Error()
				return result
			}
			if err = json.Unmarshal([]byte(prev[`data`]), &data); err != nil {
				result.Error = err.Error()
				return result
			}
			rb, err = strconv.ParseInt(data[`rb_id`], 10, 64)
			if err != nil {
				logger.LogInfo(consts.StrtoInt64Error, data[`rb_id`])
			}
			//			fmt.Println(`DATA`, prev)
			if amount, ok := data[`amount`]; ok {
				var dif decimal.Decimal
				val, _ := decimal.NewFromString(amount)
				if balance.Cmp(val) > 0 {
					dif = balance.Sub(val)
				} else {
					dif = val.Sub(balance)
				}
				sign := `+`
				if balance.Cmp(val) < 0 {
					sign = `-`
				}
				timeInt, err := strconv.ParseInt(prev["time"], 10, 64)
				if err != nil {
					logger.LogInfo(consts.StrtoInt64Error, prev["time"])
				}
				dt := time.Unix(timeInt, 0)

				list = append(list, histOper{BlockID: prev[`block_id`], Dif: sign + converter.EGSMoney(dif.String()),
					Amount: balance.String(), EGS: converter.EGSMoney(balance.String()), Time: dt.Format(`02.01.2006 15:04:05`)})
				balance = val

			}
		}
	}
	if rb == 0 {
		first, err := model.GetOneRow(`select * from dlt_transactions where recipient_wallet_id=? order by id`, wallet).String()
		if err != nil {
			result.Error = err.Error()
			return result
		}
		if len(first) > 0 {
			timeInt, err := strconv.ParseInt(first["time"], 10, 64)
			if err != nil {
				logger.LogInfo(consts.StrtoInt64Error, first["time"])
			}
			dt := time.Unix(timeInt, 0)
			list = append(list, histOper{BlockID: first[`block_id`], Dif: `+` + converter.EGSMoney(first[`amount`]),
				Amount: first[`amount`],
				EGS:    converter.EGSMoney(first[`amount`]), Time: dt.Format(`02.01.2006 15:04:05`)})
		}
	}
	result.Items = list

	return result
}
