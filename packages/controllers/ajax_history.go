package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/lib"
	"fmt"
)

const AHistory = `ajax_history`

type HistoryJson struct {
	Draw   int                  `json:"draw"`
	Total  int                  `json:"recordsTotal"`
	Filtered int                `json:"recordsFiltered"`
	Data   []map[string]string  `json:"data"`
	Error  string               `json:"error"`
}

func init() {
	newPage(AHistory, `json`)
}

func (c *Controller) AjaxHistory() interface{} {
	var ( history []map[string]string
	      err error 
	)
	walletId := c.SessWalletId
	result := HistoryJson{Draw: utils.StrToInt(c.r.FormValue("draw"))}
	length := utils.StrToInt(c.r.FormValue("length"))
	if length == -1 {
		length = 20
	}
	limit := fmt.Sprintf( `%d,%d`, utils.StrToInt(c.r.FormValue("start")), length)
	if walletId > 0 {
		total,_ := c.Single(`SELECT count(id) FROM dlt_transactions where sender_wallet_id=? OR 
		                       recipient_wallet_id=?`, walletId, walletId).Int64()	
		result.Total = int(total)
		result.Filtered = int(total)
		if length != 0 {
			history,err = c.GetAll( `SELECT d.*, w.address as sw, wr.address as rw FROM dlt_transactions as d
		        left join dlt_wallets as w on w.wallet_id=d.sender_wallet_id
		        left join dlt_wallets as wr on wr.wallet_id=d.recipient_wallet_id
				where sender_wallet_id=? OR 
		        recipient_wallet_id=? order by d.id desc limit ` + limit, -1, walletId, walletId )
		
			for ind := range history {
				max,_ := c.Single(`select max(id) from block_chain order by id desc`).Int64()
				history[ind][`confirm`] = utils.Int64ToStr(max - utils.StrToInt64(history[ind][`block_id`]))
				history[ind][`sender_address`] = lib.BytesToAddress([]byte(history[ind][`sw`]))
				history[ind][`recipient_address`] = lib.BytesToAddress([]byte(history[ind][`rw`]))
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
