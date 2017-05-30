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

// SaveQueue write a transaction in the queue of transactions
func (c *Controller) SaveQueue() (string, error) {

	var err error
	c.r.ParseForm()

	/*citizenID := utils.BytesToInt64([]byte(c.r.FormValue("citizenId")))
	walletID := utils.BytesToInt64([]byte(c.r.FormValue("walletId")))*/

	citizenID := c.SessCitizenID
	walletID := c.SessWalletID

	log.Debug("citizenID %d / walletID %d ", citizenID, walletID)

	if citizenID <= 0 && walletID == 0 {
		return `{"result":"incorrect citizenID || walletID"}`, nil
	}

	txTime := utils.StrToInt64(c.r.FormValue("time"))
	if !utils.CheckInputData(txTime, "int") {
		return `{"result":"incorrect time"}`, nil
	}
	itxType := c.r.FormValue("type")
	if !utils.CheckInputData(itxType, "type") {
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
	txType := utils.TypeInt(itxType)
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
	log.Debug("itxType", itxType)
	log.Debug("txType", txType)

	userID := walletID
	stateID := utils.StrToInt64(c.r.FormValue("stateId"))
	if stateID > 0 {
		userID = citizenID
	}
	/*if stateID == 0 {
		return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
	}*/

	var (
		data []byte
		//		key  []byte
	)
	/*	txHead := consts.TxHeader{Type: uint8(txType), Time: uint32(txTime),
		WalletId: walletID, CitizenId: citizenID}*/
	switch itxType {
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
			userID := uint64(walletID)
			stateID := uint32(utils.StrToInt64(c.r.FormValue("stateId")))
			TXHead := consts.TXHeader{Type: int32(txType), Time: uint32(txTime),
				WalletId: userID, StateId: int32(stateID), Sign: sign}
			// ---
			if stateID == 0 {
				return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
			}
			if key, err = hex.DecodeString(c.r.FormValue("publicKey")); err == nil {
				_, err = lib.BinMarshal(&data, &consts.TXNewCitizen{TXHeader: TXHead,
					PublicKey: key})
			}*/
	case "DLTTransfer":

		stateID = 0
		walletAddress := []byte(c.r.FormValue("walletAddress"))
		amount := []byte(c.r.FormValue("amount"))
		commission := []byte(c.r.FormValue("commission"))
		vcomment := c.r.FormValue("comment")
		if len(vcomment) == 0 {
			vcomment = "null"
		}
		comment := []byte(vcomment)
		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(walletAddress)...)
		data = append(data, utils.EncodeLengthPlusData(amount)...)
		data = append(data, utils.EncodeLengthPlusData(commission)...)
		data = append(data, utils.EncodeLengthPlusData(comment)...)
		data = append(data, utils.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "DLTChangeHostVote":

		stateID = 0

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData([]byte(c.r.FormValue("host")))...)
		data = append(data, utils.EncodeLengthPlusData([]byte(c.r.FormValue("addressVote")))...)
		data = append(data, utils.EncodeLengthPlusData([]byte(c.r.FormValue("fuelRate")))...)
		data = append(data, utils.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "NewState":

		stateID = 0
		stateName := []byte(c.r.FormValue("state_name"))
		currencyName := []byte(c.r.FormValue("currency_name"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(stateName)...)
		data = append(data, utils.EncodeLengthPlusData(currencyName)...)
		data = append(data, utils.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "NewColumn":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		tableName := []byte(c.r.FormValue("table_name"))
		columnName := []byte(c.r.FormValue("column_name"))
		permissions := []byte(c.r.FormValue("permissions"))
		index := []byte(c.r.FormValue("index"))
		colType := []byte(c.r.FormValue("column_type"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
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
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(tableName)...)
		data = append(data, utils.EncodeLengthPlusData(columnName)...)
		data = append(data, utils.EncodeLengthPlusData(permissions)...)
		data = append(data, binSignatures...)

	case "AppendPage":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, binSignatures...)

	case "AppendMenu":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, binSignatures...)

	case "EditPage":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		menu := []byte(c.r.FormValue("menu"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(menu)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewPage":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		menu := []byte(c.r.FormValue("menu"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(menu)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditTable":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		tableName := []byte(c.r.FormValue("table_name"))
		generalUpdate := []byte(c.r.FormValue("general_update"))
		insert := []byte(c.r.FormValue("insert"))
		newColumn := []byte(c.r.FormValue("new_column"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(tableName)...)
		data = append(data, utils.EncodeLengthPlusData(generalUpdate)...)
		data = append(data, utils.EncodeLengthPlusData(insert)...)
		data = append(data, utils.EncodeLengthPlusData(newColumn)...)
		data = append(data, binSignatures...)

	case "EditStateParameters":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewStateParameters":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewContract":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditContract":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		id := []byte(c.r.FormValue("id"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(id)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "ActivateContract":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}

		global := []byte(c.r.FormValue("global"))
		id := []byte(c.r.FormValue("id"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(id)...)
		data = append(data, binSignatures...)

	case "NewMenu":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditMenu":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditWallet":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if userID == 0 {
			userID = citizenID
		}
		walletID := []byte(c.r.FormValue("id"))
		spending := []byte(c.r.FormValue("spending_contract"))
		conditions := []byte(c.r.FormValue("conditions_change"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(walletID)...)
		data = append(data, utils.EncodeLengthPlusData(spending)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewTable":

		userID := walletID
		stateID := utils.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		tableName := []byte(c.r.FormValue("table_name"))
		columns := []byte(c.r.FormValue("columns"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(tableName)...)
		data = append(data, utils.EncodeLengthPlusData(columns)...)
		data = append(data, binSignatures...)

	case "ChangeNodeKeyDLT":

		stateID = 0

		publicKey := []byte(c.r.FormValue("publicKey"))
		privateKey := []byte(c.r.FormValue("privateKey"))

		verifyData := map[string]string{c.r.FormValue("publicKey"): "public_key", c.r.FormValue("privateKey"): "private_key"}
		err := CheckInputData(verifyData)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		myWalletID, err := c.GetMyWalletID()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if myWalletID == walletID {
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
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(utils.HexToBin(publicKey))...)
		data = append(data, binSignatures...)

	case "EditLang", "NewLang":

		name := []byte(c.r.FormValue("name"))
		trans := []byte(c.r.FormValue("trans"))

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
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
		data = append(data, utils.EncodeLengthPlusData(userID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)
		data = append(data, utils.EncodeLengthPlusData(global)...)
		data = append(data, utils.EncodeLengthPlusData(name)...)
		data = append(data, utils.EncodeLengthPlusData(value)...)
		data = append(data, utils.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewAccount":

		accountID := utils.StrToInt64(c.r.FormValue("accountId"))
		pubKey, err := hex.DecodeString(c.r.FormValue("pubkey"))
		if accountID == 0 || stateID == 0 || userID == 0 || err != nil {
			return ``, fmt.Errorf(`incorrect NewAccount parameters`)
		}
		encKey := c.r.FormValue("prvkey")
		if len(encKey) == 0 {
			return ``, fmt.Errorf(`incorrect encrypted key`)
		}
		err = c.ExecSQL(fmt.Sprintf(`INSERT INTO "%d_anonyms" (id_citizen, id_anonym, encrypted)
			VALUES (?,?,[hex])`, stateID), userID, accountID, encKey)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		data = utils.DecToBin(txType, 1)
		data = append(data, utils.DecToBin(txTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(accountID)...)
		data = append(data, utils.EncodeLengthPlusData(stateID)...)

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
		myWalletID, err := c.GetMyWalletID()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if myWalletID == walletID {
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
		data = append(data, utils.EncodeLengthPlusData(walletID)...)
		data = append(data, utils.EncodeLengthPlusData(citizenID)...)
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
			)`, md5, time.Now().Unix(), txType, walletID, citizenID)
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

// CheckInputData calls utils.CheckInputData for the each item of the map
func CheckInputData(data map[string]string) error {
	for k, v := range data {
		if !utils.CheckInputData(k, v) {
			return utils.ErrInfo(fmt.Errorf("incorrect " + v))
		}
	}
	return nil
}
