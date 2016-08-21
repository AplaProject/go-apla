package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) RollbackTransactions() error {

	var blockBody []byte

	utils.WriteSelectiveLog("SELECT data, hash FROM transactions WHERE verified = 1 AND used = 0")
	rows, err := p.Query("SELECT data, hash FROM transactions WHERE verified = 1 AND used = 0")
	if err != nil {
		utils.WriteSelectiveLog(err)
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var data, hash []byte
		err = rows.Scan(&data, &hash)
		if err != nil {
			utils.WriteSelectiveLog(err)
			return p.ErrInfo(err)
		}
		utils.WriteSelectiveLog(utils.BinToHex(hash))
		blockBody = append(blockBody, utils.EncodeLengthPlusData(data)...)
		utils.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE hex(hash) = " + string(utils.BinToHex(hash)))
		affect, err := p.ExecSqlGetAffect("UPDATE transactions SET verified = 0 WHERE hex(hash) = ?", utils.BinToHex(hash))
		if err != nil {
			utils.WriteSelectiveLog(err)
			return p.ErrInfo(err)
		}
		utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
	}

	// нужно откатить наши транзакции
	if len(blockBody) > 0 {
		parser := new(Parser)
		parser.DCDB = p.DCDB
		parser.BinaryData = blockBody
		err = parser.ParseDataRollbackFront(false)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

