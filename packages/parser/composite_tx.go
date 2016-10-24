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
	"github.com/EGaaS/go-mvp/packages/utils"
	"fmt"
)

func (p *Parser) CompositeTxInit() error {

	// get fields from DB
	// ...
	err := p.GetTxMaps([]map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CompositeTxFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string]string{}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}


	// Check the condition that must be met to complete this transaction
	// select front from composite_tx where name = "new_state_table"
	// ...



	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d", p.TxMap["type"], p.TxMap["time"], p.TxMap["state_id"], p.TxCitizenID)
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) CompositeTx() error {
/*
	retirees := getDataFromDB(ea_retirees)
for data := range retirees {
  // пока что запрещаем всё, кроме:
  // update других таблиц через наш метод selectiveLoggingAndUpd() т.к. это легко роллбечится
  // можно делать операции с данными, которые далее будут записаны через selectiveLoggingAndUpd()
  // можно вставить формулы sum := data.k1 * 0.1 / data.k2
  // вложенные циклы, условия и т.д. - всё запрещаем. особенно важно не трогать таблу, по которой цикл идет
  // insert в другие таблицы разрашаем, это роллбечить вобще легко, т.к. есть номер блока. разумеется, данные, которые были только что вставлены не должны быть использованы в этом же блоке

  // есть список запрщенных таблиц для selectiveLoggingAndUpd, например accounts

  // условные операторы - надо понять, можно ли при помощи них сделать так, чтобы роллбек что-то не учел.
}
*/
	return nil
}

func (p *Parser) CompositeTxRollback() error {

	return nil
}

/*func (p *Parser) CompositeTxRollbackFront() error {

	return nil

}
*/