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

	//	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (c *Controller) SaveQueue() (string, error) {

	var err error
	c.r.ParseForm()

	/*citizenId := utils.BytesToInt64([]byte(c.r.FormValue("citizenId")))
	walletId := utils.BytesToInt64([]byte(c.r.FormValue("walletId")))*/

	citizenId := c.SessCitizenId
	walletId := c.SessWalletId

	log.Debug("citizenId %d / walletId %d ", citizenId, walletId)

	if citizenId <= 0 && walletId == 0 {
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
	//	fmt.Printf("PublicKey %d %x\r\n", lenpub, publicKey)
	txType := utils.TypeInt(txType_)
	sign := make([]byte, 0)
	signature, err := lib.JSSignToBytes(c.r.FormValue("signature1"))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(signature) > 0 {
		sign = append(sign, utils.EncodeLengthPlusData(signature)...)
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

	userId := walletId
	stateId := utils.StrToInt64(c.r.FormValue("stateId"))
	if stateId > 0 {
		userId = citizenId
	}
	/*if stateId == 0 {
		return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
	}*/

	var (
		data []byte
		//		key  []byte
	)
	/*	txHead := consts.TxHeader{Type: uint8(txType), Time: uint32(txTime),
		WalletId: walletId, CitizenId: citizenId}*/
	switch txType_ {
	/*	case "CitizenRequest":
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
			userId := uint64(walletId)
			stateId := uint32(utils.StrToInt64(c.r.FormValue("stateId")))
			TXHead := consts.TXHeader{Type: int32(txType), Time: uint32(txTime),
				WalletId: userId, StateId: int32(stateId), Sign: sign}
			// ---
			if stateId == 0 {
				return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
			}
			if key, err = hex.DecodeString(c.r.FormValue("publicKey")); err == nil {
				_, err = lib.BinMarshal(&data, &consts.TXNewCitizen{TXHeader: TXHead,
					PublicKey: key})
			}*/
	case "DLTTransfer":

		stateId = 0
		walletAddress := []byte(c.r.FormValue("walletAddress"))
		amount := []byte(c.r.FormValue("amount"))
		commission := []byte(c.r.FormValue("commission"))
		comment_ := c.r.FormValue("comment")
		if len(comment_) == 0 {
			comment_ = "null"
		}
		comment := []byte(comment_)
		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(walletAddress)...)
		data = append(data, utils.EncodeLengthPlusData(amount)...)
		data = append(data, utils.EncodeLengthPlusData(commission)...)
		data = append(data, utils.EncodeLengthPlusData(comment)...)
		data = append(data, utils.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "DLTChangeHostVote":

		stateId = 0

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData([]byte(c.r.FormValue("host")))...)
		data = append(data, utils.EncodeLengthPlusData([]byte(c.r.FormValue("addressVote")))...)
		data = append(data, utils.EncodeLengthPlusData([]byte(c.r.FormValue("fuelRate")))...)
		data = append(data, utils.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "NewState":

		stateId = 0
		stateName := []byte(c.r.FormValue("state_name"))
		currencyName := []byte(c.r.FormValue("currency_name"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(stateName)...)
		data = append(data, utils.EncodeLengthPlusData(currencyName)...)
		data = append(data, utils.EncodeLengthPlusData(publicKey)...)
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
		index := []byte(c.r.FormValue("index"))
		colType := []byte(c.r.FormValue("column_type"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(tableName)...)
		data = append(data, utils.EncodeLengthPlusData(columnName)...)
		data = append(data, utils.EncodeLengthPlusData(permissions)...)
		data = append(data, utils.EncodeLengthPlusData(index)...)
		data = append(data, utils.EncodeLengthPlusData(colType)...)
		data = append(data, binSignatures...)

	case "EditColumn":

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

	case "AppendPage":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, binSignatures...)

	case "AppendMenu":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
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

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		menu := []byte(c.r.FormValue("menu"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(menu)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewPage":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		menu := []byte(c.r.FormValue("menu"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(menu)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditTable":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		table_name := []byte(c.r.FormValue("table_name"))
		general_update := []byte(c.r.FormValue("general_update"))
		insert := []byte(c.r.FormValue("insert"))
		new_column := []byte(c.r.FormValue("new_column"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(table_name)...)
		data = append(data, utils.EncodeLengthPlusData(general_update)...)
		data = append(data, utils.EncodeLengthPlusData(insert)...)
		data = append(data, utils.EncodeLengthPlusData(new_column)...)
		data = append(data, binSignatures...)

	case "EditStateParameters":

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

	case "NewStateParameters":

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

	case "NewContract":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
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

		global := []byte(c.r.FormValue("global"))
		id := []byte(c.r.FormValue("id"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(id)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "ActivateContract":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}

		global := []byte(c.r.FormValue("global"))
		id := []byte(c.r.FormValue("id"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(id)...)
		data = append(data, binSignatures...)

	case "NewMenu":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateId > 0 {
			userId = citizenId
		}
		if stateId == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
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

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditWallet":

		userId := walletId
		stateId := utils.StrToInt64(c.r.FormValue("stateId"))
		if userId == 0 {
			userId = citizenId
		}
		wallet_id := []byte(c.r.FormValue("id"))
		spending := []byte(c.r.FormValue("spending_contract"))
		conditions := []byte(c.r.FormValue("conditions_change"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(wallet_id)...)
		data = append(data, utils.EncodeLengthPlusData(spending)...)
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

		global := []byte(c.r.FormValue("global"))
		tableName := []byte(c.r.FormValue("table_name"))
		columns := []byte(c.r.FormValue("columns"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(tableName)...)
		data = append(data, utils.EncodeLengthPlusData(columns)...)
		data = append(data, binSignatures...)

	case "ChangeNodeKeyDLT":

		stateId = 0

		publicKey := []byte(c.r.FormValue("publicKey"))
		privateKey := []byte(c.r.FormValue("privateKey"))

		verifyData := map[string]string{c.r.FormValue("publicKey"): "public_key", c.r.FormValue("privateKey"): "private_key"}
		err := CheckInputData(verifyData)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		myWalletId, err := c.GetMyWalletID()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if myWalletId == walletId {
			err = c.ExecSQL(`INSERT INTO my_node_keys (
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
		}

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(utils.HexToBin(publicKey))...)
		data = append(data, binSignatures...)

	case "EditLang", "NewLang":

		name := []byte(c.r.FormValue("name"))
		trans := []byte(c.r.FormValue("trans"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(trans)...)
		data = append(data, binSignatures...)

	case "EditSign", "NewSign":

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewAccount":

		accountId := utils.StrToInt64(c.r.FormValue("accountId"))
		pubKey, err := hex.DecodeString(c.r.FormValue("pubkey"))
		if accountId == 0 || stateId == 0 || userId == 0 || err != nil {
			return ``, fmt.Errorf(`incorrect NewAccount parameters`)
		}
		encKey := c.r.FormValue("prvkey")
		if len(encKey) == 0 {
			return ``, fmt.Errorf(`incorrect encrypted key`)
		}
		err = c.ExecSQL(fmt.Sprintf(`INSERT INTO "%d_anonyms" (id_citizen, id_anonym, encrypted)
			VALUES (?,?,[hex])`, stateId), userId, accountId, encKey)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(accountId)...)
		data = append(data, utils.EncodeLengthPlusData(stateId)...)

		data = append(data, utils.EncodeLengthPlusData(pubKey)...)
		data = append(data, binSignatures...)

	case "ChangeNodeKey":

		publicKey := []byte(c.r.FormValue("publicKey"))
		privateKey := []byte(c.r.FormValue("privateKey"))

		verifyData := map[string]string{c.r.FormValue("publicKey"): "public_key", c.r.FormValue("privateKey"): "private_key"}
		err := CheckInputData(verifyData)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		myWalletId, err := c.GetMyWalletID()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if myWalletId == walletId {
			err = c.ExecSQL(`INSERT INTO my_node_keys (
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

	err = c.ExecSQL(`INSERT INTO transactions_status (
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

	log.Debug("INSERT INTO queue_tx (hash, data) VALUES (%s, %s)", md5, utils.BinToHex(data))
	err = c.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, utils.BinToHex(data))
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
