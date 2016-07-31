package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type BlockExplorerPage struct {
	Lang           map[string]string
	UserId         int64
	Data           string
	MyNotice       map[string]string
	BlockId        int64
	PoolAdmin      bool
	SessRestricted int64
	Start          int64
	CurrencyList   map[int64]string
}

func (c *Controller) BlockExplorer() (string, error) {

	var err error

	blockId := int64(utils.StrToFloat64(c.Parameters["blockId"]))
	start := int64(utils.StrToFloat64(c.Parameters["start"]))

	var data, sql string
	if start > 0 || (start == 0 && blockId == 0) {
		if start == 0 && blockId == 0 {
			data += "<h3>Latest Blocks</h3>"
			sql = `	SELECT data,  hash
						FROM block_chain
						ORDER BY id DESC
						LIMIT 15`
		} else {
			sql = `	SELECT data,  hash
						FROM block_chain
						ORDER BY id ASC
						LIMIT ` + utils.Int64ToStr(start-1) + `, 100`
		}
		data += `<table class="table"><tr><th>Block</th><th>Hash</th><th>Time</th><th><nobr>User id</nobr></th><th><nobr>Miner id</nobr></th><th>Level</th><th>Transactions</th></tr>`
		blocksChain, err := c.GetAll(sql, -1)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for _, blockData := range blocksChain {
			hash := utils.BinToHex([]byte(blockData["hash"]))
			binaryData := []byte(blockData["data"])
			parser := new(dcparser.Parser)
			parser.DCDB = c.DCDB
			parser.BinaryData = binaryData
			err = parser.ParseDataLite()
			parser.BlockData.Sign = utils.BinToHex(parser.BlockData.Sign)
			minerId, err := c.GetMinerId(parser.BlockData.UserId)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			data += fmt.Sprintf(`<tr><td><a href="#" onclick="dc_navigate('blockExplorer', {'blockId':%d})">%d</a></td><td>%s</td><td><nobr><span class='unixtime'>%d</span></nobr></td><td>%d</td><td>%d</td><td>%d</td><td>`, parser.BlockData.BlockId, parser.BlockData.BlockId, hash, parser.BlockData.Time, parser.BlockData.UserId, minerId, parser.BlockData.Level)
			data += utils.IntToStr(len(parser.TxMapArr))
			data += "</td></tr>"
		}
		data += "</table>"
	} else if blockId > 0 {
		data += `<table class="table">`
		blockChain, err := c.OneRow("SELECT data, hash, cur_0l_miner_id, max_miner_id FROM block_chain WHERE id = ?", blockId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		binToHexArray := []string{"sign", "public_key", "encrypted_message", "comment", "bin_public_keys"}
		hash := utils.BinToHex([]byte(blockChain["hash"]))
		binaryData := blockChain["data"]
		parser := new(dcparser.Parser)
		parser.DCDB = c.DCDB
		parser.BinaryData = []byte(binaryData)
		err = parser.ParseDataLite()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		parser.BlockData.Sign = utils.BinToHex(parser.BlockData.Sign)
		previous := parser.BlockData.BlockId - 1
		next := parser.BlockData.BlockId + 1
		levelsRange := utils.GetBlockGeneratorMinerIdRange(utils.StrToInt64(blockChain["cur_0l_miner_id"]), utils.StrToInt64(blockChain["max_miner_id"]))
		minerId, err := c.GetMinerId(parser.BlockData.UserId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		_, _, _, CurrentUserId, _, _, _ := c.TestBlock()
		maxMinerId, err := c.Single("SELECT max(miner_id) FROM miners").Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		currentMinerId, err := c.Single("SELECT miner_id FROM miners_data WHERE user_id = ?", CurrentUserId).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		NextBlockLevelsRange := utils.GetBlockGeneratorMinerIdRange(currentMinerId, maxMinerId)

		data += fmt.Sprintf(`<tr><td><strong>Raw&nbsp;data</strong></td><td><a href='ajax?controllerName=getBlock&id=%d&download=1' target='_blank'>Download</a></td></tr>`, parser.BlockData.BlockId)
		data += fmt.Sprintf(`<tr><td><strong>Block_id</strong></td><td>%d (<a href="#" onclick="dc_navigate('blockExplorer', {'blockId':%d})">Previous</a> / <a href="#" onclick="dc_navigate('blockExplorer', {'blockId':%d})">Next</a> )</td></tr>`, parser.BlockData.BlockId, previous, next)
		data += fmt.Sprintf(`<tr><td><strong>Hash</strong></td><td>%s</td></tr>`, hash)
		data += fmt.Sprintf(`<tr><td><strong>Time</strong></td><td><span class='unixtime'>%d</span> / %d</td></tr>`, parser.BlockData.Time, parser.BlockData.Time)
		data += fmt.Sprintf(`<tr><td><strong>User_id</strong></td><td>%d</td></tr>`, parser.BlockData.UserId)
		data += fmt.Sprintf(`<tr><td><strong>Miner_Id</strong></td><td>%d</td></tr>`, minerId)
		data += fmt.Sprintf(`<tr><td><strong>Level</strong></td><td>%d (%v) next: %v</td></tr>`, parser.BlockData.Level, levelsRange, NextBlockLevelsRange)
		data += fmt.Sprintf(`<tr><td><strong>Sign</strong></td><td>%s</td></tr>`, parser.BlockData.Sign)
		if len(parser.TxMapArr) > 0 {
			data += `<tr><td><strong>Transactions</strong></td><td><div><pre style='width: 700px'>`
			for i := 0; i < len(parser.TxMapArr); i++ {
				for k, data_ := range parser.TxMapArr[i] {
					if utils.InSliceString(k, binToHexArray) {
						parser.TxMapArr[i][k] = utils.BinToHex(data_)
					}
					if k == "file" {
						parser.TxMapArr[i][k] = []byte("file size: " + utils.IntToStr(len(data_)))
					} else if k == "code" {
						parser.TxMapArr[i][k] = utils.DSha256(data_)
					} else if k == "secret" {
						parser.TxMapArr[i][k] = utils.BinToHex(data_)
					}
					data += fmt.Sprintf("%v : %s\n", k, parser.TxMapArr[i][k])
				}
				data += "\n\n"
			}

			data += "</pre></div></td></tr>"
		}
		data += "</table>"
	}

	// пока панель тут
	myNotice := make(map[string]string)
	if c.SessUserId > 0 {
		myNotice, err = c.GetMyNoticeData(c.SessUserId, c.SessUserId, c.MyPrefix, c.Lang)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	TemplateStr, err := makeTemplate("block_explorer", "blockExplorer", &BlockExplorerPage{
		Lang:           c.Lang,
		CurrencyList:   c.CurrencyListCf,
		MyNotice:       myNotice,
		Data:           data,
		Start:          start,
		BlockId:        blockId,
		PoolAdmin:      c.PoolAdmin,
		SessRestricted: c.SessRestricted,
		UserId:         c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
