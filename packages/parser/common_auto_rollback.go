//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package parser

import (
	"encoding/json"
	"strings"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

func (p *Parser) restoreUpdatedDBRowToPreviousData(tx map[string]string, where string) error {
	logger := p.GetLogger()
	var rollbackInfo map[string]string
	if err := json.Unmarshal([]byte(tx["data"]), &rollbackInfo); err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollback.Data from json")
		return p.ErrInfo(err)
	}
	addSQLUpdate := ""
	for k, v := range rollbackInfo {
		if converter.InSliceString(k, []string{"hash", "pub", "tx_hash", "public_key_0", "node_public_key"}) && len(v) != 0 {
			addSQLUpdate += k + `=decode('` + string(converter.BinToHex([]byte(v))) + `','HEX'),`
		} else {
			addSQLUpdate += k + `='` + strings.Replace(v, `'`, `''`, -1) + `',`
		}
	}
	addSQLUpdate = addSQLUpdate[0 : len(addSQLUpdate)-1]
	if err := model.Update(p.DbTransaction, tx["table_name"], addSQLUpdate, where); err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "query": addSQLUpdate}).Error("updating table")
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) deleteInsertedDBRow(tx map[string]string, where string) error {
	logger := p.GetLogger()
	if err := model.Delete(tx["table_name"], where); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting from table")
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) autoRollback() error {
	logger := p.GetLogger()
	rollbackTx := &model.RollbackTx{}
	txs, err := rollbackTx.GetRollbackTransactions(p.DbTransaction, p.TxHash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback transactions")
		return utils.ErrInfo(err)
	}
	for _, tx := range txs {
		where := " WHERE id='" + tx["table_id"] + `'`
		if len(tx["data"]) > 0 {
			if err := p.restoreUpdatedDBRowToPreviousData(tx, where); err != nil {
				return err
			}
		} else {
			if err := p.deleteInsertedDBRow(tx, where); err != nil {
				return err
			}
		}
	}
	txForDelete := &model.RollbackTx{TxHash: p.TxHash}
	err = txForDelete.DeleteByHash(p.DbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting rollback transaction by hash")
		return p.ErrInfo(err)
	}
	return nil
}
