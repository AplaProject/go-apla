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
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
)

func (p *Parser) DLTTransferInit() error {

	fields := []map[string]string{{"walletAddress": "bytes"}, {"amount": "int64"},  {"commission": "int64"}, {"comment": "bytes"},{"public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["public_key"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	p.TxMap["public_key"] = utils.BinToHex(p.TxMap["public_key"])
	p.TxMaps.Bytes["sign"] = utils.BinToHex(p.TxMaps.Bytes["sign"])
	return nil
}

func (p *Parser) DLTTransferFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"walletAddress": "walletAddress", "amount": "int64", "commission": "int64", "comment": "comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxMaps.Int64["amount"] == 0 {
		return p.ErrInfo("amount=0")
	}

	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if p.TxMaps.Int64["commission"] < consts.COMMISSION {
		return p.ErrInfo("commission")
	}

	if p.TxMaps.Int64["amount"] <= 0 {
		return p.ErrInfo("amount<=0")
	}

	forSign := fmt.Sprintf("%s,%s,%d,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxWalletID, p.TxMap["walletAddress"], p.TxMap["amount"], p.TxMap["commission"], p.TxMap["comment"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// есть ли нужная сумма на кошельке
	amountAndCommission, err := p.checkSenderDLT(p.TxMaps.Int64["amount"], p.TxMaps.Int64["commission"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// вычитаем из wallets_buffer
	// amount_and_commission взято из check_sender_money()
	err = p.updateWalletsBuffer(amountAndCommission)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) DLTTransfer() error {
	log.Debug("wallet address %s", p.TxMaps.Bytes["walletAddress"])
	log.Debug("wallet address hex %x", utils.B54Decode(p.TxMaps.Bytes["walletAddress"]))
	hexAddress := utils.BinToHex(utils.B54Decode(p.TxMaps.Bytes["walletAddress"]))
	walletId, err := p.Single(`SELECT wallet_id FROM dlt_wallets WHERE address = [hex]`, hexAddress).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	//if walletId > 0 {
		if len(p.TxMaps.Bytes["public_key"]) > 0 {
			err = p.selectiveLoggingAndUpd([]string{"+amount", "public_key_0"}, []interface{}{p.TxMaps.Int64["amount"], p.TxMaps.Bytes["public_key"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(walletId)}, true)
		} else {
			err = p.selectiveLoggingAndUpd([]string{"+amount"}, []interface{}{p.TxMaps.Int64["amount"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(walletId)}, true)
		}
		if err != nil {
			return p.ErrInfo(err)
		}
	// пишем в общую историю тр-ий
	dlt_transactions_id, err := p.ExecSqlGetLastInsertId(`INSERT INTO dlt_transactions ( sender_wallet_id, recipient_wallet_id, recipient_wallet_address, amount, commission, comment, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )`, "dlt_transactions", p.TxWalletID, walletId, p.TxMaps.Bytes["walletAddress"], p.TxMaps.Int64["amount"], p.TxMaps.Int64["commission"],p.TxMaps.Bytes["comment"], p.BlockData.Time, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, "dlt_transactions", dlt_transactions_id)
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) DLTTransferRollback() error {
	/*hexAddress := utils.BinToHex(utils.B54Decode(p.TxMaps.Bytes["walletAddress"]))

	walletId, err := p.Single(`SELECT wallet_id FROM dlt_wallets WHERE address = [hex]`, hexAddress).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	rbId, err := p.Single(`SELECT rb_id FROM dlt_wallets WHERE address = [hex]`, hexAddress).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// Если это не первая запись, а обновление
	if rbId > 0 {
		if len(p.TxMaps.Bytes["public_key"]) > 0 {
			err := p.selectiveRollback([]string{"public_key_0", "amount"}, "dlt_wallets", "wallet_id="+utils.Int64ToStr(walletId), false)
			if err != nil {
				return p.ErrInfo(err)
			}
		} else {
			err := p.selectiveRollback([]string{"amount"}, "dlt_wallets", "wallet_id="+utils.Int64ToStr(walletId), false)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	} else {
		err = p.ExecSql(`DELETE FROM dlt_wallets WHERE wallet_id = ?`, walletId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	err = p.ExecSql(`DELETE FROM dlt_transactions WHERE block_id = ?`, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil*/
	return p.autoRollback()
}

/*func (p *Parser) DLTTransferRollbackFront() error {

	return nil

}
*/