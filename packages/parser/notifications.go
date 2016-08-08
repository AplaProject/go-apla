// notifications
package parser

import (
	"fmt"
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func  (p *Parser) isNotify(userId int64) bool {
	if val,ok := p.ConfigIni["notify"]; ok && val == `1`{
		return true
	}
	if myBlockId, err := utils.DB.GetMyBlockId(); err != nil || myBlockId > p.BlockData.BlockId {
		return false
	}
	
	myUsersIds,_ := utils.DB.GetCommunityUsers()
	if len(myUsersIds) == 0 {
		if myUserId,_ := utils.DB.GetMyUserId(""); myUserId > 0 {
			myUsersIds = append(myUsersIds, myUserId)
		}
	}
	if userId == 0 || utils.InSliceInt64(userId, myUsersIds) {
		return true
	}
	return false
}

func  (p *Parser) nfyRollback( blockId int64 ) {
	if !p.isNotify(0) {
		return
	}
	p.ExecSql( `delete from notifications where block_id=?`, blockId )
}


func (p *Parser) insertNotify( userId int64, cmdId int, params string) {
	p.ExecSql("insert into notifications (user_id, block_id, cmd_id, params, isread) VALUES (?, ?, ?, ?,1)", 
	          userId, p.BlockData.BlockId, cmdId, params )
}

func  (p *Parser) nfyCashRequest( userId int64,  cashRequest *utils.TypeNfyCashRequest ) {
	if !p.isNotify(userId) {
		return
	}
	params,err := json.Marshal( cashRequest ) 
	if err != nil {
		params = []byte(fmt.Sprintf( `{"error": "%s"}`, err ))
	}
	p.insertNotify( userId, utils.ECMD_CASHREQ, string(params))
}

func  (p *Parser) nfyRefReady( userId int64, refId int64 ) {
	if !p.isNotify(userId) {
		return
	}
	p.insertNotify( userId, utils.ECMD_REFREADY, fmt.Sprintf( `{"refid": "%d"}`, refId ))
}

func  (p *Parser) nfyStatus( userId int64, status string ) {
	if !p.isNotify(userId) {
		return
	}
	p.insertNotify( userId, utils.ECMD_CHANGESTAT, fmt.Sprintf( `{"status": "%s"}`, status ))
}

func  (p *Parser) nfySent( userId int64, tns *utils.TypeNfySent ) {
	if !p.isNotify(userId) {
		return
	}
	params,err := json.Marshal( tns ) 
	if err != nil {
		params = []byte(fmt.Sprintf( `{"error": "%s"}`, err ))
	}
	p.insertNotify( userId, utils.ECMD_DCSENT, string(params))
}

func  (p *Parser) nfyCame( userId int64, tnc *utils.TypeNfyCame ) {
	if !p.isNotify(userId) || tnc.Amount == 0.0 {
		return
	}
	params,err := json.Marshal( tnc ) 
	if err != nil {
		params = []byte(fmt.Sprintf( `{"error": "%s"}`, err ))
	}
	p.insertNotify( userId, utils.ECMD_DCCAME, string(params))
}
