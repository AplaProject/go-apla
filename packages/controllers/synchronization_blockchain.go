package controllers

import (
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"net/http"
	"os"
	"time"
)

var (
	// при запуске данные могут еще не успеть обновиться
	timeSynchro int64 // Когда первый запуск
	lastSBlock  int64 // последний блок
	lastSTime   int64
)

func (c *Controller) SynchronizationBlockchain() (string, error) {

	if c.DCDB == nil || c.DCDB.DB == nil {
		return "", nil
	}
	blockData, err := c.DCDB.GetInfoBlock()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))

		var ( downloadFile, blockUrl string
			fileSize int64
		)
		if len(utils.SqliteDbUrl) > 0 {
			downloadFile = *utils.Dir + "/litedb.db"
			blockUrl = utils.SqliteDbUrl
		} else {
			downloadFile = *utils.Dir + "/public/blockchain"
			nodeConfig, err := c.GetNodeConfig()
			if err != nil {
				return "", err
			}
			blockUrl = nodeConfig["first_load_blockchain_url"]
			if len(blockUrl) == 0 {
				blockUrl = consts.BLOCKCHAIN_URL
			}
		}
		resp, err := http.Get(blockUrl)
		if err != nil {
			return "", err
		}
		fileSize = resp.ContentLength
		resp.Body.Close()

		// качается блок
		file, err := os.Open(downloadFile)
		if err != nil {
			return "", err
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			return "", err
		}
		if stat.Size() > 0 {
			log.Debug("stat.Size(): %v", int(stat.Size()))
			return `{"download": "` + utils.Int64ToStr(int64(utils.Round(float64((float64(stat.Size())/float64(fileSize))*100), 0))) + `"}`, nil
		} else {
			return `{"download": "0"}`, nil
		}
	}
	blockId := blockData["block_id"]
	blockTime := blockData["time"]
	if len(blockId) == 0 {
		blockId = "0"
	}
	if len(blockTime) == 0 {
		blockTime = "0"
	}

	wTime := int64(12)
	wTimeReady := int64(2)
	if c.ConfigIni["test_mode"] == "1" {
		wTime = 2 * 365 * 86400
		wTimeReady = 2 * 365 * 86400
	}
	log.Debug("wTime: %v / utils.Time(): %v / blockData[time]: %v", wTime, utils.Time(), utils.StrToInt64(blockData["time"]))
	// если время менее 12 часов от текущего, то выдаем не подвержденные, а просто те, что есть в блокчейне
	if utils.Time()-utils.StrToInt64(blockData["time"]) < 3600*wTime {
		lastBlockData, err := c.DCDB.GetLastBlockData()
		if err != nil {
			return "", err
		}
		log.Debug("lastBlockData[lastBlockTime]: %v", lastBlockData["lastBlockTime"])
		log.Debug("time.Now().Unix(): %v", time.Now().Unix())
		// если уже почти собрали все блоки
		if time.Now().Unix()-lastBlockData["lastBlockTime"] < 3600*wTimeReady {
			blockId = "-1"
			blockTime = "-1"
		}
	}

	connections, err := c.Single(`SELECT count(*) from nodes_connection`).String()
	if err != nil {
		return "", err
	}
	confirmedBlockId, err := c.GetConfirmedBlockId()
	if err != nil {
		return "", err
	}

	currentLoadBlockchain := "nodes"
	if c.NodeConfig["current_load_blockchain"] == "file" {
		currentLoadBlockchain = c.NodeConfig["first_load_blockchain_url"]
	}
	var needReload string
	iBlock := utils.StrToInt64( blockId )
	if ( timeSynchro == 0 ) {
		timeSynchro = utils.Time()
		lastSBlock = iBlock
		lastSTime = utils.Time()
	} else if utils.Time() - timeSynchro > 300 { // Тут можно поставить минут 20 или меньше
		if lastSBlock != iBlock {
			lastSBlock = iBlock
			lastSTime = utils.Time()
		} else if utils.Time() - lastSTime > 60 { // Ставим timeout на очередной блок в 60 секунд
			// Имеет смысл проверять последний блок
			if utils.Time() - utils.StrToInt64( blockTime ) > 3600 {
				needReload = `1`
			}
		}
	}

	result := map[string]string{"block_id": blockId, "confirmed_block_id": utils.Int64ToStr(confirmedBlockId), 
	     "block_time": blockTime, "connections": connections, "current_load_blockchain": currentLoadBlockchain,
		 "need_reload": needReload}
	resultJ, _ := json.Marshal(result)

	return string(resultJ), nil
}
