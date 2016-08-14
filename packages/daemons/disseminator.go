package daemons

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io"
	"github.com/DayLightProject/go-daylight/packages/consts"
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
	d.sleepTime = 1

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



		hosts, err := d.GetHosts()
		if err != nil {
			logger.Error("%v", err)
		}

		myCBID, myWalletId, err := d.GetMyCBIDAndWalletId();
		logger.Debug("%v", myWalletId)
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}

		fullNode := true
		if myCBID > 0 {
			delegate, err := d.CheckDelegateCB(myCBID)
			if err != nil {
				d.dbUnlock()
				logger.Error("%v", err)
				if d.dSleep(d.sleepTime) {
					break BEGIN
				}
				continue
			}
			// Если мы - ЦБ и у нас указан delegate, т.е. мы делегировали полномочия по поддержанию ноды другому юзеру или ЦБ, то выходим.
			if delegate {
				fullNode = false
			}
		}

		// Есть ли мы в списке тех, кто может генерить блоки
		full_node_id, err:= d.Single("SELECT full_node_id FROM full_nodes WHERE final_delegate_cb_id = ? OR final_delegate_wallet_id = ? OR cb_id = ? OR wallet_id = ?", myCBID, myWalletId, myCBID, myWalletId).Int64()
		if err != nil {
			d.dbUnlock()
			logger.Error("%v", err)
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		if full_node_id == 0 {
			fullNode = false
		}


		var dataType int64 // это тип для того, чтобы принимающая сторона могла понять, как именно надо обрабатывать присланные данные

		// если мы - fullNode, то должны слать хэши, блоки сами стянут
		if fullNode {

			logger.Debug("dataType = 1")

			dataType = 1

			// возьмем хэш текущего блока и номер блока
			// для теста ролбеков отключим на время
			data, err := d.OneRow("SELECT block_id, hash FROM info_block WHERE sent  =  0").Bytes()
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
			toBeSent := []byte{}
			toBeSent = append(toBeSent, utils.DecToBin(full_node_id, 2)...)
			if len(data) > 0 { // блок
				// если 0, то на приемнике будем читать блок, если = 1 , то сразу хэши тр-ий
				toBeSent = append(toBeSent, utils.DecToBin(0, 1)...)
				toBeSent = append(toBeSent, utils.DecToBin(utils.BytesToInt64(data["block_id"]), 3)...)
				toBeSent = append(toBeSent, data["hash"]...)
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
			logger.Debug("toBeSent block %x", toBeSent)

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
				toBeSent = append(toBeSent, []byte(data["hash"])...)
				logger.Debug("hash %x", data["hash"])
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

			logger.Debug("toBeSent %x", toBeSent)
			// отправляем блок и хэши тр-ий, если есть что отправлять
			if len(toBeSent) > 0 {
				for _, host := range hosts {
					go d.DisseminatorType1(host+":"+consts.TCP_PORT, toBeSent, dataType)
				}
			}
		} else {

			logger.Debug("1")

			dataType = 2

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

					go func(host string) {

						logger.Debug("host %v", host)

						conn, err := utils.TcpConn(host)
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}
						defer conn.Close()


						// вначале шлем тип данных, чтобы принимающая сторона могла понять, как именно надо обрабатывать присланные данные
						_, err = conn.Write(utils.DecToBin(dataType, 2))
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}

						// в 4-х байтах пишем размер данных, которые пошлем далее
						size := utils.DecToBin(len(toBeSent), 4)
						_, err = conn.Write(size)
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}
						// далее шлем сами данные
						_, err = conn.Write(toBeSent)
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}

					}(host+":"+consts.TCP_PORT)
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

func (d *daemon) DisseminatorType1(host string, toBeSent []byte, dataType int64) {

	logger.Debug("host %v", host)

	// шлем данные указанному хосту
	conn, err := utils.TcpConn(host)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	defer conn.Close()

	// вначале шлем тип данных, чтобы принимающая сторона могла понять, как именно надо обрабатывать присланные данные
	n, err := conn.Write(utils.DecToBin(dataType, 2))
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	logger.Debug("n: %x", n)

	// в 4-х байтах пишем размер данных, которые пошлем далее
	size := utils.DecToBin(len(toBeSent), 4)
	n, err = conn.Write(size)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	logger.Debug("n: %x", n)
	n, err = conn.Write(toBeSent)
	if err != nil {
		logger.Error("%v", utils.ErrInfo(err))
		return
	}
	logger.Debug("n: %d / size: %v / len: %d", n, utils.BinToDec(size), len(toBeSent))

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
	// и если данных менее MAX_TX_SIZE, то получаем их
	if dataSize < consts.MAX_TX_SIZE && dataSize > 0 {
		binaryTxHashes := make([]byte, dataSize)
		_, err = io.ReadFull(conn, binaryTxHashes)
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

		// шлем серверу
		// в первых 4-х байтах пишем размер данных, которые пошлем далее
		size := utils.DecToBin(len(binaryTx), 4)
		_, err = conn.Write(size)
		if err != nil {
			logger.Error("%v", utils.ErrInfo(err))
			return
		}
		// далее шлем сами данные
		_, err = conn.Write(binaryTx)
		if err != nil {
			logger.Error("%v", utils.ErrInfo(err))
			return
		}
	}
}
