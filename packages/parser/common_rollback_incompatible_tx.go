// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

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
		log.Debug("DELETE FROM queue_tx  WHERE hex(hash) = %s", md5)
		err = p.ExecSql("DELETE FROM queue_tx  WHERE hex(hash) = ?", md5)
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("INSERT INTO queue_tx (hash, data) VALUES (%s, %s)", md5, utils.BinToHex([]byte(txData)))
		err = p.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, utils.BinToHex([]byte(txData)))
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}

