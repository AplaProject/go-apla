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
	p.TxMap["sign"] = utils.BinToHex(p.TxMap["sign"])
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

	// есть ли нужная сумма на кошельке
	walletId, err := p.Single(`SELECT wallet_id FROM dlt_wallets WHERE address = [hex]`, hexAddress).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxWalletID, p.TxCitizenID, p.TxMap["walletAddress"], p.TxMap["amount"], p.TxMap["commission"], p.TxMap["comment"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
/*
	err = p.checkSpamMoney(p.TxMaps.Int64["sell_currency_id"], p.TxMaps.Money["amount"])
	if err != nil {
		return p.ErrInfo(err)
	}
*/
	return nil
}

func (p *Parser) DLTTransfer() error {
	hexAddress := utils.BinToHex(utils.B54Decode(p.TxMaps.Bytes["walletAddress"]))
	walletId, err := p.Single(`SELECT wallet_id FROM dlt_wallets WHERE address = [hex]`, hexAddress).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if walletId > 0 {
		if len(p.TxMaps.Bytes["public_key"]) > 0 {
			err = p.selectiveLoggingAndUpd([]string{"+amount", "public_key_0"}, []interface{}{p.TxMaps.Int64["amount"], p.TxMaps.Bytes["public_key"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(walletId)})
			//err = p.ExecSql(`UPDATE dlt_wallets SET amount = amount + ?, public_key_0 = [hex] WHERE wallet_id = ?`, p.TxMaps.Int64["amount"],  p.TxMaps.Bytes["public_key"], walletId)
		} else {
			err = p.selectiveLoggingAndUpd([]string{"+amount"}, []interface{}{p.TxMaps.Int64["amount"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(walletId)})
			//err = p.ExecSql(`UPDATE dlt_wallets SET amount = amount + ? WHERE wallet_id = ?`, p.TxMaps.Int64["amount"], walletId)
		}
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		if len(p.TxMaps.Bytes["public_key"]) > 0 {
			err = p.ExecSql(`INSERT INTO dlt_wallets (address, amount, public_key_0) VALUES ([hex], ?, [hex])`, hexAddress, p.TxMaps.Int64["amount"], p.TxMaps.Bytes["public_key"])
		} else {
			err = p.ExecSql(`INSERT INTO dlt_wallets (address, amount) VALUES ([hex], ?)`, hexAddress, p.TxMaps.Int64["amount"])
		}
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// пишем в общую историю тр-ий
	err = p.ExecSql(`INSERT INTO dlt_transactions ( sender_wallet_id, recipient_wallet_id, recipient_wallet_address, amount, commission, comment, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )`, p.TxWalletID, walletId, p.TxMaps.Bytes["walletAddress"], p.TxMaps.Int64["amount"], p.TxMaps.Int64["commission"],p.TxMaps.Bytes["comment"], p.BlockData.Time, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DLTTransferRollback() error {
	hexAddress := utils.BinToHex(utils.B54Decode(p.TxMaps.Bytes["walletAddress"]))

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
	return nil
}

func (p *Parser) DLTTransferRollbackFront() error {

	return nil

}
