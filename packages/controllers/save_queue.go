package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
	"encoding/hex"
	"fmt"
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

	publicKey,_ := hex.DecodeString(c.r.FormValue("pubkey"))
	lenpub := len(publicKey)
	if lenpub>64 {
		publicKey = publicKey[lenpub-64:]
	} else if lenpub == 0 {
		publicKey = []byte("null")
	}
	fmt.Printf("PublicKey %d %x\r\n", lenpub, publicKey)
	txType := utils.TypeInt(txType_)
	sign := make([]byte, 0)
	for i := 1; i<=3; i++ {
		log.Debug("signature %s", c.r.FormValue( fmt.Sprintf("signature%d", i )))
		signature := utils.ConvertJSSign(c.r.FormValue( fmt.Sprintf("signature%d", i )))
		log.Debug("signature %s", signature)
		if i==1 || len(signature) > 0 {
			bsign,_ := hex.DecodeString(signature)
			log.Debug("bsign %s", bsign)
			log.Debug("bsign %x", bsign)
			sign = append(sign, utils.EncodeLengthPlusData(bsign)...)
		}
	}
	binSignatures := utils.EncodeLengthPlusData(sign)

	log.Debug("binSignatures %x", binSignatures)
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
		data = append(data, utils.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "DLTChangeHostVote":

		host := []byte(c.r.FormValue("host"))
		addressVote := []byte(c.r.FormValue("addressVote"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(walletId)...)
		data = append(data, utils.EncodeLengthPlusData(citizenId)...)
		data = append(data, utils.EncodeLengthPlusData(host)...)
		data = append(data, utils.EncodeLengthPlusData(addressVote)...)
		data = append(data, utils.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "ChangeNodeKey":


		publicKey := []byte(c.r.FormValue("publicKey"))
		privateKey := []byte(c.r.FormValue("privateKey"))

		verifyData := map[string]string{c.r.FormValue("publicKey"): "public_key", c.r.FormValue("privateKey"): "private_key"}
		err := CheckInputData(verifyData)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		err = c.ExecSql(`INSERT INTO my_node_keys (
									public_key,
									private_key
								)
								VALUES (
									[hex],
									?
								)`, publicKey, privateKey)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(walletId)...)
		data = append(data, utils.EncodeLengthPlusData(citizenId)...)
		data = append(data, utils.EncodeLengthPlusData(utils.HexToBin(publicKey))...)
		data = append(data, binSignatures...)

	}
	md5 := utils.Md5(data)

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