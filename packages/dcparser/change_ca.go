package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeCaInit() error {

	fields := []map[string]string{{"ca1": "string"}, {"ca2": "string"}, {"ca3": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeCaFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	if !utils.CheckInputData(p.TxMaps.String["ca1"], "ca_url") && p.TxMaps.String["ca1"] != "0" {
		return fmt.Errorf("incorrect ca1")
	}
	if !utils.CheckInputData(p.TxMaps.String["ca2"], "ca_url") && p.TxMaps.String["ca2"] != "0" {
		return fmt.Errorf("incorrect ca2")
	}
	if !utils.CheckInputData(p.TxMaps.String["ca3"], "ca_url") && p.TxMaps.String["ca3"] != "0" {
		return fmt.Errorf("incorrect ca3")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["ca1"], p.TxMap["ca2"], p.TxMap["ca3"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CHANGE_CA, "change_ca", consts.LIMIT_CHANGE_CA_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeCa() error {
	return p.selectiveLoggingAndUpd([]string{"ca1", "ca2", "ca3"}, []interface{}{p.TxMaps.String["ca1"], p.TxMaps.String["ca2"], p.TxMaps.String["ca3"]}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
}

func (p *Parser) ChangeCaRollback() error {
	return p.selectiveRollback([]string{"ca1", "ca2", "ca3"}, "users", "user_id="+utils.Int64ToStr(p.TxUserID), false)
}

func (p *Parser) ChangeCaRollbackFront() error {
	return p.limitRequestsRollback("change_ca")
}
