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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/model/querycost"

	log "github.com/sirupsen/logrus"
)

func (sc *SmartContract) insert(fields []string, ivalues []interface{},
	table string) (int64, string, error) {

	queryCoster := querycost.GetQueryCoster(querycost.FormulaQueryCosterType)
	var (
		tableID         string
		err             error
		cost            int64
		rollbackInfoStr string
	)
	logger := sc.GetLogger()
	generalRollback := !sc.VDE && sc.Rollback

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
			addSQLFields += field + ","
		}
	}
	addSQLFields = strings.TrimRight(addSQLFields, ",")
	jsonFields := make(map[string]map[string]string)

	isID := false
	addSQLIns0 := []string{}
	addSQLIns1 := []string{}
	for i := 0; i < len(fields); i++ {
		if fields[i] == `id` {
			isID = true
			tableID = fmt.Sprint(values[i])
		}

		if strings.Contains(fields[i], `->`) {
			colfield := strings.Split(fields[i], `->`)
			if len(colfield) == 2 {
				if jsonFields[colfield[0]] == nil {
					jsonFields[colfield[0]] = make(map[string]string)
				}
				jsonFields[colfield[0]][colfield[1]] = values[i]
				continue
			}
		}

		if fields[i][:1] == "+" || fields[i][:1] == "-" {
			addSQLIns0 = append(addSQLIns0, fields[i][1:len(fields[i])])
		} else if strings.HasPrefix(fields[i], `timestamp `) {
			addSQLIns0 = append(addSQLIns0, fields[i][len(`timestamp `):])
		} else {
			addSQLIns0 = append(addSQLIns0, fields[i])
		}
		if isBytea[fields[i]] && len(values[i]) != 0 {
			addSQLIns1 = append(addSQLIns1, `decode('`+hex.EncodeToString([]byte(values[i]))+`','HEX')`)
		} else if values[i] == `NULL` {
			addSQLIns1 = append(addSQLIns1, `NULL`)
		} else if strings.HasPrefix(fields[i], `timestamp`) {
			addSQLIns1 = append(addSQLIns1, `to_timestamp('`+values[i]+`')`)
		} else if strings.HasPrefix(values[i], `timestamp`) {
			addSQLIns1 = append(addSQLIns1, `timestamp '`+values[i][len(`timestamp `):]+`'`)
		} else {
			addSQLIns1 = append(addSQLIns1, `'`+strings.Replace(values[i], `'`, `''`, -1)+`'`)
		}
	}
	for colname, colvals := range jsonFields {
		out, err := json.Marshal(colvals)
		if err != nil {
			return 0, ``, logError(err, consts.JSONMarshallError, "marshalling update columns for jsonb")
		}
		addSQLIns0 = append(addSQLIns0, colname)
		addSQLIns1 = append(addSQLIns1, fmt.Sprintf(`'%s'::jsonb`, string(out)))
	}
	if !isID {
		id, err := model.GetNextID(sc.DbTransaction, table)
		if err != nil {
			return 0, ``, logErrorDB(err, "getting next id for table")
		}
		tableID = converter.Int64ToStr(id)
		addSQLIns0 = append(addSQLIns0, `id`)
		addSQLIns1 = append(addSQLIns1, `'`+tableID+`'`)
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
			return 0, tableID, logErrorDB(err, "creating rollback tx")
		}
	}
	return cost, tableID, nil

}

func (sc *SmartContract) update(fields []string, ivalues []interface{},
	table string, whereFields, whereValues []string) (int64, string, error) {

	queryCoster := querycost.GetQueryCoster(querycost.FormulaQueryCosterType)
	var (
		tableID         string
		err             error
		cost            int64
		rollbackInfoStr string
	)
	logger := sc.GetLogger()
	generalRollback := !sc.VDE && sc.Rollback

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
	if len(logData) == 0 {
		logger.WithFields(log.Fields{"type": consts.NotFound, "err": errUpdNotExistRecord, "query": selectQuery}).Error("updating for not existing record")
		return 0, tableID, errUpdNotExistRecord
	}
	jsonFields := make(map[string]map[string]string)

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
		} else if strings.Contains(fields[i], `->`) {
			colfield := strings.Split(fields[i], `->`)
			if len(colfield) == 2 {
				if jsonFields[colfield[0]] == nil {
					jsonFields[colfield[0]] = make(map[string]string)
				}
				jsonFields[colfield[0]][colfield[1]] = values[i]
			}
		} else {
			addSQLUpdate += fields[i] + `='` + strings.Replace(values[i], `'`, `''`, -1) + `',`
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

/*
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
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `+` + values[i] + `,`
			} else if fields[i][:1] == "-" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `-` + values[i] + `,`
			} else if values[i] == `NULL` {
				addSQLUpdate += fields[i] + `= NULL,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLUpdate += fields[i][len(`timestamp `):] + `= to_timestamp('` + values[i] + `'),`
			} else if strings.HasPrefix(values[i], `timestamp `) {
				addSQLUpdate += fields[i] + `= timestamp '` + values[i][len(`timestamp `):] + `',`
			} else if strings.Contains(fields[i], `->`) {
				colfield := strings.Split(fields[i], `->`)
				if len(colfield) == 2 {
					if jsonFields[colfield[0]] == nil {
						jsonFields[colfield[0]] = make(map[string]string)
					}
					jsonFields[colfield[0]][colfield[1]] = values[i]
				}
			} else {
				addSQLUpdate += fields[i] + `='` + strings.Replace(values[i], `'`, `''`, -1) + `',`
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
				tableID = fmt.Sprint(values[i])
			}

			if strings.Contains(fields[i], `->`) {
				colfield := strings.Split(fields[i], `->`)
				if len(colfield) == 2 {
					if jsonFields[colfield[0]] == nil {
						jsonFields[colfield[0]] = make(map[string]string)
					}
					jsonFields[colfield[0]][colfield[1]] = values[i]
					continue
				}
			}

			if fields[i][:1] == "+" || fields[i][:1] == "-" {
				addSQLIns0 = append(addSQLIns0, fields[i][1:len(fields[i])])
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLIns0 = append(addSQLIns0, fields[i][len(`timestamp `):])
			} else {
				addSQLIns0 = append(addSQLIns0, fields[i])
			}
			if converter.IsByteColumn(table, fields[i]) && len(values[i]) != 0 {
				addSQLIns1 = append(addSQLIns1, `decode('`+hex.EncodeToString([]byte(values[i]))+`','HEX')`)
			} else if values[i] == `NULL` {
				addSQLIns1 = append(addSQLIns1, `NULL`)
			} else if strings.HasPrefix(fields[i], `timestamp`) {
				addSQLIns1 = append(addSQLIns1, `to_timestamp('`+values[i]+`')`)
			} else if strings.HasPrefix(values[i], `timestamp`) {
				addSQLIns1 = append(addSQLIns1, `timestamp '`+values[i][len(`timestamp `):]+`'`)
			} else {
				addSQLIns1 = append(addSQLIns1, `'`+strings.Replace(values[i], `'`, `''`, -1)+`'`)
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
					tableID = fmt.Sprint(whereValues[i])
				}
				addSQLIns0 = append(addSQLIns0, whereFields[i])
				addSQLIns1 = append(addSQLIns1, whereValues[i])
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
*/
