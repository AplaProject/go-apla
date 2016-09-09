package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (p *Parser) CitizenRequestInit() error {
	fmt.Println(`CitizenRequestInit`)
/*	fields := []map[string]string{{"state_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["sign"] = utils.BinToHex(p.TxMaps.Bytes["sign"])*/
	fmt.Println(p.TxPtr.(*consts.CitizenRequest))
	return nil
}

func (p *Parser) CitizenRequestFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"state_id": "int64"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, есть ли такое гос-во


	// есть ли сумма, которую просит гос-во за регистрацию гражданства в DLT


	forSign := fmt.Sprintf("%s,%s,%d", p.TxMap["type"], p.TxMap["time"], p.TxWalletID)
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// есть ли нужная сумма на кошельке
	amount, err := p.Single(`SELECT value FROM dn_state_settings WHERE parameter = ?`, "citizen_dlt_price").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	amountAndCommission, err := p.checkSenderMoney(amount, consts.COMMISSION)
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
	}

	return nil
}

func (p *Parser) CitizenRequest() error {

	stateCode, err := p.Single(`SELECT state_code FROM states WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	// пишем в общую историю тр-ий
	err = p.ExecSql(`INSERT INTO `+stateCode+`_citizens_requests ( dlt_wallet_is, block_id ) VALUES ( ?, ? )`, p.TxWalletID, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CitizenRequestRollback() error {

	stateCode, err := p.Single(`SELECT state_code FROM states WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	// пишем в общую историю тр-ий
	err = p.ExecSql(`DELETE FROM `+stateCode+`_citizens_requests WHERE block_id = ?`, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CitizenRequestRollbackFront() error {

	return nil

}
