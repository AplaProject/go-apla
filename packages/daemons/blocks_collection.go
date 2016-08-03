package daemons

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/utils"
	_ "github.com/lib/pq"
	"os"
)

func BlocksCollection(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "BlocksCollection"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	if utils.Mobile() {
		d.sleepTime = 60
	} else {
		d.sleepTime = 60
	}
	if !d.CheckInstall(chBreaker, chAnswer, GoroutineName) {
		return
	}
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	//var cur bool
	var file *os.File
BEGIN:
	for {
		if file != nil {
			file.Close()
			file = nil
		}
		logger.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
		if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
			break BEGIN
		}
		logger.Debug("0")
		config, err := d.GetNodeConfig()
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("1")

		// удалим то, что мешает
		if *utils.StartBlockId > 0 {
			del := []string{"queue_tx", "my_notifications", "main_lock"}
			for _, table := range del {
				err := utils.DB.ExecSql(`DELETE FROM `+table)
				fmt.Println(`DELETE FROM `+table)
				if err != nil {
					fmt.Println(err)
					panic(err)
				}
			}
		}

		err, restart := d.dbLock()
		if restart {
			logger.Debug("restart true")
			break BEGIN
		}
		if err != nil {
			logger.Debug("restart err %v", err)
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("2")

		// если это первый запуск во время инсталяции
		currentBlockId, err := d.GetBlockId()
		if err != nil {
			if d.unlockPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		logger.Info("config", config)
		logger.Info("currentBlockId", currentBlockId)

		// на время тестов
		/*if !cur {
		    currentBlockId = 0
		    cur = true
		}*/



		parser := new(dcparser.Parser)
		parser.DCDB = d.DCDB
		parser.GoroutineName = GoroutineName
		if currentBlockId == 0 || *utils.StartBlockId > 0 {
			/*
			   IsNotExistBlockChain := false
			   if _, err := os.Stat(*utils.Dir+"/public/blockchain"); os.IsNotExist(err) {
			       IsNotExistBlockChain = true
			   }*/
			if config["first_load_blockchain"] == "file" /* && IsNotExistBlockChain*/ {

				logger.Info("first_load_blockchain=file")
				nodeConfig, err := d.GetNodeConfig()
				blockchain_url := nodeConfig["first_load_blockchain_url"]
				if len(blockchain_url) == 0 {
					blockchain_url = consts.BLOCKCHAIN_URL
				}
				logger.Debug("blockchain_url: %s", blockchain_url)
				// возможно сервер отдаст блокчейн не с первой попытки
				var blockchainSize int64
				for i := 0; i < 10; i++ {
					logger.Debug("blockchain_url: %s, i: %d", blockchain_url, i)
					blockchainSize, err = utils.DownloadToFile(blockchain_url, *utils.Dir+"/public/blockchain", 3600, chBreaker, chAnswer, GoroutineName)
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
					}
					if blockchainSize > consts.BLOCKCHAIN_SIZE {
						break
					}
				}
				logger.Debug("blockchain dw ok")
				if err != nil || blockchainSize < consts.BLOCKCHAIN_SIZE {
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
					} else {
						logger.Info(fmt.Sprintf("%v < %v", blockchainSize, consts.BLOCKCHAIN_SIZE))
					}
					if d.unlockPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}

				first := true
				/*// блокчейн мог быть загружен ранее. проверим его размер


				  stat, err := file.Stat()
				  if err != nil {
				      if d.unlockPrintSleep(err, d.sleepTime) {	break BEGIN }
				      file.Close()
				      continue BEGIN
				  }
				  if stat.Size() < consts.BLOCKCHAIN_SIZE {
				      d.unlockPrintSleep(fmt.Errorf("%v < %v", stat.Size(), consts.BLOCKCHAIN_SIZE), 1)
				      file.Close()
				      continue BEGIN
				  }*/

				logger.Debug("GO!")
				file, err = os.Open(*utils.Dir + "/public/blockchain")
				if err != nil {
					if d.unlockPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				err = d.ExecSql(`UPDATE config SET current_load_blockchain = 'file'`)
				if err != nil {
					if d.unlockPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}

				for {
					// проверим, не нужно ли нам выйти из цикла
					if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
						d.unlockPrintSleep(fmt.Errorf("DaemonsRestart"), 0)
						break BEGIN
					}
					b1 := make([]byte, 5)
					file.Read(b1)
					dataSize := utils.BinToDec(b1)
					logger.Debug("dataSize", dataSize)
					if dataSize > 0 {

						data := make([]byte, dataSize)
						file.Read(data)
						//logger.Debug("data %x\n", data)
						blockId := utils.BinToDec(data[0:5])
						if *utils.EndBlockId > 0 && blockId == *utils.EndBlockId {
							if d.dPrintSleep(err, d.sleepTime ) {
								break BEGIN
							}
							continue BEGIN
						}
						logger.Info("blockId", blockId)
						data2 := data[5:]
						length := utils.DecodeLength(&data2)
						logger.Debug("length", length)
						//logger.Debug("data2 %x\n", data2)
						blockBin := utils.BytesShift(&data2, length)
						//logger.Debug("blockBin %x\n", blockBin)

						if *utils.StartBlockId == 0 || (*utils.StartBlockId > 0 && blockId > *utils.StartBlockId) {

							// парсинг блока
							parser.BinaryData = blockBin

							if first {
								parser.CurrentVersion = consts.VERSION
								first = false
							}
							if err = parser.ParseDataFull(); err != nil {
								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}
								continue BEGIN
							}
							if err = parser.InsertIntoBlockchain(); err != nil {
								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}
								continue BEGIN
							}

							// отметимся, чтобы не спровоцировать очистку таблиц
							if err = parser.UpdMainLock(); err != nil {
								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}
								continue BEGIN
							}
							if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
								d.unlockPrintSleep(nil, 0) 
/*!!!								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}*/
								break BEGIN
//!!!   						continue BEGIN
							}
						}
						// ненужный тут размер в конце блока данных
						data = make([]byte, 5)
						file.Read(data)
					} else {
						if d.unlockPrintSleep( nil, d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					// utils.Sleep(1)
				}
				file.Close()
				file = nil
			} else {

				newBlock, err := static.Asset("static/1block.bin")
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				parser.BinaryData = newBlock
				parser.CurrentVersion = consts.VERSION

				if err = parser.ParseDataFull(); err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				if err = parser.InsertIntoBlockchain(); err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
			utils.Sleep(1)
			d.dbUnlock()
			continue BEGIN
		}
		d.dbUnlock()

		err = d.ExecSql(`UPDATE config SET current_load_blockchain = 'nodes'`)
		if err != nil {
//!!!			d.unlockPrintSleep(err, d.sleepTime) unlock был выше
			if d.dPrintSleep(err, d.sleepTime) {
				break
			}
			continue
		}

		myConfig, err := d.OneRow("SELECT local_gate_ip, static_node_user_id FROM config").String()
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue
		}

		var hosts []map[string]string
		var nodeHost string
		var dataTypeMaxBlockId, dataTypeBlockBody int64
		if len(myConfig["local_gate_ip"]) > 0 {
			hosts = append(hosts, map[string]string{"host": myConfig["local_gate_ip"], "user_id": myConfig["static_node_user_id"]})
			nodeHost, err = d.Single("SELECT tcp_host FROM miners_data WHERE user_id  =  ?", myConfig["static_node_user_id"]).String()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue
			}
			dataTypeMaxBlockId = 9
			dataTypeBlockBody = 8
			//getBlockScriptName = "ajax?controllerName=protectedGetBlock";
			//addNodeHost = "&nodeHost="+nodeHost;
		} else {
			// получим список нодов, с кем установлено рукопожатие
			hosts, err = d.GetAll("SELECT * FROM nodes_connection", -1)
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue
			}
			dataTypeMaxBlockId = 10
			dataTypeBlockBody = 7
			//getBlockScriptName = "ajax?controllerName=getBlock";
			//addNodeHost = "";
		}

		logger.Info("%v", hosts)
