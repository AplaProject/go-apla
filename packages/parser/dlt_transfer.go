package parser

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (p *Parser) DLTTransferInit() error {

	fields := []map[string]string{{"walletAddress": "bytes"}, {"amount": "int64"},  {"commission": "int64"}, {"comment": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DLTTransferFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"walletAddress": "sha1", "amount": "int", "commission": "int", "comment": "comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxMaps.Money["amount"] == 0 {
		return p.ErrInfo("amount=0")
	}

	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if p.TxMaps.Money["commission"] < consts.COMMISSION {
		return p.ErrInfo("commission")
	}

	// есть ли нужная сумма на кошельке
	// .....
/*
	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["walletAddress"], p.TxMap["sell_currency_id"], p.TxMap["sell_rate"], p.TxMap["amount"], p.TxMap["buy_currency_id"], p.TxMap["commission"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.checkSpamMoney(p.TxMaps.Int64["sell_currency_id"], p.TxMaps.Money["amount"])
	if err != nil {
		return p.ErrInfo(err)
	}
*/
	return nil
}

func (p *Parser) DLTTransfer() error {
	err := p.ExecSql(`INSERT INTO dlt_transactions ( recipient_wallet_address, amount, commission, comment, time, block_id ) VALUES ( [hex], ?, ?, ?, ?, ? )`, p.TxMaps.Bytes["walletAddress"], p.TxMaps.Int64["amount"], p.TxMaps.Int64["commission"],p.TxMaps.Bytes["comment"], p.BlockData.Time, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DLTTransferRollback() error {

	return nil
}

func (p *Parser) DLTTransferRollbackFront() error {

	return nil

}
