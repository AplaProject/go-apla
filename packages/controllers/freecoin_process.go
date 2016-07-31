package controllers

import (
	"encoding/json"
	"fmt"
//	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
/*	"net/http"
	"os"
	"time"*/
)

var (
	gBlockId  int64  // Блок в начале
	gTime     int64  // Время начала
)

func (c *Controller) FreecoinProcess() (string, error) {
	var ( err error
		  curBlock int64
	)
		
	status := `wait`
	result := func() (string, error) {
		ret := map[string]string{"status": status, "error" : "" }
		if err != nil {
			ret["status"] = "error"
			ret["error"] = err.Error()
		}
		resultJ, _ := json.Marshal(ret)
//		fmt.Println( `FreeCoin`, status, curBlock, err )
		return string(resultJ), nil
	} 
	
	if c.r.FormValue("first") == `true` {
		if gBlockId, err = c.DCDB.GetBlockId(); err != nil {
			return result()
		}
		gTime = utils.Time()	
	} else {
		now := utils.Time()
		if curBlock, err = c.DCDB.GetBlockId(); err != nil {
			return result()
		}
		if now - gTime > 60 && curBlock - gBlockId > 3 {
			status = `reload`
		}
		var last_tx []map[string]string
		if last_tx, err = c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"NewRestrictedPromisedAmount"}), 
		                    1, c.TimeFormat); err != nil {
			return result()	
		}
		if len(last_tx) > 0 {
			if len( last_tx[0][`txerror`] ) > 0 {
				status = `error`
				err = fmt.Errorf( last_tx[0][`txerror`] )
			} else if last_tx[0][`block_id`] != `0` {
				status = `success`
			}
		}
	}
	return result()
}
