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
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"fmt"
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

// общая проверка для всех _front
func (p *Parser) generalCheckStruct(moreSign string) error {
//	head := reflect.ValueOf(p.TxPtr).Elem().Field(0).Interface().(consts.TxHeader)
	head := consts.Header(p.TxPtr)
	fmt.Println(`General`, head)
	// проверим, есть ли такой юзер и заодно получим public_key
	if int64(head.Type) == utils.TypeInt("DLTTransfer") || int64(head.Type) == utils.TypeInt("DLTChangeHostVote") || int64(head.Type) == utils.TypeInt("CitizenRequest") {
		data, err := p.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM dlt_wallets WHERE wallet_id = ?", head.WalletId ).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
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
		data, err := p.OneRow("SELECT public_key_0, public_key_1, public_key_2 FROM citizens WHERE citizen_id = ?", head.CitizenId ).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
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
	forSign := fmt.Sprintf("%d,%d,%d", head.Type, head.Time, head.WalletId) + moreSign
	//fmt.Println(`forSign`, forSign)
	//fmt.Printf("PublicKeys %x \r\n", p.PublicKeys)
	//fmt.Printf("Sign %x \r\n", data.Sign)
	sign := consts.Sign(p.TxPtr)
	if len(sign) == 0 {
		return p.ErrInfo("empty sign")
	}
	checkSignResult, err := utils.CheckSign(p.PublicKeys, forSign, sign, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !checkSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}