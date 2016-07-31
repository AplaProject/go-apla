package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
)

func (p *Parser) ChangeHostInit() error {
	var fields []map[string]string

	if p.BlockData != nil && p.BlockData.BlockId < 250900 {
		fields = []map[string]string{{"http_host": "string"}, {"sign": "bytes"}}
	} else if p.BlockData != nil && p.BlockData.BlockId < 261209 {
		fields = []map[string]string{{"http_host": "string"}, {"tcp_host": "string"}, {"sign": "bytes"}}
	} else {
		fields = []map[string]string{{"http_host": "string"}, {"tcp_host": "string"}, {"e_host": "string"}, {"sign": "bytes"}}
	}

	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeHostFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}
	var verifyData map[string]string
	if p.BlockData != nil && p.BlockData.BlockId < 250900 {
		verifyData = map[string]string{"http_host": "http_host"}
	} else if p.BlockData != nil && p.BlockData.BlockId < 261209 {
		verifyData = map[string]string{"http_host": "http_host", "tcp_host": "tcp_host"}
	} else {
		verifyData = map[string]string{"http_host": "http_host", "tcp_host": "tcp_host", "e_host": "e_host"}
	}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.BlockData == nil || p.BlockData.BlockId > 281500 {
		// проверим, не занял ли кто-то хосты
		exists, err := p.Single(`SELECT user_id FROM miners_data WHERE http_host = ? OR tcp_host = ? OR e_host = ?`, p.TxMaps.String["http_host"], p.TxMaps.String["tcp_host"], p.TxMaps.String["e_host"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if exists > 0 && exists != p.TxUserID {
			return p.ErrInfo("host exists")
		}
	}

	// нодовский ключ
	nodePublicKey, err := p.GetNodePublicKey(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}

	var CheckSignResult bool
	var forSign string
	if p.BlockData != nil && p.BlockData.BlockId < 250900 {
		forSign = fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["http_host"])
	} else if p.BlockData != nil && p.BlockData.BlockId < 261209 {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["http_host"], p.TxMap["tcp_host"])
	} else {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["http_host"], p.TxMap["tcp_host"], p.TxMap["e_host"])
	}

	if p.BlockData != nil && p.BlockData.BlockId <= 240240 {
		CheckSignResult, err = utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	} else {
		CheckSignResult, err = utils.CheckSign([][]byte{nodePublicKey}, forSign, p.TxMap["sign"], true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_change_host"], "change_host", p.Variables.Int64["limit_change_host_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeHost() error {
	tcpHost := ""
	if p.BlockData != nil && p.BlockData.BlockId < 250900 {
		re := regexp.MustCompile(`^https?:\/\/([0-9a-z\_\.\-:]+)\/`)
		match := re.FindStringSubmatch(p.TxMaps.String["http_host"])
		if len(match) != 0 {
			tcpHost = match[1] + ":8088"
		}
	} else {
		tcpHost = p.TxMaps.String["tcp_host"]
	}
	err := p.selectiveLoggingAndUpd([]string{"http_host", "tcp_host", "e_host"}, []interface{}{p.TxMaps.String["http_host"], tcpHost, p.TxMaps.String["e_host"]}, "miners_data", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
	if err != nil {
		return p.ErrInfo(err)
	}
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		// обновим статус в нашей локальной табле.
		err = p.ExecSql("UPDATE " + myPrefix + "my_table SET host_status = 'approved'")
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) ChangeHostRollback() error {
	err := p.selectiveRollback([]string{"http_host", "tcp_host", "e_host"}, "miners_data", "user_id="+utils.Int64ToStr(p.TxUserID), false)
	if err != nil {
		return p.ErrInfo(err)
	}
	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxUserID == myUserId {
		// обновим статус в нашей локальной табле.
		err = p.ExecSql("UPDATE " + myPrefix + "my_table SET host_status = 'my_pending'")
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) ChangeHostRollbackFront() error {
	return p.limitRequestsRollback("change_host")
}