//		fmt.Println(`Hosts`, hosts )
		if len(hosts) == 0 {
			if d.dPrintSleep(err, 1) {
				break BEGIN
			}
			continue
		}
		
		maxBlockId := int64(1)
		maxBlockIdHost := ""
		var maxBlockIdUserId int64
		// получим максимальный номер блока
		for i := 0; i < len(hosts); i++ {
			if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
				break BEGIN
			}
			conn, err := utils.TcpConn(hosts[i]["host"])
			if err != nil {
				if d.dPrintSleep(err, 1) {
					break BEGIN
				}
				continue
			}
			// шлем тип данных
			_, err = conn.Write(utils.DecToBin(dataTypeMaxBlockId, 2))
			if err != nil {
				conn.Close()
				if d.dPrintSleep(err, 1) {
					break BEGIN
				}
				continue
			}
			if len(nodeHost) > 0 { // защищенный режим
				err = utils.WriteSizeAndData([]byte(nodeHost), conn)
				if err != nil {
					conn.Close()
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue
				}
			}
			// в ответ получаем номер блока
			blockIdBin := make([]byte, 4)
			_, err = conn.Read(blockIdBin)
			if err != nil {
				conn.Close()
				if d.dPrintSleep(err, 1) {
					break BEGIN
				}
				continue
			}
			conn.Close()
			id := utils.BinToDec(blockIdBin)
			if id > maxBlockId || i == 0 {
				maxBlockId = id
				maxBlockIdHost = hosts[i]["host"]
				maxBlockIdUserId = utils.StrToInt64(hosts[i]["user_id"])
			}
			if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
				utils.Sleep(1)
				break BEGIN
			}
		}

		// получим наш текущий имеющийся номер блока
		// ждем, пока разлочится и лочим сами, чтобы не попасть в тот момент, когда данные из блока уже занесены в БД, а info_block еще не успел обновиться
		err, restart = d.dbLock()
		if restart {
			break BEGIN
		}
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		currentBlockId, err = d.GetBlockId()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Info("currentBlockId", currentBlockId, "maxBlockId", maxBlockId)
		if maxBlockId <= currentBlockId {
			if d.unlockPrintSleep(utils.ErrInfo(errors.New("maxBlockId <= currentBlockId")), d.sleepTime) {
				break BEGIN
			}
			continue
		}

		fmt.Printf("\nnode: %s curid=%d maxid=%d\n", maxBlockIdHost, currentBlockId, maxBlockId )

		/////----///////
		// в цикле собираем блоки, пока не дойдем до максимального
		for blockId := currentBlockId + 1; blockId < maxBlockId+1; blockId++ {
			d.UpdMainLock()
			if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
				d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime ) 
				break BEGIN
			}
			variables, err := d.GetAllVariables()
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			// качаем тело блока с хоста maxBlockIdHost
			binaryBlock, err := utils.GetBlockBody(maxBlockIdHost, blockId, dataTypeBlockBody, nodeHost)

			if len(binaryBlock) == 0 {
				// баним на 1 час хост, который дал нам пустой блок, хотя должен был дать все до максимального
				// для тестов убрал, потом вставить.
				//nodes_ban ($db, $max_block_id_user_id, substr($binary_block, 0, 512)."\n".__FILE__.', '.__LINE__.', '. __FUNCTION__.', '.__CLASS__.', '. __METHOD__);
				//p.NodesBan(maxBlockIdUserId, "len(binaryBlock) == 0")
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			binaryBlockFull := binaryBlock
			utils.BytesShift(&binaryBlock, 1) // уберем 1-й байт - тип (блок/тр-я)
			// распарсим заголовок блока
			blockData := utils.ParseBlockHeader(&binaryBlock)
			logger.Info("blockData: %v, blockId: %v", blockData, blockId)

			// если существуют глючная цепочка, тот тут мы её проигнорируем
			badBlocks_, err := d.Single("SELECT bad_blocks FROM config").Bytes()
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			badBlocks := make(map[int64]string)
			if len(badBlocks_) > 0 {
				err = json.Unmarshal(badBlocks_, &badBlocks)
				if err != nil {
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
			if badBlocks[blockData.BlockId] == string(utils.BinToHex(blockData.Sign)) {
				d.NodesBan(maxBlockIdUserId, fmt.Sprintf("bad_block = %v => %v", blockData.BlockId, badBlocks[blockData.BlockId]))
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// размер блока не может быть более чем max_block_size
			if currentBlockId > 1 {
				if int64(len(binaryBlock)) > variables.Int64["max_block_size"] {
					d.NodesBan(maxBlockIdUserId, fmt.Sprintf(`len(binaryBlock) > variables.Int64["max_block_size"]  %v > %v`, len(binaryBlock), variables.Int64["max_block_size"]))
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}

			if blockData.BlockId != blockId {
				d.NodesBan(maxBlockIdUserId, fmt.Sprintf(`blockData.BlockId != blockId  %v > %v`, blockData.BlockId, blockId))
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// нам нужен хэш предыдущего блока, чтобы проверить подпись
			prevBlockHash := ""
			if blockId > 1 {
				prevBlockHash, err = d.Single("SELECT hash FROM block_chain WHERE id = ?", blockId-1).String()
				if err != nil {
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				prevBlockHash = string(utils.BinToHex([]byte(prevBlockHash)))
			} else {
				prevBlockHash = "0"
			}
			first := false
			if blockId == 1 {
				first = true
			}
			// нам нужен меркель-рут текущего блока
			mrklRoot, err := utils.GetMrklroot(binaryBlock, variables, first)
			if err != nil {
				d.NodesBan(maxBlockIdUserId, fmt.Sprintf(`%v`, err))
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// публичный ключ того, кто этот блок сгенерил
			nodePublicKey, err := d.GetNodePublicKeyWalletOrCB(blockData.WalletId, blockData.CBID)
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// SIGN от 128 байта до 512 байт. Подпись от TYPE, BLOCK_ID, PREV_BLOCK_HASH, TIME, USER_ID, LEVEL, MRKL_ROOT
			forSign := fmt.Sprintf("0,%v,%v,%v,%v,%v,%s", blockData.BlockId, prevBlockHash, blockData.Time, blockData.WalletId, blockData.CBID, mrklRoot)

			// проверяем подпись
			if !first {
				_, err = utils.CheckSign([][]byte{nodePublicKey}, forSign, blockData.Sign, true)
			}

			// качаем предыдущие блоки до тех пор, пока отличается хэш предыдущего.
			// другими словами, пока подпись с prevBlockHash будет неверной, т.е. пока что-то есть в $error
			if err != nil {
				logger.Error("%v", utils.ErrInfo(err))
				if blockId < 1 {
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				// нужно привести данные в нашей БД в соответствие с данными у того, у кого качаем более свежий блок
				//func (p *Parser) GetOldBlocks (userId, blockId int64, host string, hostUserId int64, goroutineName, getBlockScriptName, addNodeHost string) error {
				err := parser.GetOldBlocks(blockData.WalletId, blockData.CBID, blockId-1, maxBlockIdHost, maxBlockIdUserId, GoroutineName, dataTypeBlockBody, nodeHost)
				if err != nil {
					logger.Error("%v", err)
					d.NodesBan(maxBlockIdUserId, fmt.Sprintf(`blockId: %v / %v`, blockId, err))
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}

			} else {

				logger.Info("plug found blockId=%v\n", blockId)

				// получим наши транзакции в 1 бинарнике, просто для удобства
				var transactions []byte
				utils.WriteSelectiveLog("SELECT data FROM transactions WHERE verified = 1 AND used = 0")
				rows, err := d.Query("SELECT data FROM transactions WHERE verified = 1 AND used = 0")
				if err != nil {
					utils.WriteSelectiveLog(err)
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				for rows.Next() {
					var data []byte
					err = rows.Scan(&data)
					utils.WriteSelectiveLog(utils.BinToHex(data))
					if err != nil {
						rows.Close()
						if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					transactions = append(transactions, utils.EncodeLengthPlusData(data)...)
				}
				rows.Close()
				if len(transactions) > 0 {
					// отмечаем, что эти тр-ии теперь нужно проверять по новой
					utils.WriteSelectiveLog("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
					affect, err := d.ExecSqlGetAffect("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
					if err != nil {
						utils.WriteSelectiveLog(err)
						if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
					// откатываем по фронту все свежие тр-ии
					parser.BinaryData = transactions
					err = parser.ParseDataRollbackFront(false)
					if err != nil {
						utils.Sleep(1)
						continue BEGIN
					}
				}

				err = parser.RollbackTransactionsCandidateBlock(true)
				if err != nil {
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				err = d.ExecSql("DELETE FROM candidateBlock")
				if err != nil {
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}

			// теперь у нас в таблицах всё тоже самое, что у нода, у которого качаем блок
			// и можем этот блок проверить и занести в нашу БД
			parser.BinaryData = binaryBlockFull
			err = parser.ParseDataFull()
			if err == nil {
				err = parser.InsertIntoBlockchain()
				if err != nil {
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
			// начинаем всё с начала уже с другими нодами. Но у нас уже могут быть новые блоки до $block_id, взятые от нода, которого с в итоге мы баним
			if err != nil {
				d.NodesBan(maxBlockIdUserId, fmt.Sprintf(`blockId: %v / %v`, blockId, err))
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}


/*		toclose, err := collectBlocks(currentBlockId, maxBlockId, dataTypeBlockBody, maxBlockIdUserId, maxBlockIdHost, nodeHost); 
		if toclose {
			break
		}
		if err != nil {
			continue
		}*/

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break
			//continue
		}
	}
	if file != nil {
		file.Close()
	}
//  Нужно ли разлочивать базу?
// 	d.dbUnlock()
	logger.Debug("break BEGIN %v", GoroutineName)
}
