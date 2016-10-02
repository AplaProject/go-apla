// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package controllers

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/utils"
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

	publicKey, _ := hex.DecodeString(c.r.FormValue("pubkey"))
	lenpub := len(publicKey)
	if lenpub > 64 {
		publicKey = publicKey[lenpub-64:]
	} else if lenpub == 0 {
		publicKey = []byte("null")
	}
	fmt.Printf("PublicKey %d %x\r\n", lenpub, publicKey)
	txType := utils.TypeInt(txType_)
	sign := make([]byte, 0)
	for i := 1; i <= 3; i++ {
		log.Debug("signature %s", c.r.FormValue(fmt.Sprintf("signature%d", i)))
		signature := utils.ConvertJSSign(c.r.FormValue(fmt.Sprintf("signature%d", i)))
		log.Debug("signature %s", signature)
		if i == 1 || len(signature) > 0 {
			bsign, _ := hex.DecodeString(signature)
			log.Debug("bsign %s", bsign)
			log.Debug("bsign %x", bsign)
			sign = append(sign, utils.EncodeLengthPlusData(bsign)...)
			log.Debug("sign %x", sign)
		}
	}
	if len(sign) == 0 {
		return `{"result":"signature is empty"}`, nil
	}
	fmt.Printf("Len sign %d\r\n", len(sign))
	binSignatures := utils.EncodeLengthPlusData(sign)

	log.Debug("binSignatures %x", binSignatures)
	log.Debug("binSignatures %s", binSignatures)
	log.Debug("txType_", txType_)
	log.Debug("txType", txType)

	var (
		data []byte
		key  []byte
	)
	txHead := consts.TxHeader{Type: uint8(txType), Time: uint32(txTime),
		WalletId: walletId, CitizenId: citizenId}
	switch txType_ {
	case "CitizenRequest":
		_, err = lib.BinMarshal(&data, &consts.CitizenRequest{TxHeader: txHead,
			StateId: utils.StrToInt64(c.r.FormValue("stateId")), Sign: sign})
	case "NewCitizen":
		if key, err = hex.DecodeString(c.r.FormValue("publicKey")); err == nil {
			_, err = lib.BinMarshal(&data, &consts.NewCitizen{TxHeader: txHead,
				StateId:   utils.StrToInt64(c.r.FormValue("stateId")),
				PublicKey: key, Sign: sign})
		}
	case "TXNewCitizen":
		// This will be common part
		userId := walletId
		stateId := uint32(utils.StrToInt64(c.r.FormValue("stateId")))
		if stateId > 0 {
			userId = citizenId
		}
		TXHead := consts.TXHeader{Type: uint32(txType), Time: uint32(txTime),
			UserId: userId, StateId: stateId, Sign: sign}
		// ---
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}
		if key, err = hex.DecodeString(c.r.FormValue("publicKey")); err == nil {
			_, err = lib.BinMarshal(&data, &consts.TXNewCitizen{TXHeader: TXHead,
				PublicKey: key})
		}
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

	case "NewState":

		stateName := []byte(c.r.FormValue("state_name"))
		currencyName := []byte(c.r.FormValue("currency_name"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(walletId)...)
		data = append(data, utils.EncodeLengthPlusData(citizenId)...)
		data = append(data, utils.EncodeLengthPlusData(stateName)...)
		data = append(data, utils.EncodeLengthPlusData(currencyName)...)
		data = append(data, binSignatures...)

	case "NewColumn":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		tableName := []byte(c.r.FormValue("table_name"))
		columnName := []byte(c.r.FormValue("column_name"))
		permissions := []byte(c.r.FormValue("permissions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(tableName)...)
		data = append(data, utils.EncodeLengthPlusData(columnName)...)
		data = append(data, utils.EncodeLengthPlusData(permissions)...)
		data = append(data, binSignatures...)


	case "EditPage":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		menu := []byte(c.r.FormValue("menu"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(menu)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)


	case "NewContract":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditContract":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		id := []byte(c.r.FormValue("id"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(id)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditMenu":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewTable":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		tableName := []byte(c.r.FormValue("table_name"))
		columns := []byte(c.r.FormValue("columns"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(tableName)...)
		data = append(data, utils.EncodeLengthPlusData(columns)...)
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
	if err != nil {
		return "", utils.ErrInfo(err)
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

	return `{"hash":"` + string(md5) + `"}`, nil
}

func CheckInputData(data map[string]string) error {
	for k, v := range data {
		if !utils.CheckInputData(k, v) {
			return utils.ErrInfo(fmt.Errorf("incorrect " + v))
		}
	}
	return nil
}
