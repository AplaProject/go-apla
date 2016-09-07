package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"encoding/json"
)

func (p *Parser) GetTxMaps(fields []map[string]string) error {
	log.Debug("p.TxSlice %s", p.TxSlice)
	if len(p.TxSlice) != len(fields)+5 {
		return fmt.Errorf("bad transaction_array %d != %d (type=%d)", len(p.TxSlice), len(fields)+4, p.TxSlice[0])
	}
	//log.Debug("p.TxSlice", p.TxSlice)
	p.TxMap = make(map[string][]byte)
	p.TxMaps = new(txMapsType)
	p.TxMaps.Float64 = make(map[string]float64)
	p.TxMaps.Money = make(map[string]float64)
	p.TxMaps.Int64 = make(map[string]int64)
	p.TxMaps.Bytes = make(map[string][]byte)
	p.TxMaps.String = make(map[string]string)
	p.TxMaps.Bytes["hash"] = p.TxSlice[0]
	p.TxMaps.Int64["type"] = utils.BytesToInt64(p.TxSlice[1])
	p.TxMaps.Int64["time"] = utils.BytesToInt64(p.TxSlice[2])
	p.TxMaps.Int64["wallet_id"] = utils.BytesToInt64(p.TxSlice[3])
	p.TxMaps.Int64["citizen_id"] = utils.BytesToInt64(p.TxSlice[4])
	p.TxMaps.Int64["_id"] = utils.BytesToInt64(p.TxSlice[4])
	p.TxMap["hash"] = p.TxSlice[0]
	p.TxMap["type"] = p.TxSlice[1]
	p.TxMap["time"] = p.TxSlice[2]
	p.TxMap["wallet_id"] = p.TxSlice[3]
	p.TxMap["citizen_id"] = p.TxSlice[4]

	if p.TxMaps.Int64["type"] == 0 {
		return fmt.Errorf(`p.TxMaps.Int64["type"] == 0`)

	}
	if  p.TxMaps.Int64["type"] <= int64(len(consts.TxTypes)) && consts.TxTypes[int(p.TxMaps.Int64["type"])] == "new_citizen" {
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

		fields = []map[string]string{}
		for _, date := range additionalFieldsMap {
			fields = append(fields, map[string]string{date["name"]: date["txType"]})
		}
		fields = append(fields, map[string]string{"sign": "bytes"})
	}

	for i := 0; i < len(fields); i++ {
		for field, fType := range fields[i] {
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
			}
		}
	}
	log.Debug("%s", p.TxMaps)
	p.TxCitizenID = p.TxMaps.Int64["citizen_id"]
	p.TxWalletID = p.TxMaps.Int64["wallet_id"]
	p.TxTime = p.TxMaps.Int64["time"]
	p.PublicKeys = nil
	//log.Debug("p.TxMaps", p.TxMaps)
	//log.Debug("p.TxMap", p.TxMap)
	return nil
}