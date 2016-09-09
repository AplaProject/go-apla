package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
	"encoding/json"
)

func (p *Parser) NewCitizenInit() error {

	fields := []map[string]string{{"public_key": "bytes"}, {"state_id": "int64"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMap["public_key_hex"] = utils.BinToHex(p.TxMap["public_key"])
	p.TxMaps.Bytes["public_key_hex"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	return nil
}

func (p *Parser) NewCitizenFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// To not record too small or too big key
	if !utils.CheckInputData(p.TxMap["public_key_hex"], "public_key") {
		return utils.ErrInfo(fmt.Errorf("incorrect public_key %s", p.TxMap["public_key_hex"]))
	}

	// We get a set of custom fields that need to be in the tx
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

	// Citizens can only add a citizen of the same country

	// One who adds a citizen must be a valid representative body appointed in ds_state_settings


	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d", p.TxMap["type"], p.TxMap["time"], p.TxWalletID)
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) NewCitizen() error {

	stateCode, err := p.Single(`SELECT state_code FROM states WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`INSERT INTO `+stateCode+`_citizens ( public_key_0, block_id ) VALUES ( [hex], ? )`, p.TxMap["public_key_hex"], p.BlockData.BlockId)
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
	err = p.ExecSql(`DELETE FROM `+stateCode+`_citizens WHERE block_id = ?`, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewCitizenRollbackFront() error {

	return nil

}
