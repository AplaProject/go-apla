package daemons

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"strings"
	//	"io"
)

/*
 * просто шлем всем, кто есть в nodes_connection хэши блока и тр-ий
 * если мы не майнер, то шлем всю тр-ию целиком, блоки слать не можем
 * если майнер - то шлем только хэши, т.к. у нас есть хост, откуда всё можно скачать
 * */
func Disseminator(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "Disseminator"
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
		d.sleepTime = 1
	}
	if !d.CheckInstall(chBreaker, chAnswer, GoroutineName) {
		return
	}
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}

BEGIN:
	for {
		logger.Info(GoroutineName)
		MonitorDaemonCh <- []string{GoroutineName, utils.Int64ToStr(utils.Time())}

		// проверим, не нужно ли нам выйти из цикла
		if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
			break BEGIN
		}

		var hosts []map[string]string
		var nodeData map[string]string
		nodeConfig, err := d.GetNodeConfig()
		if len(nodeConfig["local_gate_ip"]) == 0 {
			// обычный режим
			hosts, err = d.GetAll(`
					SELECT *
					FROM nodes_connection
					`, -1)
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue
			}
			if len(hosts) == 0 {
				if d.dSleep(d.sleepTime) {
					break BEGIN
				}
				logger.Debug("len(hosts) == 0")
				continue
			}
		} else {
			// защищенный режим
			nodeData, err = d.OneRow("SELECT node_public_key, CASE WHEN m.pool_user_id > 0 then (SELECT tcp_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE tcp_host end as tcp_host FROM miners_data as m WHERE m.user_id = ?", nodeConfig["static_node_user_id"]).String()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue
			}
			hosts = append(hosts, map[string]string{"host": nodeConfig["local_gate_ip"], "node_public_key": nodeData["node_public_key"], "user_id": nodeConfig["static_node_user_id"]})
		}

		myUsersIds, err := d.GetMyUsersIds(false, false)
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue
		}
		myMinersIds, err := d.GetMyMinersIds(myUsersIds)
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Debug("%v", myUsersIds)
		logger.Debug("%v", myMinersIds)

		// если среди тр-ий есть смена нодовского ключа, то слать через отправку хэшей с последющей отдачей данных может не получиться
		// т.к. при некорректном нодовском ключе придет зашифрованый запрос на отдачу данных, а мы его не сможем расшифровать т.к. ключ у нас неверный
		var changeNodeKey int64
		if len(myUsersIds) > 0 {
			changeNodeKey, err = d.Single(`
				SELECT count(*)
				FROM transactions
				WHERE type = ? AND
							 user_id IN (`+strings.Join(utils.SliceInt64ToString(myUsersIds), ",")+`)
				`, utils.TypeInt("ChangeNodeKey")).Int64()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}

		var dataType int64 // это тип для того, чтобы принимающая сторона могла понять, как именно надо обрабатывать присланные данные

		// если я майнер и работаю в обычном режиме, то должен слать хэши
		if len(myMinersIds) > 0 && len(nodeConfig["local_gate_ip"]) == 0 && changeNodeKey == 0 {

			logger.Debug("0")

			dataType = 1

			// определим, от кого будем слать
			r := utils.RandInt(0, len(myMinersIds))
			myMinerId := myMinersIds[r]
			myUserId, err := d.Single("SELECT user_id FROM miners_data WHERE miner_id  =  ?", myMinerId).Int64()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// возьмем хэш текущего блока и номер блока
			// для теста ролбеков отключим на время
			data, err := d.OneRow("SELECT block_id, hash, head_hash FROM info_block WHERE sent  =  0").Bytes()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			err = d.ExecSql("UPDATE info_block SET sent = 1")
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			/*
			 * Составляем данные на отправку
			 * */
			// 5 байт = наш user_id. Но они будут не первые, т.к. m_curl допишет вперед user_id получателя (нужно для пулов)
			toBeSent := utils.DecToBin(myUserId, 5)
			if len(data) > 0 { // блок
				// если 5-й байт = 0, то на приемнике будем читать блок, если = 1 , то сразу хэши тр-ий
				toBeSent = append(toBeSent, utils.DecToBin(0, 1)...)
				toBeSent = append(toBeSent, utils.DecToBin(utils.BytesToInt64(data["block_id"]), 3)...)
				toBeSent = append(toBeSent, data["hash"]...)
				toBeSent = append(toBeSent, data["head_hash"]...)
				err = d.ExecSql("UPDATE info_block SET sent = 1")
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			} else { // тр-ии без блока
				toBeSent = append(toBeSent, utils.DecToBin(1, 1)...)
			}

			// возьмем хэши тр-ий
			//utils.WriteSelectiveLog("SELECT hash, high_rate FROM transactions WHERE sent = 0 AND for_self_use = 0")
			transactions, err := d.GetAll("SELECT hash, high_rate FROM transactions WHERE sent = 0 AND for_self_use = 0", -1)
			if err != nil {
				utils.WriteSelectiveLog(err)
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			// нет ни транзакций, ни блока для отправки...
			if len(transactions) == 0 && len(toBeSent) < 10 {
				//utils.WriteSelectiveLog("len(transactions) == 0")
				//log.Debug("len(transactions) == 0")
				if d.dSleep(d.sleepTime) {
					break BEGIN
				}
				logger.Debug("len(transactions) == 0 && len(toBeSent) == 0")
				continue BEGIN
			}
			for _, data := range transactions {
				hexHash := utils.BinToHex([]byte(data["hash"]))
				toBeSent = append(toBeSent, utils.DecToBin(utils.StrToInt64(data["high_rate"]), 1)...)
				toBeSent = append(toBeSent, []byte(data["hash"])...)
				utils.WriteSelectiveLog("UPDATE transactions SET sent = 1 WHERE hex(hash) = " + string(hexHash))
				affect, err := d.ExecSqlGetAffect("UPDATE transactions SET sent = 1 WHERE hex(hash) = ?", hexHash)
				if err != nil {
					utils.WriteSelectiveLog(err)
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
			}

			// отправляем блок и хэши тр-ий, если есть что отправлять
			if len(toBeSent) > 0 {
				for _, host := range hosts {
					go d.DisseminatorType1(host["host"], utils.StrToInt64(host["user_id"]), host["node_public_key"], toBeSent, dataType)
				}
			}
		} else {

			logger.Debug("1")

			var remoteNodeHost string
			// если просто юзер или работаю в защищенном режиме, то шлю тр-ии целиком. слать блоки не имею права.
			if len(nodeConfig["local_gate_ip"]) > 0 {
				dataType = 3
				remoteNodeHost = nodeData["host"]
			} else {
				dataType = 2
				remoteNodeHost = ""
			}

			logger.Debug("dataType: %d", dataType)

			var toBeSent []byte // сюда пишем все тр-ии, которые будем слать другим нодам
			// возьмем хэши и сами тр-ии
			utils.WriteSelectiveLog("SELECT hash, data FROM transactions WHERE sent  =  0")
			rows, err := d.Query("SELECT hash, data FROM transactions WHERE sent  =  0")
			if err != nil {
				utils.WriteSelectiveLog(err)
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			for rows.Next() {
				var hash, data []byte
				err = rows.Scan(&hash, &data)
				if err != nil {
					rows.Close()
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				logger.Debug("hash %x", hash)
				hashHex := utils.BinToHex(hash)
				utils.WriteSelectiveLog("UPDATE transactions SET sent = 1 WHERE hex(hash) = " + string(hashHex))
				affect, err := d.ExecSqlGetAffect("UPDATE transactions SET sent = 1 WHERE hex(hash) = ?", hashHex)
				if err != nil {
					utils.WriteSelectiveLog(err)
					rows.Close()
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				utils.WriteSelectiveLog("affect: " + utils.Int64ToStr(affect))
				toBeSent = append(toBeSent, data...)
			}
			rows.Close()

			// шлем тр-ии
			if len(toBeSent) > 0 {
				for _, host := range hosts {

					userId := utils.StrToInt64(host["user_id"])
					go func(host string, userId int64, node_public_key string) {

						logger.Debug("host %v / userId %v", host, userId)

						conn, err := utils.TcpConn(host)
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}
						defer conn.Close()

						randcandidateBlockHash, err := d.Single("SELECT head_hash FROM queue_candidateBlock").String()
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}
						// получаем IV + ключ + зашифрованный текст
						encryptedData, _, _, err := utils.EncryptData(toBeSent, []byte(node_public_key), randcandidateBlockHash)
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}

						// вначале шлем тип данных, чтобы принимающая сторона могла понять, как именно надо обрабатывать присланные данные
						_, err = conn.Write(utils.DecToBin(dataType, 2))
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}

						// т.к. на приеме может быть пул, то нужно дописать в начало user_id, чьим нодовским ключем шифруем
						/*_, err = conn.Write(utils.DecToBin(userId, 5))
						if err != nil {
							log.Error("%v", utils.ErrInfo(err))
							return
						}*/
						encryptedData = append(utils.DecToBin(userId, 5), encryptedData...)

						// это может быть защищенное локальное соедниение (dataType = 3) и принимающему ноду нужно знать, куда дальше слать данные и чьим они зашифрованы ключем
						if len(remoteNodeHost) > 0 {
							/*
								_, err = conn.Write([]byte(remoteNodeHost))
								if err != nil {
									log.Error("%v", utils.ErrInfo(err))
									return
								}*/
							encryptedData = append([]byte(remoteNodeHost), encryptedData...)
						}

						logger.Debug("encryptedData %x", encryptedData)

						// в 4-х байтах пишем размер данных, которые пошлем далее
						size := utils.DecToBin(len(encryptedData), 4)
						_, err = conn.Write(size)
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}
						// далее шлем сами данные
						_, err = conn.Write(encryptedData)
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}

					}(host["host"], userId, host["node_public_key"])
				}
			}
		}

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}

