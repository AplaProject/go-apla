package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) UpdBlockInfo() {

	blockId := p.BlockData.BlockId
	// для локальных тестов
	if p.BlockData.BlockId == 1 {
		if *utils.StartBlockId != 0 {
			blockId = *utils.StartBlockId
		}
	}
	forSha := fmt.Sprintf("%d,%s,%s,%d,%d,%d", blockId, p.PrevBlock.Hash, p.MrklRoot, p.BlockData.Time, p.BlockData.WalletId, p.BlockData.CBID)
	log.Debug("forSha", forSha)
	p.BlockData.Hash = utils.DSha256(forSha)
	log.Debug("%v", p.BlockData.Hash)
	log.Debug("%v", blockId)
	log.Debug("%v", p.BlockData.Time)
	log.Debug("%v", p.CurrentVersion)

	if p.BlockData.BlockId == 1 {
		err := p.ExecSql("INSERT INTO info_block (hash, block_id, time, cb_id, wallet_id, current_version) VALUES ([hex], ?, ?, ?, ?, ?)",
			p.BlockData.Hash, blockId, p.BlockData.Time, p.BlockData.CBID, p.BlockData.WalletId, p.CurrentVersion)
		if err!=nil {
			log.Error("%v", err)
		}
	} else {
		err := p.ExecSql("UPDATE info_block SET hash = [hex], block_id = ?, time = ?, cb_id = ?, wallet_id = ?, sent = 0",
			p.BlockData.Hash, blockId, p.BlockData.Time, p.BlockData.CBID, p.BlockData.WalletId)
		if err!=nil {
			log.Error("%v", err)
		}
		err = p.ExecSql("UPDATE config SET my_block_id = ? WHERE my_block_id < ?", blockId, blockId)
		if err!=nil {
			log.Error("%v", err)
		}
	}
}
