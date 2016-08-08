package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeNodeKeyInit() error {

	fields := []map[string]string{{"new_node_public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["new_node_public_key"] = utils.BinToHex(p.TxMaps.Bytes["new_node_public_key"])
	p.TxMap["new_node_public_key"] = utils.BinToHex(p.TxMap["new_node_public_key"])
	return nil
}

func (p *Parser) ChangeNodeKeyFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"new_node_public_key": "public_key"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	nodePublicKey, err := p.GetNodePublicKey(p.TxUserID)
	if err != nil || len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["new_node_public_key"])
	CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.TxMap["sign"], true)
	if err != nil || !CheckSignResult {
		forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["new_node_public_key"])
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
		if err != nil || !CheckSignResult {
			return p.ErrInfo("incorrect sign")
		}
	}

	err = p.limitRequest(p.Variables.Int64["limit_node_key"], "node_key", p.Variables.Int64["limit_node_key_period"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) ChangeNodeKey() error {

	// Всегда есть, что логировать, т.к. это обновление ключа
	logData, err := p.OneRow("SELECT * FROM miners_data WHERE user_id  =  ?", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	nodePublicKeyHex := utils.BinToHex([]byte(logData["node_public_key"]))

	logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_miners_data ( node_public_key, block_id, prev_log_id ) VALUES ( [hex], ?, ? )", "log_id", nodePublicKeyHex, p.BlockData.BlockId, logData["log_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("UPDATE miners_data SET node_public_key = [hex], log_id = ? WHERE user_id = ?", p.TxMaps.Bytes["new_node_public_key"], logId, p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, не наш ли это user_id
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxMaps.Int64["user_id"])
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		// обновим статус в нашей локальной табле.
		err = p.ExecSql("UPDATE "+myPrefix+"my_node_keys SET status = 'approved', block_id = ?, time = ? WHERE hex(public_key) = ? AND status = 'my_pending'", p.BlockData.BlockId, p.BlockData.Time, p.TxMaps.Bytes["new_node_public_key"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) ChangeNodeKeyRollback() error {
	// получим log_id, по которому можно найти данные, которые были до этого
	// $log_id всегда больше нуля, т.к. это откат обновления ключа

	logId, err := p.Single("SELECT log_id FROM miners_data WHERE user_id  =  ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// данные, которые восстановим
	data, err := p.OneRow("SELECT node_public_key, prev_log_id FROM log_miners_data WHERE log_id  =  ?", logId).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	nodePublicKeyHex := utils.BinToHex([]byte(data["node_public_key"]))
	err = p.ExecSql("UPDATE miners_data SET node_public_key =[hex], log_id = ? WHERE user_id = ?", nodePublicKeyHex, data["prev_log_id"], p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// подчищаем _log
	err = p.ExecSql("DELETE FROM log_miners_data WHERE log_id = ?", logId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("log_miners_data", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	// проверим, не наш ли это user_id
	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxMaps.Int64["user_id"])
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId {
		// обновим статус в нашей локальной табле.
		err = p.ExecSql("UPDATE "+myPrefix+"my_node_keys SET status = 'my_pending', block_id = 0, time = 0 WHERE hex(public_key) = ? AND status = 'approved' AND block_id = ?", p.TxMaps.Bytes["new_node_public_key"], p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) ChangeNodeKeyRollbackFront() error {
	return p.limitRequestsRollback("node_key")
}
