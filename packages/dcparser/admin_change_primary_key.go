package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (p *Parser) AdminChangePrimaryKeyInit() error {

	fields := []map[string]string{{"for_user_id": "int64"}, {"public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["public_key_hex"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	p.TxMap["public_key_hex"] = utils.BinToHex(p.TxMap["public_key"])
	return nil
}

func (p *Parser) AdminChangePrimaryKeyFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"for_user_id": "user_id", "public_key_hex": "public_key"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	var txTime int64
	if p.BlockData != nil { // тр-ия пришла в блоке
		txTime = p.BlockData.Time
	} else {
		txTime = time.Now().Unix() - 30 // просто на всякий случай небольшой запас
	}

	data, err := p.OneRow("SELECT user_id, change_key, change_key_time, change_key_close FROM users WHERE user_id  =  ?", p.TxMaps.Int64["for_user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(data) == 0 {
		return p.ErrInfo("incorrect for_user_id")
	}
	// разрешил ли юзер смену ключа админом
	if data["change_key"] == 0 {
		return p.ErrInfo("change_key = 0")
	}
	// юзер отменил запрос на смену ключа
	if data["change_key_close"] == 1 {
		return p.ErrInfo("change_key_close = 1")
	}

	// прошел ли месяц с момента, когда кто-то запросил смену ключа
	if p.BlockData != nil && p.BlockData.BlockId > 170770 {
		if txTime-data["change_key_time"] < consts.CHANGE_KEY_PERIOD {
			return p.ErrInfo("CHANGE_KEY_PERIOD")
		}
	} else {
		if txTime-data["change_key_time"] < consts.CHANGE_KEY_PERIOD_170770 {
			return p.ErrInfo("CHANGE_KEY_PERIOD_170770")
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["for_user_id"], p.TxMap["public_key_hex"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) AdminChangePrimaryKey() error {
	return p.selectiveLoggingAndUpd([]string{"public_key_0", "public_key_1", "public_key_2", "change_key_close"}, []interface{}{p.TxMaps.Bytes["public_key_hex"], "", "", "1"}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["for_user_id"])})
}

func (p *Parser) AdminChangePrimaryKeyRollback() error {
	return p.selectiveRollback([]string{"public_key_0", "public_key_1", "public_key_2", "change_key_close"}, "users", "user_id="+utils.Int64ToStr(p.TxMaps.Int64["for_user_id"]), false)
}

func (p *Parser) AdminChangePrimaryKeyRollbackFront() error {
	return nil
}
