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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// roll back the ID to the number of affected rows
func (p *Parser) rollbackAI(table string, num int64) error {
	if num == 0 {
		return nil
	}
	AiID, err := model.GetAiID(table)
	if err != nil {
		return utils.ErrInfo(err)
	}
	tblname := converter.EscapeName(table)
	log.Debug("AiID: %s", AiID)
	// if the table was cleaned up, then 0 appears, that's why we can not clean the tables to zero
	current, err := model.GetCurrentSeqID(AiID, tblname)
	if err != nil {
		return utils.ErrInfo(err)
	}
	NewAi := current + num
	log.Debug("NewAi: %d", NewAi)
	pgSerialSeq, err := model.GetSerialSequence(table, AiID)
	if err != nil {
		return utils.ErrInfo(err)
	}
	err = model.SequenceRestartWith(pgSerialSeq, NewAi)
	if err != nil {
		return utils.ErrInfo(err)
	}
	return nil
}
