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
package smart

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/model/querycost"

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

	for i, v := range ivalues {
		switch v.(type) {
		case string:
			if strings.HasPrefix(strings.TrimSpace(v.(string)), `timestamp`) {
				if err = checkNow(v.(string)); err != nil {
					return 0, ``, err
				}
			}
			if len(fields) > i && converter.IsByteColumn(table, fields[i]) {
				if vbyte, err := hex.DecodeString(v.(string)); err == nil {
					ivalues[i] = vbyte
				}
			}
		}
	}

	values, err := converter.InterfaceSliceToStr(ivalues)
	if err != nil {
		return 0, ``, err
	}

	addSQLFields := `id,`
	for i, field := range fields {
		field = strings.TrimSpace(strings.ToLower(field))
		fields[i] = field
		if field[:1] == "+" || field[:1] == "-" {
			addSQLFields += field[1:] + ","
		} else if strings.HasPrefix(field, `timestamp `) {
			addSQLFields += field[len(`timestamp `):] + `,`
		} else if strings.Contains(field, `->`) {
			addSQLFields += field[:strings.Index(field, `->`)] + `,`
		} else {
			addSQLFields += `"` + field + `",`
		}
	}

	addSQLWhere := ""
	if whereFields != nil && whereValues != nil {
		for i := 0; i < len(whereFields); i++ {
			if val := converter.StrToInt64(whereValues[i]); val != 0 {
				addSQLWhere += whereFields[i] + "= " + escapeSingleQuotes(whereValues[i]) + " AND "
			} else {
				addSQLWhere += whereFields[i] + "= '" + escapeSingleQuotes(whereValues[i]) + "' AND "
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
	jsonFields := make(map[string]map[string]string)
	if whereFields != nil && len(logData) > 0 {
		rollbackInfo := make(map[string]string)
		for k, v := range logData {
			if k == `id` {
				continue
			}
			if converter.IsByteColumn(table, k) && v != "" {
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
			if converter.IsByteColumn(table, fields[i]) && len(values[i]) != 0 {
				addSQLUpdate += fields[i] + `=decode('` + hex.EncodeToString([]byte(values[i])) + `','HEX'),`
			} else if fields[i][:1] == "+" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `+` + escapeSingleQuotes(values[i]) + `,`
			} else if fields[i][:1] == "-" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `-` + escapeSingleQuotes(values[i]) + `,`
			} else if values[i] == `NULL` {
				addSQLUpdate += fields[i] + `= NULL,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLUpdate += fields[i][len(`timestamp `):] + `= to_timestamp('` + values[i] + `'),`
			} else if strings.HasPrefix(values[i], `timestamp `) {
				addSQLUpdate += fields[i] + `= timestamp '` + escapeSingleQuotes(values[i][len(`timestamp `):]) + `',`
			} else if strings.Contains(fields[i], `->`) {
				colfield := strings.Split(fields[i], `->`)
				if len(colfield) == 2 {
					if jsonFields[colfield[0]] == nil {
						jsonFields[colfield[0]] = make(map[string]string)
					}
					jsonFields[colfield[0]][colfield[1]] = values[i]
				}
			} else {
				addSQLUpdate += `"` + fields[i] + `"='` + escapeSingleQuotes(values[i]) + `',`
			}
		}
		for colname, colvals := range jsonFields {
			var initial string
			out, err := json.Marshal(colvals)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "type": consts.JSONMarshallError}).Error("marshalling update columns for jsonb")
				return 0, ``, err
			}
			if len(logData[colname]) > 0 && logData[colname] != `NULL` {
				initial = colname
			} else {
				initial = `'{}'`
			}
			addSQLUpdate += fmt.Sprintf(`%s=%s::jsonb || '%s'::jsonb,`, colname, initial, string(out))
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
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "sql": addSQLUpdate}).Error("getting update query")
			return 0, tableID, err
		}
		tableID = logData[`id`]
	} else {
		isID := false
		addSQLIns0 := []string{}
		addSQLIns1 := []string{}
		for i := 0; i < len(fields); i++ {
			if fields[i] == `id` {
				isID = true
				tableID = escapeSingleQuotes(values[i])
			}

			if strings.Contains(fields[i], `->`) {
				colfield := strings.Split(fields[i], `->`)
				if len(colfield) == 2 {
					if jsonFields[colfield[0]] == nil {
						jsonFields[colfield[0]] = make(map[string]string)
					}
					jsonFields[colfield[0]][colfield[1]] = escapeSingleQuotes(values[i])
					continue
				}
			}

			if fields[i][:1] == "+" || fields[i][:1] == "-" {
				addSQLIns0 = append(addSQLIns0, fields[i][1:len(fields[i])])
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLIns0 = append(addSQLIns0, fields[i][len(`timestamp `):])
			} else {
				addSQLIns0 = append(addSQLIns0, `"`+fields[i]+`"`)
			}
			if converter.IsByteColumn(table, fields[i]) && len(values[i]) != 0 {
				addSQLIns1 = append(addSQLIns1, `decode('`+hex.EncodeToString([]byte(values[i]))+`','HEX')`)
			} else if values[i] == `NULL` {
				addSQLIns1 = append(addSQLIns1, `NULL`)
			} else if strings.HasPrefix(fields[i], `timestamp`) {
				addSQLIns1 = append(addSQLIns1, `to_timestamp('`+escapeSingleQuotes(values[i])+`')`)
			} else if strings.HasPrefix(values[i], `timestamp`) {
				addSQLIns1 = append(addSQLIns1, `timestamp '`+escapeSingleQuotes(values[i][len(`timestamp `):])+`'`)
			} else {
				addSQLIns1 = append(addSQLIns1, `'`+escapeSingleQuotes(values[i])+`'`)
			}
		}
		for colname, colvals := range jsonFields {
			out, err := json.Marshal(colvals)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "type": consts.JSONMarshallError}).Error("marshalling update columns for jsonb")
				return 0, ``, err
			}
			addSQLIns0 = append(addSQLIns0, colname)
			addSQLIns1 = append(addSQLIns1, fmt.Sprintf(`'%s'::jsonb`, string(out)))
		}
		if whereFields != nil && whereValues != nil {
			for i := 0; i < len(whereFields); i++ {
				if whereFields[i] == `id` {
					isID = true
					tableID = whereValues[i]
				}
				addSQLIns0 = append(addSQLIns0, whereFields[i])
				addSQLIns1 = append(addSQLIns1, escapeSingleQuotes(whereValues[i]))
			}
		}
		if !isID {
			id, err := model.GetNextID(sc.DbTransaction, table)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id for table")
				return 0, ``, err
			}
			tableID = converter.Int64ToStr(id)
			addSQLIns0 = append(addSQLIns0, `id`)
			addSQLIns1 = append(addSQLIns1, `'`+tableID+`'`)
		}
		insertQuery := `INSERT INTO "` + table + `" (` + strings.Join(addSQLIns0, ",") +
			`) VALUES (` + strings.Join(addSQLIns1, ",") + `)`
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

func escapeSingleQuotes(val string) string {
	return strings.Replace(val, `'`, `''`, -1)
}
