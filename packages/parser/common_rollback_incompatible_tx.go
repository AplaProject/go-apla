package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) RollbackIncompatibleTx(typesArr []string) error {

	var whereType string
	for _, txType := range typesArr {
		whereType += utils.Int64ToStr(utils.TypeInt(txType)) + ","
	}
	whereType = whereType[:len(whereType)-1]

	utils.WriteSelectiveLog(`SELECT data FROM transactions WHERE type IN (` + whereType + `) AND verified=1 AND used = 0`)
	transactions, err := p.GetList(`SELECT data FROM transactions WHERE type IN (` + whereType + `) AND verified=1 AND used = 0`).String()
	if err != nil {
		utils.WriteSelectiveLog(err)
		return utils.ErrInfo(err)
	}
	for _, txData := range transactions {

		md5 := utils.Md5(txData)
		utils.WriteSelectiveLog("md5: " + string(md5))
		// откатим фронтальные записи
		p.BinaryData = utils.EncodeLengthPlusData([]byte(txData))
		err = p.ParseDataRollback()
		if err != nil {
			return utils.ErrInfo(err)
		}
		// Удаляем уже записанные тр-ии.

		utils.WriteSelectiveLog("DELETE FROM transactions WHERE hex(hash) = " + string(md5))
		affect, err := p.ExecSqlGetAffect("DELETE FROM transactions WHERE hex(hash) = ?", md5)
		if err != nil {
			utils.WriteSelectiveLog(err)
			return utils.ErrInfo(err)
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))


		// создаем тр-ию, которую потом заново проверим
		err = p.ExecSql("DELETE FROM queue_tx  WHERE hex(hash) = ?", md5)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = p.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, utils.BinToHex([]byte(txData)))
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

