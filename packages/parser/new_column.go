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
	"encoding/json"
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewColumnParser struct {
	*Parser
	NewColumn *tx.NewColumn
}

func (p *NewColumnParser) Init() error {
	newColumn := &tx.NewColumn{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newColumn); err != nil {
		return p.ErrInfo(err)
	}
	p.NewColumn = newColumn
	return nil
}

func (p *NewColumnParser) Validate() error {
	err := p.generalCheck(`new_column`, &p.NewColumn.Header, map[string]string{"permissions": p.NewColumn.Permissions})
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string][]interface{}{"string": []interface{}{p.NewColumn.TableName, p.NewColumn.ColumnName}, "conditions": []interface{}{p.NewColumn.Permissions}, "int64": []interface{}{p.NewColumn.Index}, "column_type": []interface{}{p.NewColumn.ColumnType}}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	prefix, err := utils.GetPrefix(p.NewColumn.TableName, converter.Int64ToStr(p.NewColumn.Header.StateID))
	if err != nil {
		return p.ErrInfo(err)
	}
	tEx := &model.Table{}
	tEx.SetTablePrefix(prefix)

	exists, err := tEx.IsExistsByPermissionsAndTableName(p.NewColumn.ColumnName, p.NewColumn.TableName)
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists {
		return p.ErrInfo(`column exists`)
	}

	count, err := model.GetColumnCount(p.NewColumn.TableName)
	if count >= int64(syspar.GetMaxColumns()) {
		return fmt.Errorf(`Too many columns. Limit is %d`, syspar.GetMaxColumns())
	}
	if converter.StrToInt64(p.NewColumn.Index) > 0 {
		count, err := model.NumIndexes(p.NewColumn.TableName)
		if err != nil {
			return p.ErrInfo(err)
		}
		if count >= syspar.GetMaxIndexes() {
			return fmt.Errorf(`Too many indexes. Limit is %d`, syspar.GetMaxIndexes())
		}
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewColumn.ForSign(), p.NewColumn.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err := p.AccessTable(p.NewColumn.TableName, "new_column"); err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *NewColumnParser) Action() error {
	tblname := p.NewColumn.TableName
	prefix, err := utils.GetPrefix(tblname, converter.Int64ToStr(p.NewColumn.StateID))
	if err != nil {
		return p.ErrInfo(err)
	}
	table := prefix + `_tables`
	logData, err := model.GetColumnsAndPermissionsAndRbIDWhereTable(table, tblname)
	if err != nil {
		return err
	}

	jsonMap := make(map[string]string)
	for k, v := range logData {
		if k == p.AllPkeys[table] {
			continue
		}
		jsonMap[k] = v
		if k == "rb_id" {
			k = "prev_rb_id"
		}
	}
	jsonData, _ := json.Marshal(jsonMap)
	if err != nil {
		return err
	}
	rb := &model.Rollback{Data: string(jsonData), BlockID: p.BlockData.BlockID}
	err = rb.Create(p.DbTransaction)
	if err != nil {
		return err
	}
	tableM := &model.Table{}
	_, err = tableM.SetActionByName(p.DbTransaction, table, p.NewColumn.TableName, "update, "+p.NewColumn.ColumnName, p.NewColumn.Permissions, rb.RbID)
	if err != nil {
		return p.ErrInfo(err)
	}

	rbTx := &model.RollbackTx{
		BlockID:   p.BlockData.BlockID,
		TxHash:    p.TxHash,
		NameTable: table,
		TableID:   p.NewColumn.TableName,
	}
	err = rbTx.Create(p.DbTransaction)
	if err != nil {
		log.Errorf("something wrong: %s", err)
		//return err
	}

	colType := ``
	switch p.NewColumn.ColumnType {
	case "text":
		colType = `varchar(102400)`
	case "int64":
		colType = `bigint`
	case "time":
		colType = `timestamp`
	case "hash":
		colType = `bytea`
	case "money":
		colType = `decimal(30,0)`
	case "double":
		colType = `double precision`
	}

	err = model.AlterTableAddColumn(p.DbTransaction, tblname, p.NewColumn.ColumnName, colType)
	if err != nil {
		return err
	}

	if p.NewColumn.Index == "1" {
		err = model.CreateIndex(p.DbTransaction, tblname+"_"+p.NewColumn.ColumnName+"_index", tblname, p.NewColumn.ColumnName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *NewColumnParser) Rollback() error {
	err := p.autoRollback()
	if err != nil {
		return err
	}
	err = model.AlterTableDropColumn(p.NewColumn.TableName, p.NewColumn.ColumnName)
	if err != nil {
		return err
	}
	return nil
}

func (p NewColumnParser) Header() *tx.Header {
	return &p.NewColumn.Header
}
