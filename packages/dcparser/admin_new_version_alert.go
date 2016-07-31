package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) AdminNewVersionAlertInit() error {

	fields := []map[string]string{{"soft_type": "string"}, {"version": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminNewVersionAlertFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"version": "version", "soft_type": "soft_type"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	alert, err := p.Single("SELECT alert FROM new_version WHERE version  =  ?", p.TxMaps.String["version"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if alert == 1 {
		return p.ErrInfo("alert == 1")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["soft_type"], p.TxMap["version"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) AdminNewVersionAlert() error {
	err := p.ExecSql("UPDATE new_version SET alert = 1 WHERE version = ?", p.TxMaps.String["version"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminNewVersionAlertRollback() error {
	err := p.ExecSql("UPDATE new_version SET alert = 0 WHERE version = ?", p.TxMaps.String["version"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminNewVersionAlertRollbackFront() error {
	return nil
}
