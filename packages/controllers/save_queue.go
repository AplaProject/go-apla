package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (c *Controller) SaveQueue() (string, error) {

	var err error
	c.r.ParseForm()

	userId := []byte(c.r.FormValue("user_id"))
	if !utils.CheckInputData(userId, "int") {
		return `{"result":"incorrect userId"}`, nil
	}
	txTime := utils.StrToInt64(c.r.FormValue("time"))
	if !utils.CheckInputData(txTime, "int") {
		return `{"result":"incorrect time"}`, nil
	}
	txType_ := c.r.FormValue("type")
	if !utils.CheckInputData(txType_, "type") {
		return `{"result":"incorrect type"}`, nil
	}
	txType := utils.TypeInt(txType_)
	signature1 := c.r.FormValue("signature1")
	signature2 := c.r.FormValue("signature2")
	signature3 := c.r.FormValue("signature3")
	sign := utils.EncodeLengthPlusData(utils.HexToBin([]byte(signature1)))
	if len(signature2) > 0 {
		sign = append(sign, utils.EncodeLengthPlusData(utils.HexToBin([]byte(signature2)))...)
	}
	if len(signature3) > 0 {
		sign = append(sign, utils.EncodeLengthPlusData(utils.HexToBin([]byte(signature3)))...)
	}
	binSignatures := utils.EncodeLengthPlusData([]byte(sign))

	log.Debug("txType_", txType_)

	var data []byte
	switch txType_ {

	case "DelForexOrder":

		orderId := []byte(c.r.FormValue("order_id"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(orderId)...)
		data = append(data, binSignatures...)


	case "ChangeNodeKey":

		publicKey := []byte(c.r.FormValue("public_key"))

		verifyData := map[string]string{c.r.FormValue("public_key"): "public_key", c.r.FormValue("private_key"): "private_key"}
		err := CheckInputData(verifyData)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(utils.HexToBin(publicKey))...)
		data = append(data, binSignatures...)


	case "NewForexOrder":

		sellCurrencyId := []byte(c.r.FormValue("sell_currency_id"))
		sellRate := []byte(c.r.FormValue("sell_rate"))
		amount := []byte(c.r.FormValue("amount"))
		buyCurrencyId := []byte(c.r.FormValue("buy_currency_id"))
		commission := []byte(c.r.FormValue("commission"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(sellCurrencyId)...)
		data = append(data, utils.EncodeLengthPlusData(sellRate)...)
		data = append(data, utils.EncodeLengthPlusData(amount)...)
		data = append(data, utils.EncodeLengthPlusData(buyCurrencyId)...)
		data = append(data, utils.EncodeLengthPlusData(commission)...)
		data = append(data, binSignatures...)

	}

	md5 := utils.Md5(data)
	if !utils.InSliceString(txType_, []string{"new_pct", "new_max_promised_amounts", "new_reduction", "votes_node_new_miner", "new_max_other_currencies"}) {
		err := c.ExecSql(`INSERT INTO transactions_status (
				hash,
				time,
				type,
				user_id
			)
			VALUES (
				[hex],
				?,
				?,
				?
			)`, md5, time.Now().Unix(), txType, userId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}
	err = c.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, utils.BinToHex(data))
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return `{"error":"null"}`, nil
}

func CheckInputData(data map[string]string) error {
	for k, v := range data {
		if !utils.CheckInputData(k, v) {
			return utils.ErrInfo(fmt.Errorf("incorrect " + v))
		}
	}
	return nil
}
