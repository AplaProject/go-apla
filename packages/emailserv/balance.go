// emailserv
package main

import (
	"net/http"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
//	"html/template"
	"bytes"
//	"io"
	"fmt"
	"strings"
	"sync"
)

type balanceTask struct {
	UserId  int64
	Error   error
}

var (
	queueBalance  []*balanceTask
	bCurrent  int   
	bMutex    sync.Mutex
)

func init() {
	queueBalance = make([]*balanceTask, 0, 200 )
}

func balanceDaemon() {
	for {
		if bCurrent < len( queueBalance ) {
			BalanceProceed()
		}
		utils.Sleep( 10 )
	}
}

func BalanceProceed() {
	bMutex.Lock()
	task := queueBalance[ bCurrent ]
	// Защита от повторной рассылки
	for i:=0; i<bCurrent; i++ {
		if queueBalance[i].Error == nil && queueBalance[i].UserId == task.UserId {
			task.Error = fmt.Errorf(`It has already been sent`)
			bCurrent++
			bMutex.Lock()
			return
		}
	}
	data, err := CheckUser( task.UserId )
	if err != nil {
		task.Error = err
	} else {
		getBalance( task.UserId, &data )
		data[`Money`] = len(data[`List`].(map[int64]*infoBalance))
		if data[`Money`].(int) == 0 {
//	    	if data[`Money`].(int) != 1 {
//			} else if data[`Money`].(int) > 1 {	task.Error = fmt.Errorf(`Sent yesterday`) }
			task.Error = fmt.Errorf(`No dcoins`)
			bCurrent++
			bMutex.Unlock()
			return	
		}
//		data[`nobcc`] = true
		
		if data[`List`].(map[int64]*infoBalance)[72] != nil && data[`List`].(map[int64]*infoBalance)[72].Summary > 100 {
					task.Error = fmt.Errorf(`Too much Summary=%f`, data[`List`].(map[int64]*infoBalance)[72].Summary )
		} else if EmailUser( task.UserId, data, utils.ECMD_BALANCE ) {
			icur := int64(72)
			if data[`List`].(map[int64]*infoBalance)[icur] == nil || data[`List`].(map[int64]*infoBalance)[icur].Tdc == 0 {
				for icur := range data[`List`].(map[int64]*infoBalance) {
					if icur != 1 {
						break
					}
				}
			}
			if data[`List`].(map[int64]*infoBalance)[icur] != nil {
				task.Error = fmt.Errorf(`Sent Currency=%d Wallet=%f Tdc=%f Summary=%f`, icur,
					data[`List`].(map[int64]*infoBalance)[icur].Wallet,
					data[`List`].(map[int64]*infoBalance)[icur].Tdc,
					data[`List`].(map[int64]*infoBalance)[icur].Summary )
			}
		} else {
			task.Error = fmt.Errorf(`Error sending`)
		}
	}
	bCurrent++
	bMutex.Unlock()
}

type infoBalance struct {
	Currency   string
	CurrencyId int64
	Wallet     float64
	Tdc        float64
	Summary    float64
	Promised   float64
	Restricted float64
	Top        float64
	Dif        float64
}

func getBalance( userId int64, data *map[string]interface{} ) error {
	
	list := make(map[int64]*infoBalance)
	if wallet, err := utils.DB.GetBalances(userId); err == nil {
		for _, iwallet := range wallet {
			list[iwallet.CurrencyId] = &infoBalance{ CurrencyId: iwallet.CurrencyId,
			                Wallet: utils.Round(iwallet.Amount, 6) }
		}
	} else {
		return err
	}
	if vars, err := utils.DB.GetAllVariables(); err == nil {
		if _, dc, _, err := utils.DB.GetPromisedAmounts( userId, vars.Int64["cash_request_time"]); err == nil {
			for _, idc := range dc {
				if _, ok:= list[idc.CurrencyId]; ok {
					list[idc.CurrencyId].Tdc += utils.Round(idc.Tdc,6)
					list[idc.CurrencyId].Promised += idc.Amount
				} else {
					list[idc.CurrencyId] = &infoBalance{ CurrencyId: idc.CurrencyId,
			                Promised: idc.Amount, Tdc: utils.Round(idc.Tdc, 6) }
				}
			}
		} else {
			return err
		}
	} else {
		return err
	}
//	c := new(controllers.Controller)
//	c.SessUserId = userId
//	c.DCDB = utils.DB
	if profit,_, err := utils.DB.GetPromisedAmountCounter(userId); err == nil && profit > 0 {
		currency := int64(72)
		if _, ok:= list[currency]; ok {
			list[currency].Restricted = utils.Round( profit - 30, 6)
		} else {
			list[currency] = &infoBalance{ CurrencyId: currency,
			                Restricted: utils.Round( profit - 30, 6) }
		}
	}
	forjson := make( map[string]float64 )
	prevjson := make( map[string]float64 )

	prev,_ := GDB.Single(`select balance from balance where user_id=? order by id desc`, userId ).String()
	if len(prev) > 0 {
		json.Unmarshal( []byte(prev), &prevjson )
	}
	for i := range list {
		list[i].Currency,_ = utils.DB.Single(`select name from currency where id=?`, list[i].CurrencyId ).String()
		list[i].Summary = utils.Round( list[i].Wallet + list[i].Tdc + list[i].Restricted, 6 )
		curstr := utils.Int64ToStr(list[i].CurrencyId)
		forjson[ curstr ] = list[i].Summary
		if dif, ok := prevjson[curstr]; ok {
			list[i].Dif = RoundMoney(list[i].Summary - dif)
		}
		list[i].Top = RoundMoney(list[i].Summary)
	}
	out,_ := json.Marshal( forjson )
	if len(out) > 0 {
		GDB.ExecSql(`insert into balance ( user_id, balance, uptime) values( ?, ?,  datetime('now'))`,
		                  userId, out )
	}
	(*data)[`List`] = list
	return nil
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	
	_,_,ok := checkLogin( w, r )
	if !ok {
		return
	}
	data := make( map[string]interface{})
	out := new(bytes.Buffer)
	r.ParseForm()
	users := strings.Split( r.PostFormValue(`idusers`), `,` )
	clear := r.PostFormValue(`clearqueueBalance`)
	if len(clear) > 0 {
		bMutex.Lock()
		queueBalance = queueBalance[:0]
		bCurrent = 0
		data[`message`] = `Очередь очищена`
		bMutex.Unlock()
	} else if len(users) > 0 && len(users[0]) > 0 {
		if users[0] == `*` {
			if list, err := GDB.GetAll("select user_id, email from users where verified >= 0", -1); err == nil {
				users = users[:0]
				for _, icur := range list {
					users = append( users, icur[`user_id`])
				}
			}
		}
		bMutex.Lock()
		for _, iduser := range users { 
			userId := utils.StrToInt64( iduser )
			if !utils.InSliceInt64(userId, []int64{ 30, 158, 879, 705, 385, 490, 111 }) {
				queueBalance = append( queueBalance, &balanceTask{ UserId: userId })		
			}		
		}
		bMutex.Unlock()
		http.Redirect(w, r, `/` + GSettings.Admin + `/balance`, http.StatusFound )
	} else {
		data[`message`] = `Не указаны пользователи`
	}
	data[`count`],_ = GDB.Single(`select count(id) from users where verified>=0`).Int64()
	data[`tasks`] = queueBalance[:bCurrent]
	data[`todo`] = len(queueBalance) - bCurrent
	if err := GPageTpl.ExecuteTemplate(out, `balance`, data); err != nil {
		w.Write( []byte(err.Error()))
		return
	}
	w.Write(out.Bytes())
}
