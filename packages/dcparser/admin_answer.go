package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) AdminAnswerInit() error {

	fields := []map[string]string{{"to_user_id": "int64"}, {"encrypted_message": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["encrypted_message"] = utils.BinToHex(p.TxMaps.Bytes["encrypted_message"])
	p.TxMap["encrypted_message"] = utils.BinToHex(p.TxMap["encrypted_message"])
	return nil
}

func (p *Parser) AdminAnswerFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.TxMaps.Bytes["encrypted_message"]) > 20480 {
		return p.ErrInfo("len encrypted_message>20480")
	}
	verifyData := map[string]string{"to_user_id": "user_id"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["to_user_id"], p.TxMap["encrypted_message"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) AdminAnswer() error {

	// проверим, не наш ли это user_id
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxMaps.Int64["to_user_id"] == myUserId && myBlockId <= p.BlockData.BlockId {
		err = p.ExecSql("INSERT INTO "+myPrefix+"my_admin_messages ( encrypted, type, status ) VALUES ( [hex], 'from_admin', 'approved' )", p.TxMaps.Bytes["encrypted_message"])
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
		err = p.ExecSql("UPDATE x_my_admin_messages SET status = 'approved' WHERE hex(encrypted) = ? AND status = 'my_pending'", p.TxMaps.Bytes["encrypted_message"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) AdminAnswerRollback() error {
	// проверим, не наш ли это user_id
	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxMaps.Int64["to_user_id"] == myUserId {
		err = p.ExecSql("DELETE FROM "+myPrefix+"my_admin_messages WHERE hex(encrypted) = ? AND type = 'from_admin'", p.TxMaps.Bytes["encrypted_message"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.rollbackAI(myPrefix+"my_admin_messages", 1)
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
		err = p.ExecSql("UPDATE x_my_admin_messages SET status = 'approved' WHERE hex(encrypted) = ? AND status = 'my_pending'", p.TxMaps.Bytes["encrypted_message"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) AdminAnswerRollbackFront() error {
	return nil
}
