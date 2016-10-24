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
	"github.com/EGaaS/go-mvp/packages/consts"
	"github.com/EGaaS/go-mvp/packages/utils"
)

func (p *Parser) CitizenRequestInit() error {
	/*	fields := []map[string]string{{"state_id": "int64"}, {"sign": "bytes"}}
		err := p.GetTxMaps(fields)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.TxMaps.Bytes["sign"] = utils.BinToHex(p.TxMaps.Bytes["sign"])*/
	p.TxVars[`state_code`] = utils.Int64ToStr(p.TxPtr.(*consts.CitizenRequest).StateId)
	//	fmt.Println(data, p.States)
	return nil
}

func (p *Parser) CitizenRequestFront() error {

	if err := p.generalCheckStruct(``); err != nil {
		return p.ErrInfo(err)
	}
	// проверим, есть ли такое гос-во

	// есть ли сумма, которую просит гос-во за регистрацию гражданства в DLT
	// Проверка подписи перенесена в generalCheckStruct

	/*amount, err := p.Single(`SELECT value FROM `+p.TxVars[`state_code`]+`_state_parameters WHERE name = ?`, "citizenship_price").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}*/

	/*amountAndCommission, err := p.checkSenderDLT(amount, consts.COMMISSION)
	if err != nil {
		return p.ErrInfo(err)
	}
	if amount > amountAndCommission {
		return p.ErrInfo("incorrect amount")
	}
	// вычитаем из wallets_buffer
	// amount_and_commission взято из check_sender_money()
	err = p.updateWalletsBuffer(amountAndCommission)
	if err != nil {
		return p.ErrInfo(err)
	}*/
	return nil
}

func (p *Parser) CitizenRequest() error {

	// пишем в общую историю тр-ий
	err := p.ExecSql(`INSERT INTO `+p.TxVars[`state_code`]+
		`_citizenship_requests ( dlt_wallet_id, block_id ) VALUES ( ?, ? )`,
		p.TxWalletID, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}

	// вычитаем
	amount, err := p.Single(`SELECT value FROM `+p.TxVars[`state_code`]+`_state_parameters WHERE name = ?`, "citizenship_price").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"-amount"}, []interface{}{amount}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) CitizenRequestRollback() error {
	// пишем в общую историю тр-ий
	err := p.ExecSql(`DELETE FROM `+p.TxVars[`state_code`]+
		`_citizenship_requests WHERE block_id = ?`, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.autoRollback()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

/*func (p *Parser) CitizenRequestRollbackFront() error {
	err := p.ExecSql("DELETE FROM wallets_buffer WHERE hex(hash) = ?", p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}*/
