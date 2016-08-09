package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (c *Controller) SaveQueue() (string, error) {

	var err error
	c.r.ParseForm()

	citizenId := utils.BytesToInt64([]byte(c.r.FormValue("citizenId")))
	walletId := utils.BytesToInt64([]byte(c.r.FormValue("walletId")))

	if citizenId <= 0 && walletId <= 0 {
		return `{"result":"incorrect citizenId || walletId"}`, nil
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
	log.Debug("txType", txType)

	var data []byte
	switch txType_ {


	case "DLTTransfer":

		walletAddress := []byte(c.r.FormValue("walletAddress"))
		amount := []byte(c.r.FormValue("amount"))
		commission := []byte(c.r.FormValue("commission"))
		comment := []byte(c.r.FormValue("comment"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(walletId)...)
		data = append(data, utils.EncodeLengthPlusData(citizenId)...)
		data = append(data, utils.EncodeLengthPlusData(walletAddress)...)
		data = append(data, utils.EncodeLengthPlusData(amount)...)
		data = append(data, utils.EncodeLengthPlusData(commission)...)
		data = append(data, utils.EncodeLengthPlusData(comment)...)
		data = append(data, binSignatures...)

	case "DLTChangeHostVote":

		host := []byte(c.r.FormValue("host"))
		vote := []byte(c.r.FormValue("vote"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(walletId)...)
		data = append(data, utils.EncodeLengthPlusData(citizenId)...)
		data = append(data, utils.EncodeLengthPlusData(host)...)
		data = append(data, utils.EncodeLengthPlusData(vote)...)
		data = append(data, binSignatures...)


	case "DelForexOrder":

		orderId := []byte(c.r.FormValue("order_id"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
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
		data = append(data, utils.EncodeLengthPlusData(sellCurrencyId)...)
		data = append(data, utils.EncodeLengthPlusData(sellRate)...)
		data = append(data, utils.EncodeLengthPlusData(amount)...)
		data = append(data, utils.EncodeLengthPlusData(buyCurrencyId)...)
		data = append(data, utils.EncodeLengthPlusData(commission)...)
		data = append(data, binSignatures...)

	}

	log.Debug("md5")

	md5 := utils.Md5(data)


	log.Debug("md5")

	err = c.ExecSql(`INSERT INTO transactions_status (
				hash,
				time,
				type,
				wallet_id,
				citizen_id
			)
			VALUES (
				[hex],
				?,
				?,
				?,
				?
			)`, md5, time.Now().Unix(), txType, walletId, citizenId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	err = c.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, utils.BinToHex(data))
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return `{"hash":"`+string(md5)+`"}`, nil
}

func CheckInputData(data map[string]string) error {
	for k, v := range data {
		if !utils.CheckInputData(k, v) {
			return utils.ErrInfo(fmt.Errorf("incorrect " + v))
		}
	}
	return nil
}
