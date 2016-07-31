package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (p *Parser) MoneyBackRequestInit() error {

	fields := []map[string]string{{"order_id": "int64"}, {"arbitrator0_enc_text": "bytes"}, {"arbitrator1_enc_text": "bytes"}, {"arbitrator2_enc_text": "bytes"}, {"arbitrator3_enc_text": "bytes"}, {"arbitrator4_enc_text": "bytes"}, {"seller_enc_text": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) MoneyBackRequestFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"order_id": "bigint", "seller_enc_text": "comment", "arbitrator0_enc_text": "comment", "arbitrator1_enc_text": "comment", "arbitrator2_enc_text": "comment", "arbitrator3_enc_text": "comment", "arbitrator4_enc_text": "comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	var txTime int64
	if p.BlockData != nil { // тр-ия пришла в блоке
		txTime = p.BlockData.Time
	} else { // голая тр-ия
		txTime = time.Now().Unix() - 30 // просто на всякий случай небольшой запас
	}

	// проверим, есть ли такой ордер, не был ли ранее запрос, точно ли покупатель наш юзер
	orderId, err := p.Single("SELECT id FROM orders WHERE id  =  ? AND status  =  'normal' AND end_time > ? AND buyer  =  ?", p.TxMaps.Int64["order_id"], txTime, p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if orderId == 0 {
		return p.ErrInfo("orderId==0")
	}

	forSign := ""
	if p.BlockData != nil && p.BlockData.BlockId < 197115 {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["order_id"], p.TxMap["arbitrator0_enc_text"], p.TxMap["arbitrator1_enc_text"], p.TxMap["arbitrator2_enc_text"], p.TxMap["arbitrator3_enc_text"], p.TxMap["arbitrator4_enc_text"], p.TxMap["seller_enc_text"])

	} else {
		encData := make(map[string]string)
		for i := 0; i < 5; i++ {
			iStr := utils.IntToStr(i)
			encData["arbitrator"+iStr+"_enc_text"] = string(utils.BinToHex(p.TxMap["arbitrator"+iStr+"_enc_text"]))
			if encData["arbitrator"+iStr+"_enc_text"] == "00" {
				encData["arbitrator"+iStr+"_enc_text"] = "0"
			}
		}
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["order_id"], encData["arbitrator0_enc_text"], encData["arbitrator1_enc_text"], encData["arbitrator2_enc_text"], encData["arbitrator3_enc_text"], encData["arbitrator4_enc_text"], utils.BinToHex(p.TxMap["seller_enc_text"]))
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_MONEY_BACK_REQUEST, "money_back_request", consts.LIMIT_MONEY_BACK_REQUEST_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) MoneyBackRequest() error {

	err := p.selectiveLoggingAndUpd([]string{"status"}, []interface{}{"refund"}, "orders", []string{"id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["order_id"])})
	if err != nil {
		return p.ErrInfo(err)
	}

	orderData, err := p.OneRow("SELECT seller, arbitrator0, arbitrator1, arbitrator2, arbitrator3, arbitrator4 FROM orders WHERE id  =  ?", p.TxMaps.Int64["order_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, не является ли мы продавцом или арбитром
	myUserId, _, myPrefix, _, err := p.GetMyUserId(orderData["seller"])
	if err != nil {
		return p.ErrInfo(err)
	}
	if orderData["seller"] == myUserId {
		err = p.ExecSql("INSERT INTO "+myPrefix+"my_comments ( type, id, comment, comment_status ) VALUES ( 'seller', ?, ?, 'encrypted' )", p.TxMaps.Int64["order_id"], utils.BinToHex(p.TxMap["seller_enc_text"]))
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	for i := 0; i < 5; i++ {
		iStr := utils.IntToStr(i)
		if orderData["arbitrator"+iStr] == 0 {
			continue
		}
		myUserId, _, myPrefix, _, err := p.GetMyUserId(orderData["arbitrator"+iStr])
		if err != nil {
			return p.ErrInfo(err)
		}
		if orderData["arbitrator"+iStr] == myUserId {
			err = p.ExecSql("INSERT INTO "+myPrefix+"my_comments ( type, id, comment, comment_status ) VALUES ( 'arbitrator', ?, ?, 'encrypted' )", p.TxMaps.Int64["order_id"], utils.BinToHex(p.TxMaps.Bytes["arbitrator"+iStr+"_enc_text"]))
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	return nil
}

func (p *Parser) MoneyBackRequestRollback() error {
	orderData, err := p.OneRow("SELECT seller, arbitrator0, arbitrator1, arbitrator2, arbitrator3, arbitrator4 FROM orders WHERE id  =  ?", p.TxMaps.Int64["order_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, не является ли мы продавцом или арбитром
	myUserId, _, myPrefix, _, err := p.GetMyUserId(orderData["seller"])
	if err != nil {
		return p.ErrInfo(err)
	}
	if orderData["seller"] == myUserId {
		err = p.ExecSql("DELETE FROM "+myPrefix+"my_comments WHERE type = 'seller' AND id = ?", p.TxMaps.Int64["order_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	for i := 0; i < 5; i++ {
		iStr := utils.IntToStr(i)
		if orderData["arbitrator"+iStr] == 0 {
			continue
		}
		myUserId, _, myPrefix, _, err := p.GetMyUserId(orderData["arbitrator"+iStr])
		if err != nil {
			return p.ErrInfo(err)
		}
		if orderData["arbitrator"+iStr] == myUserId {
			err = p.ExecSql("DELETE FROM "+myPrefix+"my_comments WHERE type = 'arbitrator' AND id = ?", p.TxMaps.Int64["order_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	err = p.selectiveRollback([]string{"status"}, "orders", "id="+utils.Int64ToStr(p.TxMaps.Int64["order_id"]), false)

	return nil
}

func (p *Parser) MoneyBackRequestRollbackFront() error {
	return p.limitRequestsRollback("money_back_request")
}
