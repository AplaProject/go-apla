package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) DLTChangeHostVoteInit() error {

	fields := []map[string]string{{"host": "string"}, {"vote": "bytes"}, {"public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}

	p.TxMaps.Bytes["public_key"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	p.TxMap["public_key"] = utils.BinToHex(p.TxMap["public_key"])
	p.TxMaps.Bytes["vote"] = utils.BinToHex(p.TxMaps.Bytes["vote"])
	p.TxMap["vote"] = utils.BinToHex(p.TxMap["vote"])
	return nil
}

func (p *Parser) DLTChangeHostVoteFront() error {

	/*err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"host": "host", "vote": "sha1"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}*/

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
		err = p.selectiveLoggingAndUpd([]string{"host", "vote", "public_key_0"}, []interface{}{p.TxMaps.String["host"], p.TxMaps.Bytes["vote"], p.TxMaps.Bytes["public_key"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)})
	} else {
		err = p.selectiveLoggingAndUpd([]string{"host", "vote"}, []interface{}{p.TxMaps.String["host"], p.TxMaps.String["vote"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)})
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DLTChangeHostVoteRollback() error {
	var err error
	if len(p.TxMaps.Bytes["public_key"]) > 0 {
		err = p.selectiveRollback([]string{"host", "vote", "public_key_0"}, "dlt_wallets", "", false)
	} else {
		err = p.selectiveRollback([]string{"host", "vote"}, "dlt_wallets", "", false)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DLTChangeHostVoteRollbackFront() error {
	return nil
}
