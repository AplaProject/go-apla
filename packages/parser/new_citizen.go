package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
	"encoding/json"
)

func (p *Parser) NewCitizenInit() error {

	err := p.GetTxMaps(nil)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewCitizenFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// получим набор доп. полей, которые должны быть в данной тр-ии
	additionalFields, err := p.Single(`SELECT fields FROM citizen_fields WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).Bytes()
	if err != nil {
		return p.ErrInfo(err)
	}

	additionalFieldsMap := []map[string]string{}
	err = json.Unmarshal(additionalFields, &additionalFieldsMap)
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := make(map[string]string)
	for _, date := range additionalFieldsMap {
		verifyData[date["name"]] = date["txType"]
	}
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
	amountAndCommission, err := p.checkSenderMoney(p.TxMaps.Int64["amount"], p.TxMaps.Int64["commission"])
	if err != nil {
		return p.ErrInfo(err)
	}

	amount, err := p.Single(`SELECT value FROM dn_state_settings WHERE parameter = ?`, "citizen_dlt_price").Int64()
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

func (p *Parser) NewCitizen() error {

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

func (p *Parser) NewCitizenRollback() error {

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

func (p *Parser) NewCitizenRollbackFront() error {

	return nil

}
