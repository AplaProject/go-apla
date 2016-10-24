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

package parser

import (
	"fmt"
	"github.com/EGaaS/go-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

func (p *Parser) GetTxMaps(fields []map[string]string) error {
	log.Debug("p.TxSlice %s", p.TxSlice)

	//log.Debug("p.TxSlice", p.TxSlice)
	p.TxMap = make(map[string][]byte)
	p.TxMaps = new(txMapsType)
	p.TxMaps.Float64 = make(map[string]float64)
	p.TxMaps.Money = make(map[string]float64)
	p.TxMaps.Int64 = make(map[string]int64)
	p.TxMaps.Bytes = make(map[string][]byte)
	p.TxMaps.String = make(map[string]string)
	p.TxMaps.Decimal = make(map[string]decimal.Decimal)
	p.TxMaps.Bytes["hash"] = p.TxSlice[0]
	p.TxMaps.Int64["type"] = utils.BytesToInt64(p.TxSlice[1])
	p.TxMaps.Int64["time"] = utils.BytesToInt64(p.TxSlice[2])
	p.TxMaps.Int64["user_id"] = utils.BytesToInt64(p.TxSlice[3])
	p.TxMaps.Int64["state_id"] = utils.BytesToInt64(p.TxSlice[4])
	p.TxMaps.Int64["_id"] = utils.BytesToInt64(p.TxSlice[4])
	p.TxMap["hash"] = p.TxSlice[0]
	p.TxMap["type"] = p.TxSlice[1]
	p.TxMap["time"] = p.TxSlice[2]
	p.TxMap["user_id"] = p.TxSlice[3]
	p.TxMap["state_id"] = p.TxSlice[4]

	if p.TxMaps.Int64["state_id"] > 0 {
		p.TxStateID = uint32(p.TxMaps.Int64["state_id"])
		p.TxStateIDStr = utils.Int64ToStr(p.TxMaps.Int64["state_id"])
		p.TxMap["citizen_id"] = p.TxMap["user_id"]
		p.TxMaps.Int64["citizen_id"] = p.TxMaps.Int64["user_id"]
		p.TxCitizenID = p.TxMaps.Int64["user_id"]
		p.TxWalletID = 0
		p.TxMap["wallet_id"] = utils.Int64ToByte(0)
		p.TxMaps.Int64["wallet_id"] = 0
	} else {
		p.TxStateID = 0
		p.TxStateIDStr = ""
		p.TxMap["wallet_id"] = p.TxMap["user_id"]
		p.TxMaps.Int64["wallet_id"] = p.TxMaps.Int64["user_id"]
		p.TxWalletID = p.TxMaps.Int64["user_id"]
		p.TxCitizenID = 0
		p.TxMap["citizen_id"] = utils.Int64ToByte(0)
		p.TxMaps.Int64["citizen_id"] = 0
	}

	if p.TxMaps.Int64["type"] == 0 {
		return fmt.Errorf(`p.TxMaps.Int64["type"] == 0`)

	}
	var allFields []map[string]string
	allFields = append(allFields, fields...)
	/*	if  p.TxMaps.Int64["type"] <= int64(len(consts.TxTypes)) && consts.TxTypes[int(p.TxMaps.Int64["type"])] == "new_citizen" {
		// получим набор доп. полей, которые должны быть в данной тр-ии
		additionalFields, err := p.Single(`SELECT fields FROM citizen_fields WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).Bytes()
		if err != nil {
			return p.ErrInfo(err)
		}

		additionalFieldsMap := []map[string]string{}
		err = json.Unmarshal(additionalFields, &additionalFieldsMap)
		if err != nil {
			return p.ErrInfo(err)
		}

		for _, date := range additionalFieldsMap {
			allFields = append(allFields, map[string]string{date["name"]: date["txType"]})
		}
		allFields = append(allFields, map[string]string{"sign": "bytes"})
	}*/
	log.Debug("%v", allFields)
	log.Debug("%d %d", len(allFields), len(p.TxSlice))
	log.Debug("%s", p.TxMap)
	if len(p.TxSlice) != len(allFields)+5 {
		return fmt.Errorf("bad transaction_array %d != %d (type=%d)", len(p.TxSlice), len(allFields)+5, p.TxSlice[0])
	}
	for i := 0; i < len(allFields); i++ {
		for field, fType := range allFields[i] {
			p.TxMap[field] = p.TxSlice[i+5]
			switch fType {
			case "int64":
				p.TxMaps.Int64[field] = utils.BytesToInt64(p.TxSlice[i+5])
			case "float64":
				p.TxMaps.Float64[field] = utils.BytesToFloat64(p.TxSlice[i+5])
			case "money":
				p.TxMaps.Money[field] = utils.StrToMoney(string(p.TxSlice[i+5]))
			case "bytes":
				p.TxMaps.Bytes[field] = p.TxSlice[i+5]
			case "string":
				p.TxMaps.String[field] = string(p.TxSlice[i+5])
			case "decimal":
				dec, err := decimal.NewFromString(string(p.TxSlice[i+5]))
				if err!=nil {
					return err
				}
				p.TxMaps.Decimal[field] = dec
			}
		}
	}
	log.Debug("%s", p.TxMaps)
	p.TxTime = p.TxMaps.Int64["time"]
	p.PublicKeys = nil
	//log.Debug("p.TxMaps", p.TxMaps)
	//log.Debug("p.TxMap", p.TxMap)
	return nil
}
