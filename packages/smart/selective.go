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

package smart

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/model/querycost"

	log "github.com/sirupsen/logrus"
)

func prepareValues(sc *SmartContract, table string, fields []string,
	ivalues []interface{}) ([]string, error) {

	if !sc.VDE && sc.Rollback && sc.BlockData == nil {
		return nil, logErrorShort(errUndefBlock, consts.EmptyObject)
	}

	for i, v := range ivalues {
		if len(fields) > i && converter.IsByteColumn(table, fields[i]) {
			switch v.(type) {
			case string:
				if vbyte, err := hex.DecodeString(v.(string)); err == nil {
					ivalues[i] = vbyte
				}
			}
		}
	}

	values, err := converter.InterfaceSliceToStr(ivalues)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func addRollback(sc *SmartContract, table, tableID, rollbackInfoStr string) error {
	if !sc.VDE && sc.Rollback {
		rollbackTx := &model.RollbackTx{
			BlockID:   sc.BlockData.BlockID,
			TxHash:    sc.TxHash,
			NameTable: table,
			TableID:   tableID,
			Data:      rollbackInfoStr,
		}
		err := rollbackTx.Create(sc.DbTransaction)
		if err != nil {
			return logErrorDB(err, "creating rollback tx")
		}
	}
	return nil
}

func (sc *SmartContract) getSQLIns(table string, fields []string, values []string) (tableID string,
	addSQLIns0, addSQLIns1 []string, err error) {

	jsonFields := make(map[string]map[string]string)
	isID := false
	addSQLIns0 = []string{}
	addSQLIns1 = []string{}
	for i := 0; i < len(fields); i++ {
		if fields[i] == `id` {
			isID = true
			tableID = escapeSingleQuotes(fmt.Sprint(values[i]))
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
		insField := fields[i]
		if fields[i][:1] == "+" || fields[i][:1] == "-" {
			insField = fields[i][1:len(fields[i])]
		} else if strings.HasPrefix(fields[i], `timestamp `) {
			insField = fields[i][len(`timestamp `):]
		}
		addSQLIns0 = append(addSQLIns0, insField)

		var insVal string
		if converter.IsByteColumn(table, fields[i]) && len(values[i]) != 0 {
			insVal = `decode('` + hex.EncodeToString([]byte(values[i])) + `','HEX')`
		} else if values[i] == `NULL` {
			insVal = `NULL`
		} else if strings.HasPrefix(fields[i], `timestamp`) {
			insVal = `to_timestamp('` + escapeSingleQuotes(values[i]) + `')`
		} else if strings.HasPrefix(values[i], `timestamp`) {
			insVal = `timestamp '` + escapeSingleQuotes(values[i][len(`timestamp `):]) + `'`
		} else {
			insVal = `'` + escapeSingleQuotes(values[i]) + `'`
		}
		addSQLIns1 = append(addSQLIns1, insVal)
	}
	for colname, colvals := range jsonFields {
		var out []byte
		out, err = marshalJSON(colvals, `update columns for jsonb`)
		if err != nil {
			return
		}
		addSQLIns0 = append(addSQLIns0, colname)
		addSQLIns1 = append(addSQLIns1, fmt.Sprintf(`'%s'::jsonb`, string(out)))
	}
	if !isID {
		var id int64
		id, err = model.GetNextID(sc.DbTransaction, table)
		if err != nil {
			logErrorDB(err, "getting next id for table")
			return
		}
		tableID = converter.Int64ToStr(id)
		addSQLIns0 = append(addSQLIns0, `id`)
		addSQLIns1 = append(addSQLIns1, `'`+tableID+`'`)
	}
	return
}

func (sc *SmartContract) insert(fields []string, ivalues []interface{},
	table string) (int64, string, error) {
	var (
		err  error
		cost int64
	)
	queryCoster := querycost.GetQueryCoster(querycost.FormulaQueryCosterType)
	values, err := prepareValues(sc, table, fields, ivalues)
	if err != nil {
		return 0, ``, err
	}
	tableID, addSQLIns0, addSQLIns1, err := sc.getSQLIns(table, fields, values)
	if err != nil {
		return 0, ``, err
	}
	insertQuery := `INSERT INTO "` + table + `" (` + strings.Join(addSQLIns0, ",") +
		`) VALUES (` + strings.Join(addSQLIns1, ",") + `)`
	if !sc.VDE {
		insertCost, err := queryCoster.QueryCost(sc.DbTransaction, insertQuery)
		if err != nil {
			return 0, tableID, logErrorValue(err, consts.DBError,
				"getting total query cost for insert query", insertQuery)
		}
		cost += insertCost
	}
	err = model.GetDB(sc.DbTransaction).Exec(insertQuery).Error
	if err != nil {
		return 0, tableID, logErrorValue(err, consts.DBError, "executing insert query", insertQuery)
	}
	return cost, tableID, addRollback(sc, table, tableID, ``)
}

func (sc *SmartContract) getFieldsAndWhere(fields, whereFields, whereValues []string) (addSQLFields,
	addSQLWhere string) {
	addSQLFields = `id,`
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
			addSQLFields += field + ","
		}
	}
	addSQLFields = strings.TrimRight(addSQLFields, ",")

	for i := 0; i < len(whereFields); i++ {
		if val := converter.StrToInt64(whereValues[i]); val != 0 {
			addSQLWhere += whereFields[i] + "= " + escapeSingleQuotes(whereValues[i]) + " AND "
		} else {
			addSQLWhere += whereFields[i] + "= '" + escapeSingleQuotes(whereValues[i]) + "' AND "
		}
	}
	if len(addSQLWhere) > 0 {
		addSQLWhere = " WHERE " + addSQLWhere[0:len(addSQLWhere)-5]
	}
	return
}

