package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"log"
	//"encoding/json"
	//"regexp"
	//"math"
	//"strings"
	//	"os"
	//	"time"
	//"strings"
	//"bytes"
	//"github.com/DayLightProject/go-daylight/packages/consts"
	//	"math"
	//	"database/sql"
	//	"bytes"
)

func (p *Parser) MessageToAdminInit() error {
	fields := []map[string]string{{"encrypted_message": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["encrypted_message"] = utils.BinToHex(p.TxMaps.Bytes["encrypted_message"])
	p.TxMap["encrypted_message"] = utils.BinToHex(p.TxMap["encrypted_message"])
	return nil
}

func (p *Parser) MessageToAdminFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// в бинарном виде проверить можем только размер
	if len(p.TxMaps.Bytes["encrypted_message"]) > 20480 || len(p.TxMaps.Bytes["encrypted_message"]) == 0 {
		return p.ErrInfo("encrypted_message len")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["encrypted_message"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_message_to_admin"], "message_to_admin", p.Variables.Int64["limit_message_to_admin_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// пишется только в локальную таблицу юзера-отправителя и админа
func (p *Parser) MessageToAdmin() error {

	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		myId, err := p.Single("SELECT id FROM "+myPrefix+"my_admin_messages WHERE hex(encrypted) = ? AND status  =  'my_pending'", p.TxMap["encrypted_message"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if myId > 0 {
			// обновим статус в нашей локальной табле.
			err = p.ExecSql("UPDATE "+myPrefix+"my_admin_messages SET status = 'approved' WHERE hex(encrypted) = ? AND status = 'my_pending'", p.TxMap["encrypted_message"])
			if err != nil {
				return p.ErrInfo(err)
			}
		} else {
			err = p.ExecSql("INSERT INTO "+myPrefix+"my_admin_messages ( encrypted, status ) VALUES ( [hex], 'approved' )", p.TxMap["encrypted_message"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	// админ
	admin, err := p.GetAdminUserId()
	if err != nil {
		return p.ErrInfo(err)
	}
	if myUserId == admin {
		err = p.ExecSql("INSERT INTO x_my_admin_messages ( encrypted, type, user_id ) VALUES ( [hex], 'from_user', ? )", p.TxMap["encrypted_message"], p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) MessageToAdminRollback() error {

	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId {
		// обновим статус в нашей локальной табле.
		err = p.ExecSql("UPDATE "+myPrefix+"my_admin_messages SET status = 'my_pending' WHERE hex(message) = ? AND status = 'approved'", p.TxMap["encrypted_message"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// админ
	admin, err := p.GetAdminUserId()
	if err != nil {
		return p.ErrInfo(err)
	}
	if myUserId == admin {
		err = p.ExecSql("DELETE FROM x_my_admin_messages WHERE hex(encrypted) = ? AND type = 'from_user' AND user_id = ?", p.TxMap["encrypted_message"], p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.rollbackAI("x_my_admin_messages", 1)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) MessageToAdminRollbackFront() error {
	return p.limitRequestsRollback("message_to_admin")
}
