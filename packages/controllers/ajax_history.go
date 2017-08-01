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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

const aHistory = `ajax_history`

// HistoryJSON is a structure for the answer of ajax_history ajax request
type HistoryJSON struct {
	Draw     int                 `json:"draw"`
	Total    int                 `json:"recordsTotal"`
	Filtered int                 `json:"recordsFiltered"`
	Data     []map[string]string `json:"data"`
	Error    string              `json:"error"`
}

func init() {
	newPage(aHistory, `json`)
}

// AjaxHistory is a controller of ajax_history request
func (c *Controller) AjaxHistory() interface{} {
	var (
		history []map[string]string
		err     error
	)
	walletID := c.SessWalletID
	result := HistoryJSON{Draw: converter.StrToInt(c.r.FormValue("draw"))}
	length := converter.StrToInt(c.r.FormValue("length"))
	if length == -1 {
		length = 20
	}
	log.Debug("a/h walletId %s / c.SessAddress %s", walletID, c.SessAddress)
	if walletID != 0 {
		dltTransaction := &model.DltTransaction{}
		total, _ := dltTransaction.GetCount(walletID, walletID, c.SessAddress)
		result.Total = int(total)
		result.Filtered = int(total)
		if length != 0 {
			wt := &model.WalletedTransaction{}
			transactions, err := wt.Get(walletID, walletID, c.SessAddress, length, converter.StrToInt(c.r.FormValue("start")))
			if err != nil {
				log.Error("%s", err)
			}
			history := make([]map[string]string, 0)

			for _, transaction := range transactions {
				block := &model.Block{}
				block.GetMaxBlock()
				max := block.ID
				row := transaction.ToMap()
				row[`confirm`] = converter.Int64ToStr(max - transaction.BlockID)
				row[`sender_address`] = converter.AddressToString(transaction.Sw)
				recipient := string(transaction.Rw)
				if len(recipient) < 10 {
					recipient = string(transaction.RecepientWalletAddress)
				}
				row[`recipient_address`] = converter.AddressToString(converter.StringToAddress(recipient))
				history = append(history, row)
			}
		}
	}
	if err != nil {
		result.Error = err.Error()
	} else {
		if history == nil {
			history = []map[string]string{}
		}
		result.Data = history
	}
	return result
}
