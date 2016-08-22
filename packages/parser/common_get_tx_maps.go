package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
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