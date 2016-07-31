package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"sort"
	"strings"
)

func (p *Parser) ChangeArbitratorListInit() error {

	fields := []map[string]string{{"arbitration_trust_list": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeArbitratorListFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"arbitration_trust_list": "arbitration_trust_list"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.TxMaps.String["arbitration_trust_list"]) > 255 {
		return p.ErrInfo("len arbitration_trust_list > 255")
	}
	var arbitrationTrustList []int
	if p.TxMaps.String["arbitration_trust_list"] != "[0]" {
		err = json.Unmarshal(p.TxMap["arbitration_trust_list"], &arbitrationTrustList)
		if err != nil {
			return p.ErrInfo(err)
		}
		sort.Ints(arbitrationTrustList)
	}
	// юзер мог удалить весь список доверенных
	if len(arbitrationTrustList) > 0 {
		// указанные id должны быть ID юзеров. Являются ли эти юзеры арбитрами будет проверяться при отправке монет
		count, err := p.Single("SELECT count(user_id) FROM users WHERE user_id IN (" + strings.Join(utils.IntSliceToStr(arbitrationTrustList), ",") + ")").Int()
		if err != nil {
			return p.ErrInfo(err)
		}
		if count != len(arbitrationTrustList) {
			return p.ErrInfo("count != len(arbitrationTrustList)")
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["arbitration_trust_list"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CHANGE_ARBITRATION_TRUST_LIST, "change_arbitration_trust_list", consts.LIMIT_CHANGE_ARBITRATION_TRUST_LIST_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeArbitratorList() error {

	logArbitrationTrustList, err := p.GetList("SELECT arbitrator_user_id FROM arbitration_trust_list WHERE user_id  =  ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	var logId int64

	// логируем текущие значения, если они есть
	if len(logArbitrationTrustList) > 0 {

		logArbitrationTrustListJson, err := json.Marshal(logArbitrationTrustList)
		if err != nil {
			return p.ErrInfo(err)
		}

		prevLogId, err := p.Single("SELECT log_id FROM arbitration_trust_list WHERE user_id  =  ?", p.TxUserID).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		// логируем текущие значения, если они есть
		logId, err = p.ExecSqlGetLastInsertId("INSERT INTO log_arbitration_trust_list ( arbitration_trust_list, prev_log_id ) VALUES ( ?, ? )", "log_id", logArbitrationTrustListJson, prevLogId)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("DELETE FROM arbitration_trust_list WHERE user_id = ?", p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		logId = 0
	}

	var arbitrationTrustList []int
	if p.TxMaps.String["arbitration_trust_list"] != "[0]" {
		err = json.Unmarshal(p.TxMap["arbitration_trust_list"], &arbitrationTrustList)
		if err != nil {
			return p.ErrInfo(err)
		}
		for i := 0; i < len(arbitrationTrustList); i++ {
			err = p.ExecSql("INSERT INTO arbitration_trust_list ( user_id, arbitrator_user_id, log_id ) VALUES ( ?, ?, ? )", p.TxUserID, arbitrationTrustList[i], logId)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	} else {
		err = p.ExecSql("INSERT INTO arbitration_trust_list ( user_id, arbitrator_user_id, log_id ) VALUES ( ?, ?, ? )", p.TxUserID, 0, logId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) ChangeArbitratorListRollback() error {

	logId, err := p.Single("SELECT log_id FROM arbitration_trust_list WHERE user_id  =  ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("DELETE FROM arbitration_trust_list WHERE user_id = ?", p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if logId > 0 {
		logData, err := p.OneRow("SELECT prev_log_id, arbitration_trust_list FROM log_arbitration_trust_list WHERE log_id  =  ?", logId).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		// подчищаем _log
		err = p.ExecSql("DELETE FROM log_arbitration_trust_list WHERE log_id = ?", logId)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.rollbackAI("log_arbitration_trust_list", 1)
		if err != nil {
			return p.ErrInfo(err)
		}

		var arbitrationTrustList []int
		err = json.Unmarshal([]byte(logData["arbitration_trust_list"]), &arbitrationTrustList)
		if err != nil {
			return p.ErrInfo(err)
		}
		for i := 0; i < len(arbitrationTrustList); i++ {
			err = p.ExecSql("INSERT INTO arbitration_trust_list ( user_id, arbitrator_user_id, log_id ) VALUES ( ?, ?, ? )", p.TxUserID, arbitrationTrustList[i], logData["prev_log_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	return nil
}

func (p *Parser) ChangeArbitratorListRollbackFront() error {
	return p.limitRequestsRollback("change_arbitration_trust_list")
}
