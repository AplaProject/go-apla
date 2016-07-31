package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) VotesExchangeInit() error {
	fields := []map[string]string{{"e_owner_id": "int64"}, {"result": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesExchangeFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"e_owner_id": "bigint", "result": "vote"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, есть ли такой юзер
	userId, err := p.Single("SELECT user_id FROM users WHERE user_id  =  ?", p.TxMaps.Int64["e_owner_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if userId == 0 {
		return p.ErrInfo("userId == 0")
	}

	// не было ли уже такого же голоса от этого юзера
	num, err := p.Single("SELECT count(user_id) FROM votes_exchange WHERE user_id  =  ? AND e_owner_id  =  ? AND result = ?", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["e_owner_id"], p.TxMaps.Int64["result"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if num > 0 {
		return p.ErrInfo("exists")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["e_owner_id"], p.TxMap["result"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// лимиты на голоса, чтобы не задосили голосами
	err = p.maxDayVotes()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesExchange() error {

	err := p.selectiveLoggingAndUpd([]string{"result"}, []interface{}{p.TxMaps.Int64["result"]}, "votes_exchange", []string{"user_id", "e_owner_id"}, []string{string(p.TxMap["user_id"]), string(p.TxMap["e_owner_id"])})
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesExchangeRollback() error {
	err := p.selectiveRollback([]string{"result"}, "votes_exchange", "user_id="+string(p.TxMap["user_id"])+" AND e_owner_id = "+string(p.TxMap["e_owner_id"]), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) VotesExchangeRollbackFront() error {
	return p.maxDayVotesRollback()
}
