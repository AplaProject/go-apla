package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

func (p *Parser) AdminBanUnbanChatInit() error {

	fields := []map[string]string{{"users_ids": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminBanUnbanChatFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"users_ids": "users_ids"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["users_ids"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) AdminBanUnbanChat() error {

	users_ids := strings.Split(p.TxMaps.String["users_ids"], ",")

	for i := 0; i < len(users_ids); i++ {
		userId := utils.StrToInt64(users_ids[i])
		err := p.selectiveLoggingAndUpd([]string{"chat_ban"}, []interface{}{1}, "users", []string{"user_id"}, []string{utils.Int64ToStr(userId)})
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) AdminBanUnbanChatRollback() error {

	users_ids := strings.Split(p.TxMaps.String["users_ids"], ",")

	for i := 0; i < len(users_ids); i++ {
		err := p.selectiveRollback([]string{"chat_ban"}, "users", "user_id="+users_ids[i], false)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) AdminBanUnbanChatRollbackFront() error {
	return nil
}
