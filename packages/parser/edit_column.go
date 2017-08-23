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
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type EditColumnParser struct {
	*Parser
	EditColumn *tx.EditColumn
}

func (p *EditColumnParser) Init() error {
	editColumn := &tx.EditColumn{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editColumn); err != nil {
		return p.ErrInfo(err)
	}
	p.EditColumn = editColumn
	return nil
}

func (p *EditColumnParser) Validate() error {
	err := p.generalCheck(`edit_column`, &p.EditColumn.Header, map[string]string{"permissions": p.EditColumn.Permissions})
	if err != nil {
		return p.ErrInfo(err)
	}

	stateIdStr := converter.Int64ToStr(p.EditColumn.Header.StateID)
	prefix := stateIdStr
	if strings.HasPrefix(stateIdStr, `global`) {
		prefix = `global`
	}
	tEx := &model.Table{}
	tEx.SetTablePrefix(prefix)
	exists, err := tEx.IsExistsByPermissionsAndTableName(p.EditColumn.ColumnName, p.EditColumn.TableName)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !exists {
		return p.ErrInfo(`column not exists`)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditColumn.ForSign(), p.EditColumn.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessTable(p.EditColumn.TableName, `general_update`); err != nil {
		return err
	}

	return nil
}

func (p *EditColumnParser) Action() error {
	stateIdStr := converter.Int64ToStr(p.EditColumn.Header.StateID)
	table := stateIdStr + `_tables`
	if strings.HasPrefix(p.EditColumn.TableName, `global`) {
		table = `global_tables`
	}
	logData, err := model.GetTableWhereUpdatePermissionAndTableName(table, p.EditColumn.ColumnName, p.EditColumn.TableName)
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
	rb := &model.Rollback{
		Data:    string(jsonData),
		BlockID: p.BlockData.BlockID}
	err = rb.Create()
	if err != nil {
		return err
	}
	tableM := &model.Table{}
	_, err = tableM.SetActionByName(table, p.EditColumn.TableName, "update, "+p.EditColumn.ColumnName, `"`+converter.EscapeForJSON(p.EditColumn.Permissions)+`"`, rb.RbID)
	if err != nil {
		return p.ErrInfo(err)
	}

	rbTx := &model.RollbackTx{
		BlockID:   p.BlockData.BlockID,
		TxHash:    p.TxHash,
		NameTable: table,
		TableID:   p.EditColumn.TableName,
	}
	err = rbTx.Create()
	if err != nil {
		return err
	}

	return nil
}

func (p *EditColumnParser) Rollback() error {
	return p.autoRollback()
}

func (p EditColumnParser) Header() *tx.Header {
	return &p.EditColumn.Header
}