func (d *daemon) DisseminatorType1(host string, userId int64, node_public_key string, toBeSent []byte, dataType int64) {

	logger.Debug("host %v / userId %v", host, userId)

	// шлем данные указанному хосту
	conn, err := utils.TcpConn(host)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	defer conn.Close()

	randcandidateBlockHash, err := d.Single("SELECT head_hash FROM queue_candidateBlock").String()
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	// получаем IV + ключ + зашифрованный текст
	dataToBeSent, key, iv, err := utils.EncryptData(toBeSent, []byte(node_public_key), randcandidateBlockHash)
	logger.Debug("key: %s", key)
	logger.Debug("iv: %s", iv)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}

	// вначале шлем тип данных, чтобы принимающая сторона могла понять, как именно надо обрабатывать присланные данные
	n, err := conn.Write(utils.DecToBin(dataType, 2))
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	logger.Debug("n: %x", n)

	// т.к. на приеме может быть пул, то нужно дописать в начало user_id, чьим нодовским ключем шифруем
	dataToBeSent = append(utils.DecToBin(userId, 5), dataToBeSent...)
	logger.Debug("dataToBeSent: %x", dataToBeSent)

	// в 4-х байтах пишем размер данных, которые пошлем далее
	size := utils.DecToBin(len(dataToBeSent), 4)
	n, err = conn.Write(size)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	logger.Debug("n: %x", n)
	n, err = conn.Write(dataToBeSent)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	logger.Debug("n: %d / size: %v / len: %d", n, utils.BinToDec(size), len(dataToBeSent))

	// в ответ получаем размер данных, которые нам хочет передать сервер
	buf := make([]byte, 4)
	n, err = conn.Read(buf)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	logger.Debug("n: %x", n)
	dataSize := utils.BinToDec(buf)
	logger.Debug("dataSize %d", dataSize)
	// и если данных менее 1мб, то получаем их
	if dataSize < 1048576 {
		encBinaryTxHashes := make([]byte, dataSize)
		_, err = io.ReadFull(conn, encBinaryTxHashes)
		if err != nil {
			logger.Error("%v", utils.ErrInfo(err))
			return
		}
		// разбираем полученные данные
		binaryTxHashes, err := utils.DecryptCFB(iv, encBinaryTxHashes, key)
		if err != nil {
			logger.Error("%v", utils.ErrInfo(err))
			return
		}
		logger.Debug("binaryTxHashes %x", binaryTxHashes)
		var binaryTx []byte
		for {
			// Разбираем список транзакций
			txHash := make([]byte, 16)
			if len(binaryTxHashes) >= 16 {
				txHash = utils.BytesShift(&binaryTxHashes, 16)
			}
			txHash = utils.BinToHex(txHash)
			logger.Debug("txHash %s", txHash)
			utils.WriteSelectiveLog("SELECT data FROM transactions WHERE hex(hash) = " + string(txHash))
			tx, err := d.Single("SELECT data FROM transactions WHERE hex(hash) = ?", txHash).Bytes()
			logger.Debug("tx %x", tx)
			if err != nil {
				utils.WriteSelectiveLog(err)
				logger.Error("%v", utils.ErrInfo(err))
				return
			}
			utils.WriteSelectiveLog("tx: " + string(utils.BinToHex(tx)))
			if len(tx) > 0 {
				binaryTx = append(binaryTx, utils.EncodeLengthPlusData(tx)...)
			}
			if len(binaryTxHashes) == 0 {
				break
			}
		}

		logger.Debug("binaryTx %x", binaryTx)
		// шифруем тр-ии. Вначале encData добавляется IV
		encData, _, err := utils.EncryptCFB(binaryTx, key, iv)
		if err != nil {
			logger.Error("%v", utils.ErrInfo(err))
			return
		}
		logger.Debug("encData %x", encData)

		// шлем серверу
		// в первых 4-х байтах пишем размер данных, которые пошлем далее
		size := utils.DecToBin(len(encData), 4)
		_, err = conn.Write(size)
		if err != nil {
			logger.Error("%v", utils.ErrInfo(err))
			return
		}
		// далее шлем сами данные
		_, err = conn.Write(encData)
		if err != nil {
			logger.Error("%v", utils.ErrInfo(err))
			return
		}
	}
}
