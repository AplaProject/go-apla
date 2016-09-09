package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

// общая проверка для всех _front
func (p *Parser) generalCheck() error {
	log.Debug("%s", p.TxMap)
	if !utils.CheckInputData(p.TxMap["wallet_id"], "int64") {
		return utils.ErrInfoFmt("incorrect wallet_id")
	}
	if !utils.CheckInputData(p.TxMap["citizen_id"], "int64") {
		return utils.ErrInfoFmt("incorrect citizen_id")
	}
	if !utils.CheckInputData(p.TxMap["time"], "int") {
		return utils.ErrInfoFmt("incorrect time")
	}

	// проверим, есть ли такой юзер и заодно получим public_key
	if p.TxMaps.Int64["type"] == utils.TypeInt("DLTTransfer") || p.TxMaps.Int64["type"] == utils.TypeInt("DLTChangeHostVote") || p.TxMaps.Int64["type"] == utils.TypeInt("CitizenRequest") {
		data, err := p.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM dlt_wallets WHERE wallet_id = ?", utils.BytesToInt64(p.TxMap["wallet_id"])).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("datausers", data)
		if len(data["public_key_0"]) == 0 {
			if len(p.TxMap["public_key"]) == 0 {
				return utils.ErrInfoFmt("incorrect public_key")
			}
			// возможно юзер послал ключ с тр-ией
			log.Debug("lower(hex(address) %s", string(utils.HashSha1Hex([]byte(p.TxMap["public_key"]))))
			walletId, err := p.Single(`SELECT wallet_id FROM dlt_wallets WHERE address = [hex]`, string(utils.HashSha1Hex([]byte(p.TxMap["public_key"])))).Int64()
			if err != nil {
				return utils.ErrInfo(err)
			}
			if walletId == 0 {
				return utils.ErrInfoFmt("incorrect wallet_id or public_key")
			}
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key"]))
		}
		p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_0"]))
		if len(data["public_key_1"]) > 10 {
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_1"]))
		}
		if len(data["public_key_2"]) > 10 {
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_2"]))
		}
	} else {
		data, err := p.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM citizens WHERE citizen_id = ?", utils.BytesToInt64(p.TxMap["citizen_id"])).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("datausers", data)
		if len(data["public_key_0"]) == 0 {
			return utils.ErrInfoFmt("incorrect user_id")
		}
		p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_0"]))
		if len(data["public_key_1"]) > 10 {
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_1"]))
		}
		if len(data["public_key_2"]) > 10 {
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_2"]))
		}
	}
	// чтобы не записали слишком длинную подпись
	// 128 - это нод-ключ
	if len(p.TxMap["sign"]) < 64 || len(p.TxMap["sign"]) > 5120 {
		return utils.ErrInfoFmt("incorrect sign size %d", len(p.TxMap["sign"]))
	}
	return nil
}