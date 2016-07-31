// statserv
package main

import (
	"github.com/DayLightProject/go-daylight/packages/stat"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
	//	"fmt"
	"encoding/json"
	"log"
)

func daemon() {
	var (
		cur, max int64
		err      error
		first    bool = true
		iBalance *stat.InfoBalance
		pause    uint32 = 10
		timeClear  time.Time
	)
	
	for {
		if cur == max {
			if max, err = utils.DB.Single(`select user_id from users order by user_id desc`).Int64(); err != nil {
				log.Println(`Error`, err)
			}
			cur = 1
			if first {
				if last,err := GDB.Single(`select user_id from balance where date(uptime)=date('now') order by id desc`).Int64(); 
				      err == nil && last > 0 {
					cur = last
				}
				first = false
				timeClear = time.Now()
			}
			if err = stat.SetCashReqTime(); err != nil {
				log.Println(`Error`, err)
			}
			pause = GSettings.Period * 3600 / uint32(max)
			if pause == 0 {
				pause = 1
			}
			log.Println(`Start loop`, cur, `/`, max, `/`, pause, `sec`)
//			max = 20
		}
		if iBalance, err = stat.GetBalance(cur); err != nil {
			log.Println(err)
		} else if len(iBalance.Currencies) > 0 {
			if idExist, err := GDB.Single(`select id from balance where user_id=? and date(uptime)=date('now')`,
				cur).Int64(); err == nil {
				if out, err := json.Marshal(iBalance); err == nil {
					if idExist > 0 {
						err = GDB.ExecSql(`update balance set data=?, uptime=datetime('now') where id=?`, out, idExist)
					} else {
						err = GDB.ExecSql(`insert into balance ( user_id, data, uptime) values( ?, ?,  datetime('now'))`,
							cur, out)
					}
					if err != nil {
						log.Println(err)
					}
				} else {
					log.Println(err)
				}
			} else {
				log.Println(err)
			}
		}
		cur++
		if time.Now().Sub( timeClear ).Hours() > 12 {
			GDB.ExecSql(`delete from balance where date( uptime, '+14 day' ) < date('now')`)
			GDB.ExecSql(`delete from req_balance where date( uptime, '+14 day' ) < date('now')`)
			cbal,_ := GDB.Single(`select count(id) from balance`).Int64()
			creq,_ := GDB.Single(`select count(id) from req_balance`).Int64()
			log.Println(`Delete old records: Balance`, cbal, `/ Req_balance`, creq)
			timeClear = time.Now()
		}
		time.Sleep(time.Duration(pause) * time.Second)
	}
}
