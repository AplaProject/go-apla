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
	"time"

	"github.com/EGaaS/go-mvp/packages/consts"
	"github.com/EGaaS/go-mvp/packages/utils"
)

func (p *Parser) CheckBlockHeader() error {
	var err error
	// инфа о предыдущем блоке (т.е. последнем занесенном).
	// в GetBlocks p.PrevBlock определяется снаружи, поэтому тут важно не перезаписать данными из block_chain
	if p.PrevBlock == nil || p.PrevBlock.BlockId != p.BlockData.BlockId-1 {
		p.PrevBlock, err = p.GetBlockDataFromBlockChain(p.BlockData.BlockId - 1)
		log.Debug("PrevBlock 0", p.PrevBlock)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	log.Debug("PrevBlock.BlockId: %v / PrevBlock.Time: %v / PrevBlock.WalletId: %v / PrevBlock.CBID: %v / PrevBlock.Sign: %v", p.PrevBlock.BlockId, p.PrevBlock.Time, p.PrevBlock.WalletId, p.PrevBlock.CBID, p.PrevBlock.Sign)

	log.Debug("p.PrevBlock.BlockId", p.PrevBlock.BlockId)
	// для локальных тестов
	if p.PrevBlock.BlockId == 1 {
		if *utils.StartBlockId != 0 {
			p.PrevBlock.BlockId = *utils.StartBlockId
		}
	}

	var first bool
	if p.BlockData.BlockId == 1 {
		first = true
	} else {
		first = false
	}
	log.Debug("%v", first)

	// меркель рут нужен для проверки подписи блока, а также проверки лимитов MAX_TX_SIZE и MAX_TX_COUNT
	//log.Debug("p.Variables: %v", p.Variables)
	p.MrklRoot, err = utils.GetMrklroot(p.BinaryData, first)
	log.Debug("p.MrklRoot: %s", p.MrklRoot)
	if err != nil {
		return utils.ErrInfo(err)
	}

	// проверим время
	if !utils.CheckInputData(p.BlockData.Time, "int") {
		log.Debug("p.BlockData.Time", p.BlockData.Time)
		return utils.ErrInfo(fmt.Errorf("incorrect time"))
	}

	// не слишком ли рано прислан этот блок. допустима погрешность = error_time
	if !first {

		sleepTime, err := p.GetSleepTime(p.BlockData.WalletId, p.BlockData.CBID, p.PrevBlock.CBID, p.PrevBlock.WalletId)
		if err != nil {
			return utils.ErrInfo(err)
		}

		log.Debug("p.PrevBlock.Time %v + sleepTime %v - p.BlockData.Time %v > consts.ERROR_TIME %v", p.PrevBlock.Time, sleepTime, p.BlockData.Time, consts.ERROR_TIME)
		if p.PrevBlock.Time+sleepTime-p.BlockData.Time > consts.ERROR_TIME {
			return utils.ErrInfo(fmt.Errorf("incorrect block time %d + %d - %d > %d", p.PrevBlock.Time, consts.GAPS_BETWEEN_BLOCKS, p.BlockData.Time, consts.ERROR_TIME))
		}
	}

	// исключим тех, кто сгенерил блок с бегущими часами
	if p.BlockData.Time > time.Now().Unix() {
		utils.ErrInfo(fmt.Errorf("incorrect block time"))
	}

	// проверим ID блока
	if !utils.CheckInputData(p.BlockData.BlockId, "int") {
		return utils.ErrInfo(fmt.Errorf("incorrect block_id"))
	}

	// проверим, верный ли ID блока
	if !first {
		if p.BlockData.BlockId != p.PrevBlock.BlockId+1 {
			return utils.ErrInfo(fmt.Errorf("incorrect block_id %d != %d +1", p.BlockData.BlockId, p.PrevBlock.BlockId))
		}
	}
	// проверим, есть ли такой майнер и заодно получим public_key
	nodePublicKey, err := p.GetNodePublicKeyWalletOrCB(p.BlockData.WalletId, p.BlockData.CBID)
	if err != nil {
		return utils.ErrInfo(err)
	}
	if !first {
		if len(nodePublicKey) == 0 {
			return utils.ErrInfo(fmt.Errorf("empty nodePublicKey"))
		}
		// SIGN от 128 байта до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
		forSign := fmt.Sprintf("0,%d,%s,%d,%d,%d,%s", p.BlockData.BlockId, p.PrevBlock.Hash, p.BlockData.Time, p.BlockData.WalletId, p.BlockData.CBID, p.MrklRoot)
		log.Debug(forSign)
		// проверим подпись
		resultCheckSign, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.BlockData.Sign, true)
		if err != nil {
			return utils.ErrInfo(fmt.Errorf("err: %v / p.PrevBlock.BlockId: %d", err, p.PrevBlock.BlockId))
		}
		if !resultCheckSign {
			return utils.ErrInfo(fmt.Errorf("incorrect signature / p.PrevBlock.BlockId: %d", p.PrevBlock.BlockId))
		}
	}
	return nil
}
