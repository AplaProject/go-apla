// check_blocks
package daemons

import (
	"fmt"
	"time"
	"strings"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
/*	"errors"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
	"github.com/DayLightProject/go-daylight/packages/static"
	_ "github.com/lib/pq"
	"os"*/
)

var (
	checkId    int64     // The latest checked block
	checkTime  time.Time // The time of the previous comparison
	prevNotify time.Time // The time of previous notify email 
)

func CheckBlocks() {
	defer time.AfterFunc( 30*time.Second, CheckBlocks )
	if utils.DB == nil || utils.DB.DB == nil {
		return
	}
	current, err := utils.DB.GetBlockId()
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
	}
	if current - checkId < 5 && checkTime.Add(15*time.Minute).After( time.Now()){
		return
	}
	hashes := make( map[string]*map[string]string )
	
	local,err := utils.DB.GetAll(`select id,hash from block_chain order by id desc`, 5 )
	if err != nil {
		return
	}
	current = 0
	localhash := make( map[string]string)
	for _, val := range local {
		localhash[val[`id`]] = fmt.Sprintf("%x", val[`hash`])
		if utils.StrToInt64(val[`id`]) > current {
			current = utils.StrToInt64(val[`id`])
		}
	}	

	q := "SELECT http_host FROM miners_data WHERE miner_id > 0 GROUP BY http_host"
	if configIni["db_type"] == "postgresql" {
		q = "SELECT DISTINCT ON (http_host) http_host FROM miners_data WHERE miner_id > 0"
	}
	hosts, err := utils.DB.GetAll( q, 20 )
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
	}
	for _, item := range hosts {
		host := item[`http_host`]
		if !strings.HasPrefix( host, `http`) {
			continue
		}
		jsonData, err := utils.GetHttpTextAnswer( host + "/ajaxjson?controllerName=CheckHash&block_id="+
		                                          utils.Int64ToStr(current))
		if err != nil {
			continue
		}
		var jsonMap map[string]string
		err = json.Unmarshal([]byte(jsonData), &jsonMap)
		if err != nil || jsonMap == nil {
			continue
		}
		hashes[host] = &jsonMap
	}
	allmin := current  // последний нормальный блок
	forks := make([]string, 0 )
	for hostname,item := range hashes {
		minid := current + 1
		for block,hash := range localhash {
			blockId := utils.StrToInt64(block)
			if ival,ok := (*item)[block]; ok && ival != hash && blockId < minid {
				minid = blockId
			}
		}	
		if minid > current {
			continue
		} else if minid == current - 4  {
			// Notification
			forks = append(forks, fmt.Sprintf( `%s - %d`, hostname, minid ))
		} else {
			if allmin > minid - 1 {
				allmin = minid - 1 
			}
		} 
	}
	if len(forks) > 0 && prevNotify.Add(2*time.Hour).Before( time.Now()) {
		logger.Error(`Fork of Blockchain %s`, forks)
		userId,_ := utils.DB.GetMyUserId(``)
		email,_ := utils.DB.Single(`select pool_email from config`).String()
		if len(email) > 0 {
			params := map[string]string{ `user_id`: utils.Int64ToStr(userId),
		                  `forks`: strings.Join( forks, `, ` ) }
			err = utils.SendEmail( email, utils.EXCHANGE_USER, utils.ECMD_FORKBLOCK, &params )
			if err == nil {
				prevNotify = time.Now()
			}
		}
	}
	
	checkTime = time.Now()
	checkId = allmin
}
