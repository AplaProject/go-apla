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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
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

	txTime := converter.StrToInt64(c.r.FormValue("time"))
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
	signature, err := crypto.JSSignToBytes(c.r.FormValue("signature1"))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(signature) > 0 {
		sign = append(sign, converter.EncodeLengthPlusData(signature)...)
	}
	if len(sign) == 0 {
		return `{"result":"signature is empty"}`, nil
	}
	fmt.Printf("Len sign %d\r\n", len(sign))
	binSignatures := converter.EncodeLengthPlusData(sign)

	log.Debug("binSignatures %x", binSignatures)
	log.Debug("binSignatures %s", binSignatures)
	log.Debug("itxType", itxType)
	log.Debug("txType", txType)

	userID := walletID
	stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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
		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(walletAddress)...)
		data = append(data, converter.EncodeLengthPlusData(amount)...)
		data = append(data, converter.EncodeLengthPlusData(commission)...)
		data = append(data, converter.EncodeLengthPlusData(comment)...)
		data = append(data, converter.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "DLTChangeHostVote":

		stateID = 0

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData([]byte(c.r.FormValue("host")))...)
		data = append(data, converter.EncodeLengthPlusData([]byte(c.r.FormValue("addressVote")))...)
		data = append(data, converter.EncodeLengthPlusData([]byte(c.r.FormValue("fuelRate")))...)
		data = append(data, converter.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "NewState":

		stateID = 0
		stateName := []byte(c.r.FormValue("state_name"))
		currencyName := []byte(c.r.FormValue("currency_name"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(stateName)...)
		data = append(data, converter.EncodeLengthPlusData(currencyName)...)
		data = append(data, converter.EncodeLengthPlusData(publicKey)...)
		data = append(data, binSignatures...)

	case "NewColumn":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(tableName)...)
		data = append(data, converter.EncodeLengthPlusData(columnName)...)
		data = append(data, converter.EncodeLengthPlusData(permissions)...)
		data = append(data, converter.EncodeLengthPlusData(index)...)
		data = append(data, converter.EncodeLengthPlusData(colType)...)
		data = append(data, binSignatures...)

	case "EditColumn":

		tableName := []byte(c.r.FormValue("table_name"))
		columnName := []byte(c.r.FormValue("column_name"))
		permissions := []byte(c.r.FormValue("permissions"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(tableName)...)
		data = append(data, converter.EncodeLengthPlusData(columnName)...)
		data = append(data, converter.EncodeLengthPlusData(permissions)...)
		data = append(data, binSignatures...)

	case "AppendPage":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, binSignatures...)

	case "AppendMenu":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, binSignatures...)

	case "EditPage":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(menu)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewPage":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(menu)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditTable":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(tableName)...)
		data = append(data, converter.EncodeLengthPlusData(generalUpdate)...)
		data = append(data, converter.EncodeLengthPlusData(insert)...)
		data = append(data, converter.EncodeLengthPlusData(newColumn)...)
		data = append(data, binSignatures...)

	case "EditStateParameters":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewStateParameters":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewContract":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditContract":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(id)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "ActivateContract":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}

		global := []byte(c.r.FormValue("global"))
		id := []byte(c.r.FormValue("id"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(id)...)
		data = append(data, binSignatures...)

	case "NewMenu":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditMenu":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "EditWallet":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
		if userID == 0 {
			userID = citizenID
		}
		walletID := []byte(c.r.FormValue("id"))
		spending := []byte(c.r.FormValue("spending_contract"))
		conditions := []byte(c.r.FormValue("conditions_change"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(walletID)...)
		data = append(data, converter.EncodeLengthPlusData(spending)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewTable":

		userID := walletID
		stateID := converter.StrToInt64(c.r.FormValue("stateId"))
		if stateID > 0 {
			userID = citizenID
		}
		if stateID == 0 {
			return "", utils.ErrInfo(fmt.Errorf(`StateId is not defined`))
		}

		global := []byte(c.r.FormValue("global"))
		tableName := []byte(c.r.FormValue("table_name"))
		columns := []byte(c.r.FormValue("columns"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(tableName)...)
		data = append(data, converter.EncodeLengthPlusData(columns)...)
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(converter.HexToBin(publicKey))...)
		data = append(data, binSignatures...)

	case "EditLang", "NewLang":

		name := []byte(c.r.FormValue("name"))
		trans := []byte(c.r.FormValue("trans"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(trans)...)
		data = append(data, binSignatures...)

	case "EditSign", "NewSign":

		global := []byte(c.r.FormValue("global"))
		name := []byte(c.r.FormValue("name"))
		value := []byte(c.r.FormValue("value"))
		conditions := []byte(c.r.FormValue("conditions"))

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(userID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)
		data = append(data, converter.EncodeLengthPlusData(global)...)
		data = append(data, converter.EncodeLengthPlusData(name)...)
		data = append(data, converter.EncodeLengthPlusData(value)...)
		data = append(data, converter.EncodeLengthPlusData(conditions)...)
		data = append(data, binSignatures...)

	case "NewAccount":

		accountID := converter.StrToInt64(c.r.FormValue("accountId"))
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(accountID)...)
		data = append(data, converter.EncodeLengthPlusData(stateID)...)

		data = append(data, converter.EncodeLengthPlusData(pubKey)...)
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

		data = converter.DecToBin(txType, 1)
		data = append(data, converter.DecToBin(txTime, 4)...)
		data = append(data, converter.EncodeLengthPlusData(walletID)...)
		data = append(data, converter.EncodeLengthPlusData(citizenID)...)
		data = append(data, converter.EncodeLengthPlusData(converter.HexToBin(publicKey))...)
		data = append(data, binSignatures...)
	}

	if err != nil {
		return "", utils.ErrInfo(err)
	}

	hash, err := crypto.Hash(data)
	fmt.Println("hash", hash)
	if err != nil {
		log.Fatal(err)
	}
	hash = converter.BinToHex(hash)
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
			)`, hash, time.Now().Unix(), txType, walletID, citizenID)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	fmt.Println("data", converter.BinToHex(data))
	fmt.Println("str ", string(hash))

	log.Debug("INSERT INTO queue_tx (hash, data) VALUES (%s, %s)", hash, converter.BinToHex(data))
	err = c.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", hash, converter.BinToHex(data))
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return `{"hash":"` + string(hash) + `"}`, nil
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
