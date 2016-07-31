// stat
package stat

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
	"net/http"
	"encoding/json"
	"io/ioutil"
/*	"strings"
	"time"*/

//	"fmt"
)

const (
//	STAT_SERVER = `http://localhost:8091`
	STAT_SERVER = `http://pool.dcoin.club:8201`
)

type CurrencyBalance struct {
	CurrencyId int64    `json:"cur_id"`
	Wallet     float64  `json:"wallet"`
	Tdc        float64  `json:"tdc"` 
	Promised   float64  `json:"promised"`
	Restricted float64  `json:"restricted"`
	Summary    float64  `json:"summary"`
}

type InfoBalance struct {
	Currencies map[string] *CurrencyBalance
	Time       int64
}

type HistoryBalance struct {
	Success  bool            `json:"success"`
	Error    string          `json:"error"`
	History  []*InfoBalance  `json:"history"`
}

type ResultBalance struct {
	CurrencyBalance
	Currency   string
	Top        float64
}

type ListBalance map[int64][]*ResultBalance

var (
	cashReqTime  int64
	currencies   map[int64]string = make(map[int64]string)
)

func SetCashReqTime() error {
	if vars, err := utils.DB.GetAllVariables(); err != nil {
		return err
	} else {
		cashReqTime = vars.Int64["cash_request_time"]
	}
	return nil
}

func RoundMoney(in float64, num int ) (out float64) {
	off := float64(10)
	for k:=0; k<num + 1; k++ {
		if in < off {
			out = utils.Round( in, num - k )
			break
		}
		off *= 10
	}
	if out == 0 {
		out = utils.Round( in, 0 )
	}
	return
}

func GetBalance(userId int64) (*InfoBalance,error) {
	
	ret := new(InfoBalance)		
	list := make(map[string]*CurrencyBalance)

	if wallet, err := utils.DB.GetBalances(userId); err == nil {
		for _, iwallet := range wallet {
			list[utils.Int64ToStr(iwallet.CurrencyId)] = &CurrencyBalance{ CurrencyId: iwallet.CurrencyId,
			                Wallet: utils.Round(iwallet.Amount, 6) }
		}
	} else {
		return ret, err
	}
	if cashReqTime == 0 {
		if err := SetCashReqTime(); err!=nil {
			return ret, err
		}
	}
	if _, dc, _, err := utils.DB.GetPromisedAmounts(userId, cashReqTime); err == nil {
		for _, idc := range dc {
			currency := utils.Int64ToStr(idc.CurrencyId)
			if _, ok:= list[currency]; ok {
				list[currency].Tdc += utils.Round(idc.Tdc,6)
				list[currency].Promised += idc.Amount
			} else {
				list[currency] = &CurrencyBalance{ CurrencyId: idc.CurrencyId,
			                Promised: idc.Amount, Tdc: utils.Round(idc.Tdc, 6) }
			} 
		}
	} else {
		return ret,err
	}

	if profit,_, err := utils.DB.GetPromisedAmountCounter(userId); err == nil && profit > 0 {
		currency := `72`
		if _, ok:= list[currency]; ok {
			list[currency].Restricted = utils.Round( profit - 30, 6)
		} else {
			list[currency] = &CurrencyBalance{ CurrencyId: utils.StrToInt64(currency),
			                Restricted: utils.Round( profit - 30, 6) }
		}
	}
	for i := range list {
		list[i].Summary = utils.Round( list[i].Wallet + list[i].Tdc + list[i].Restricted, 6 )
	}
	ret.Currencies = list
	ret.Time = time.Now().Unix()
	return ret,nil
}

func currencyToResult(cur *CurrencyBalance, result *ResultBalance ) {
	result.CurrencyId = cur.CurrencyId
	result.Wallet = cur.Wallet
	result.Tdc = cur.Tdc
	result.Promised = cur.Promised
	result.Restricted = cur.Restricted
	result.Summary = cur.Summary
	if _,ok := currencies[cur.CurrencyId];!ok {
		currencies[cur.CurrencyId],_ = utils.DB.Single(`select name from currency where id=?`, cur.CurrencyId ).String()
	}
	result.Currency = currencies[cur.CurrencyId]
	result.Summary = utils.Round( cur.Wallet + cur.Tdc + cur.Restricted, 6 )
	result.Top = RoundMoney(cur.Summary, 5)
}

func TodayBalance(userId int64) (*ListBalance, error) {
	list := make( ListBalance )
	
	info,err := GetBalance(userId)
	if err == nil {
		for _,icur := range info.Currencies {
			var result ResultBalance 
			
			currencyToResult( icur, &result )
			if result.Summary > 0 {
				list[icur.CurrencyId] = make([]*ResultBalance, 1)
				list[icur.CurrencyId][0] = &result
			}
		}
	}
	return &list, err
}


func GetHistoryBalance(list *ListBalance, userId int64) (int, error) {
	var info HistoryBalance
	resp, err := http.Get( STAT_SERVER + `/balance?user_id=` + utils.Int64ToStr(userId))
	if err != nil {
		return 0,err
	}
	defer resp.Body.Close()
	history, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0,err
	}
	if err = json.Unmarshal(history, &info); err != nil {
		return 0,err
	}
	if  info.History != nil && len(info.History) > 0 {
		for key := range *list {
			cur := utils.Int64ToStr(key)
			for _,ihist := range info.History {
				if icur,ok:=ihist.Currencies[cur]; ok {
					var result ResultBalance
					currencyToResult( icur, &result )
					(*list)[key] = append((*list)[key], &result)
				}
			}
		}
	}
	max := 2
	for key := range *list {
		if count := len( (*list)[key] ); count > max {
			max = count
		}
	}	
	for key := range *list {
		if count := len( (*list)[key] ); count < max {
			for ;count < max; count++ {
				(*list)[key] = append((*list)[key], &ResultBalance{})
			}
		}
	}	
	
	return max-1,nil
}
