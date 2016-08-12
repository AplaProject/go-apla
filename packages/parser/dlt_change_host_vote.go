package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) DLTChangeHostVoteInit() error {

	fields := []map[string]string{{"host": "string"}, {"vote": "string"}, {"public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}

	p.TxMaps.Bytes["public_key"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	p.TxMap["public_key"] = utils.BinToHex(p.TxMap["public_key"])
	return nil
}

func (p *Parser) DLTChangeHostVoteFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"host": "host", "vote": "sha1"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

/*
	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["walletAddress"], p.TxMap["sell_currency_id"], p.TxMap["sell_rate"], p.TxMap["amount"], p.TxMap["buy_currency_id"], p.TxMap["commission"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
*/
	return nil
}

func (p *Parser) DLTChangeHostVote() error {
	var err error
	if len(p.TxMaps.Bytes["public_key"]) > 0 {
		err = p.ExecSql(`UPDATE dlt_wallets SET host = ?, vote = [hex], public_key_0 = [hex] WHERE wallet_id = ?`, p.TxMaps.String["host"], p.TxMaps.String["vote"], p.TxMaps.Bytes["public_key"], p.TxWalletID)
	} else {
		err = p.ExecSql(`UPDATE dlt_wallets SET host = ?, vote = [hex] WHERE wallet_id = ?`, p.TxMaps.String["host"], p.TxMaps.String["vote"], p.TxWalletID)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DLTChangeHostVoteRollback() error {

	return nil
}

func (p *Parser) DLTChangeHostVoteRollbackFront() error {

	return nil

}
