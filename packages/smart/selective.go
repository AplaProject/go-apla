// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package smart

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/model/querycost"

	log "github.com/sirupsen/logrus"
)

var (
	errUpdNotExistRecord = errors.New(`Update for not existing record`)
)

func (sc *SmartContract) selectiveLoggingAndUpd(fields []string, ivalues []interface{},
	table string, whereFields, whereValues []string, generalRollback bool, exists bool) (int64, string, error) {
	queryCoster := querycost.GetQueryCoster(querycost.FormulaQueryCosterType)
	var (
		tableID         string
		err             error
		cost            int64
		rollbackInfoStr string
	)
	logger := sc.GetLogger()

	if generalRollback && sc.BlockData == nil {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("Block is undefined")
		return 0, ``, fmt.Errorf(`It is impossible to write to DB when Block is undefined`)
	}

	isBytea := GetBytea(sc.DbTransaction, table)
	for i, v := range ivalues {
		if len(fields) > i && isBytea[fields[i]] {
			switch v.(type) {
			case string:
				if vbyte, err := hex.DecodeString(v.(string)); err == nil {
					ivalues[i] = vbyte
				}
			}
		}
	}

	values := converter.InterfaceSliceToStr(ivalues)

	addSQLFields := `id,`
	for i, field := range fields {
		field = strings.TrimSpace(field)
		fields[i] = field
		if field[:1] == "+" || field[:1] == "-" {
			addSQLFields += field[1:] + ","
		} else if strings.HasPrefix(field, `timestamp `) {
			addSQLFields += field[len(`timestamp `):] + `,`
		} else {
			addSQLFields += field + ","
		}
	}

	addSQLWhere := ""
	if whereFields != nil && whereValues != nil {
		for i := 0; i < len(whereFields); i++ {
			if val := converter.StrToInt64(whereValues[i]); val != 0 {
				addSQLWhere += whereFields[i] + "= " + whereValues[i] + " AND "
			} else {
				addSQLWhere += whereFields[i] + "= '" + whereValues[i] + "' AND "
			}
		}
	}
	if len(addSQLWhere) > 0 {
		addSQLWhere = " WHERE " + addSQLWhere[0:len(addSQLWhere)-5]
	}
	addSQLFields = strings.TrimRight(addSQLFields, ",")
	selectQuery := `SELECT ` + addSQLFields + ` FROM "` + table + `" ` + addSQLWhere
	selectCost, err := queryCoster.QueryCost(sc.DbTransaction, selectQuery)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": selectQuery}).Error("getting query total cost")
		return 0, tableID, err
	}
	logData, err := model.GetOneRowTransaction(sc.DbTransaction, selectQuery).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": selectQuery}).Error("getting one row transaction")
		return 0, tableID, err
	}
	cost += selectCost
	if exists && len(logData) == 0 {
		logger.WithFields(log.Fields{"type": consts.NotFound, "err": errUpdNotExistRecord, "query": selectQuery}).Error("updating for not existing record")
		return 0, tableID, errUpdNotExistRecord
	}
	if whereFields != nil && len(logData) > 0 {
		rollbackInfo := make(map[string]string)
		for k, v := range logData {
			if k == `id` {
				continue
			}
			if (isBytea[k] || converter.InSliceString(k, []string{"hash", "tx_hash", "pub", "tx_hash", "public_key_0", "node_public_key"})) && v != "" {
				rollbackInfo[k] = string(converter.BinToHex([]byte(v)))
			} else {
				rollbackInfo[k] = v
			}
			if k[:1] == "+" || k[:1] == "-" {
				addSQLFields += k[1:] + ","
			} else if strings.HasPrefix(k, `timestamp `) {
				addSQLFields += k[len(`timestamp `):] + `,`
			} else {
				addSQLFields += k + ","
			}
		}
		jsonRollbackInfo, err := json.Marshal(rollbackInfo)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling rollback info to json")
			return 0, tableID, err
		}
		rollbackInfoStr = string(jsonRollbackInfo)

		addSQLUpdate := ""
		for i := 0; i < len(fields); i++ {
			if isBytea[fields[i]] && len(values[i]) != 0 {
				addSQLUpdate += fields[i] + `=decode('` + hex.EncodeToString([]byte(values[i])) + `','HEX'),`
			} else if fields[i][:1] == "+" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `+` + values[i] + `,`
			} else if fields[i][:1] == "-" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `-` + values[i] + `,`
			} else if values[i] == `NULL` {
				addSQLUpdate += fields[i] + `= NULL,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLUpdate += fields[i][len(`timestamp `):] + `= to_timestamp('` + values[i] + `'),`
			} else if strings.HasPrefix(values[i], `timestamp `) {
				addSQLUpdate += fields[i] + `= timestamp '` + values[i][len(`timestamp `):] + `',`
			} else {
				addSQLUpdate += fields[i] + `='` + strings.Replace(values[i], `'`, `''`, -1) + `',`
			}
		}
		addSQLUpdate = strings.TrimRight(addSQLUpdate, `,`)
		if !sc.VDE {
			updateQuery := `UPDATE "` + table + `" SET ` + addSQLUpdate + addSQLWhere
			updateCost, err := queryCoster.QueryCost(sc.DbTransaction, updateQuery)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": updateQuery}).Error("getting query total cost for update query")
				return 0, tableID, err
			}
			cost += updateCost
		}
		err = model.Update(sc.DbTransaction, table, addSQLUpdate, addSQLWhere)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting update query")
			return 0, tableID, err
		}
		tableID = logData[`id`]
	} else {
		isID := false
		addSQLIns0 := ""
		addSQLIns1 := ""
		for i := 0; i < len(fields); i++ {
			if fields[i] == `id` {
				isID = true
				tableID = fmt.Sprint(values[i])
			}
			if fields[i][:1] == "+" || fields[i][:1] == "-" {
				addSQLIns0 += fields[i][1:len(fields[i])] + `,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLIns0 += fields[i][len(`timestamp `):] + `,`
			} else {
				addSQLIns0 += fields[i] + `,`
			}
			if isBytea[fields[i]] && len(values[i]) != 0 {
				addSQLIns1 += `decode('` + hex.EncodeToString([]byte(values[i])) + `','HEX'),`
			} else if values[i] == `NULL` {
				addSQLIns1 += `NULL,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLIns1 += `to_timestamp('` + values[i] + `'),`
			} else if strings.HasPrefix(values[i], `timestamp `) {
				addSQLIns1 += `timestamp '` + values[i][len(`timestamp `):] + `',`
			} else {
				addSQLIns1 += `'` + strings.Replace(values[i], `'`, `''`, -1) + `',`
			}
		}
		if whereFields != nil && whereValues != nil {
			for i := 0; i < len(whereFields); i++ {
				if whereFields[i] == `id` {
					isID = true
					tableID = fmt.Sprint(whereValues[i])
				}
				addSQLIns0 += `` + whereFields[i] + `,`
				addSQLIns1 += `'` + whereValues[i] + `',`
			}
		}
		if !isID {
			id, err := model.GetNextID(sc.DbTransaction, table)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id for table")
				return 0, ``, err
			}
			tableID = converter.Int64ToStr(id)
			addSQLIns0 += `id,`
			addSQLIns1 += `'` + tableID + `',`
		}
		insertQuery := `INSERT INTO "` + table + `" (` + addSQLIns0[:len(addSQLIns0)-1] +
			`) VALUES (` + addSQLIns1[:len(addSQLIns1)-1] + `)`
		insertCost, err := queryCoster.QueryCost(sc.DbTransaction, insertQuery)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": insertQuery}).Error("getting total query cost for insert query")
			return 0, tableID, err
		}
		cost += insertCost
		err = model.GetDB(sc.DbTransaction).Exec(insertQuery).Error
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": insertQuery}).Error("executing insert query")
		}
	}
	if err != nil {
		return 0, tableID, err
	}

	if generalRollback {
		rollbackTx := &model.RollbackTx{
			BlockID:   sc.BlockData.BlockID,
			TxHash:    sc.TxHash,
			NameTable: table,
			TableID:   tableID,
			Data:      rollbackInfoStr,
		}

		err = rollbackTx.Create(sc.DbTransaction)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating rollback tx")
			return 0, tableID, err
		}
	}
	return cost, tableID, nil
}
