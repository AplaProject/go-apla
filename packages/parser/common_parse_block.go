package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ParseBlock() error {
	/*
		Заголовок
		TYPE (0-блок, 1-тр-я)     1
		BLOCK_ID   				       4
		TIME       					       4
		WALLET_ID                         1-8
		CB_ID                         1
		SIGN                               от 128 до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, WALLET_ID, CB_ID, MRKL_ROOT
		Далее - тело блока (Тр-ии)
	*/
	p.BlockData = utils.ParseBlockHeader(&p.BinaryData)
	log.Debug("%v", p.BlockData)

	p.CurrentBlockId = p.BlockData.BlockId

	return nil
}