func getSQLUpdate(table string, fields, values []string, logData map[string]string) (
	addSQLUpdate string, err error) {
	jsonFields := make(map[string]map[string]string)

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
			addSQLUpdate += fields[i] + `='` + escapeSingleQuotes(values[i]) + `',`
		}
	}
	for colname, colvals := range jsonFields {
		var (
			initial string
			out     []byte
		)
		out, err = marshalJSON(colvals, `update columns for jsonb`)
		if err != nil {
			return
		}
		if len(logData[colname]) > 0 && logData[colname] != `NULL` {
			initial = colname
		} else {
			initial = `'{}'`
		}
		addSQLUpdate += fmt.Sprintf(`%s=%s::jsonb || '%s'::jsonb,`, colname, initial, string(out))
	}
	addSQLUpdate = strings.TrimRight(addSQLUpdate, `,`)
	return
}

func (sc *SmartContract) update(fields []string, ivalues []interface{},
	table string, whereFields, whereValues []string) (int64, string, error) {
	var (
		tableID         string
		err             error
		cost            int64
		rollbackInfoStr string
	)
	queryCoster := querycost.GetQueryCoster(querycost.FormulaQueryCosterType)
	logger := sc.GetLogger()

	values, err := prepareValues(sc, table, fields, ivalues)
	if err != nil {
		return 0, ``, err
	}
	if whereFields == nil || whereValues == nil {
		return 0, ``, logErrorShort(errUpdNotExistRecord, consts.NotFound)
	}
	addSQLFields, addSQLWhere := sc.getFieldsAndWhere(fields, whereFields, whereValues)
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
	if len(logData) == 0 {
		logger.WithFields(log.Fields{"type": consts.NotFound, "err": errUpdNotExistRecord, "query": selectQuery}).Error("updating for not existing record")
		return 0, tableID, errUpdNotExistRecord
	}
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
	jsonRollbackInfo, err := marshalJSON(rollbackInfo, `rollback info to json`)
	if err != nil {
		return 0, tableID, err
	}
	rollbackInfoStr = string(jsonRollbackInfo)
	addSQLUpdate, err := getSQLUpdate(table, fields, values, logData)
	if err != nil {
		return 0, ``, err
	}

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

	return cost, tableID, addRollback(sc, table, tableID, rollbackInfoStr)
}

func escapeSingleQuotes(val string) string {
	return strings.Replace(val, `'`, `''`, -1)
}
