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

package rollback

import (
	"encoding/json"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	log "github.com/sirupsen/logrus"
)

func rollbackUpdatedRow(tx map[string]string, where string, dbTransaction *model.DbTransaction, logger *log.Entry) error {
	var rollbackInfo map[string]string
	if err := json.Unmarshal([]byte(tx["data"]), &rollbackInfo); err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollback.Data from json")
		return err
	}
	addSQLUpdate := ""
	for k, v := range rollbackInfo {
		if v == "NULL" {
			addSQLUpdate += k + `=NULL,`
		} else if converter.IsByteColumn(tx["table_name"], k) && len(v) != 0 {
			addSQLUpdate += k + `=decode('` + string(converter.BinToHex([]byte(v))) + `','HEX'),`
		} else {
			addSQLUpdate += k + `='` + strings.Replace(v, `'`, `''`, -1) + `',`
		}
	}
	addSQLUpdate = addSQLUpdate[0 : len(addSQLUpdate)-1]
	if err := model.Update(dbTransaction, tx["table_name"], addSQLUpdate, where); err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "query": addSQLUpdate}).Error("updating table")
		return err
	}
	return nil
}

func rollbackInsertedRow(tx map[string]string, where string, dbTransaction *model.DbTransaction, logger *log.Entry) error {
	if err := model.Delete(dbTransaction, tx["table_name"], where); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting from table")
		return err
	}
	return nil
}

func rollbackTransaction(txHash []byte, dbTransaction *model.DbTransaction, logger *log.Entry) error {
	rollbackTx := &model.RollbackTx{}
	txs, err := rollbackTx.GetRollbackTransactions(dbTransaction, txHash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback transactions")
		return err
	}
	for _, tx := range txs {
		if tx["table_name"] == smart.SysName {
			var sysData smart.SysRollData
			err := json.Unmarshal([]byte(tx["data"]), &sysData)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollback.Data from json")
				return err
			}
			switch sysData.Type {
			case "NewTable":
				smart.SysRollbackTable(dbTransaction, txHash, sysData, tx["table_id"])
			case "NewColumn":
				smart.SysRollbackColumn(dbTransaction, sysData, tx["table_id"])
			case "NewContract":
				smart.SysRollbackNewContract(sysData, tx["table_id"])
			case "EditContract":
				smart.SysRollbackEditContract(dbTransaction, txHash, tx["table_id"])
			case "NewEcosystem":
				smart.SysRollbackEcosystem(dbTransaction, txHash)
			case "ActivateContract":
				smart.SysRollbackActivate(sysData)
			case "DeactivateContract":
				smart.SysRollbackDeactivate(sysData)
			case "DeleteColumn":
				smart.SysRollbackDeleteColumn(dbTransaction, sysData)
			}
			continue
		}
		where := " WHERE id='" + tx["table_id"] + `'`
		if len(tx["data"]) > 0 {
			if err := rollbackUpdatedRow(tx, where, dbTransaction, logger); err != nil {
				return err
			}
		} else {
			if err := rollbackInsertedRow(tx, where, dbTransaction, logger); err != nil {
				return err
			}
		}
	}
	txForDelete := &model.RollbackTx{TxHash: txHash}
	err = txForDelete.DeleteByHash(dbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting rollback transaction by hash")
		return err
	}
	return nil
}
