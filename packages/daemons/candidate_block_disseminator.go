package daemons

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

/**
 * Демон, который мониторит таблу candidateBlock и если видит status=active,
 * то шлет блок строго тем, кто находятся на одном с нами уровне. Если пошлет
 * тем, кто не на одном уровне, то блок просто проигнорируется
 *
 */
func candidateBlockDisseminator(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "candidateBlockDisseminator"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	if utils.Mobile() {
		d.sleepTime = 3600
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

	err = d.notMinerSetSleepTime(1800)
	if err != nil {
		logger.Error("%v", err)
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

		nodeConfig, err := d.GetNodeConfig()
		if len(nodeConfig["local_gate_ip"]) != 0 {
			if d.dPrintSleep("local_gate_ip", d.sleepTime) {
				break BEGIN
			}
			continue
		}

		_, _, _, _, level, levelsRange, err := d.Candidate_block()
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Debug("level: %v", level)
		logger.Debug("levelsRange: %v", levelsRange)
		// получим id майнеров, которые на нашем уровне
		nodesIds := utils.GetOurLevelNodes(level, levelsRange)
		if len(nodesIds) == 0 {
			logger.Debug("len(nodesIds) == 0")
			if d.dSleep(d.sleepTime) {
				break BEGIN
			}
			continue
		}
		logger.Debug("nodesIds: %v", nodesIds)

		// получим хосты майнеров, которые на нашем уровне
		hosts_, err := d.GetList("SELECT CASE WHEN m.pool_user_id > 0 then (SELECT tcp_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE tcp_host end FROM miners_data as m WHERE miner_id IN (" + strings.Join(utils.SliceInt64ToString(nodesIds), `,`) + ")").String()
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue
		}

		hosts := []string{}
		for _, host := range hosts_ {
			if !stringInSlice(host, hosts) {
				hosts = append(hosts, host)
			}
		}

		logger.Debug("hosts: %v", hosts)

		// шлем block_id, user_id, mrkl_root, signature
		data, err := d.OneRow("SELECT block_id, time, user_id, mrkl_root, signature FROM candidateBlock WHERE status  =  'active' AND sent=0").String()
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue
		}
		if len(data) > 0 {

			err = d.ExecSql("UPDATE candidateBlock SET sent=1")
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue
			}

			dataToBeSent := utils.DecToBin(utils.StrToInt64(data["block_id"]), 4)
			dataToBeSent = append(dataToBeSent, utils.DecToBin(data["time"], 4)...)
			dataToBeSent = append(dataToBeSent, utils.DecToBin(data["user_id"], 4)...)
			dataToBeSent = append(dataToBeSent, []byte(data["mrkl_root"])...)
			dataToBeSent = append(dataToBeSent, utils.EncodeLengthPlusData(data["signature"])...)

			for _, host := range hosts {
				go func(host string) {
					logger.Debug("host: %v", host)
					conn, err := utils.TcpConn(host)
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
						return
					}
					defer conn.Close()

					// вначале шлем тип данных
					_, err = conn.Write(utils.DecToBin(6, 2))
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
						return
					}

					// в 4-х байтах пишем размер данных, которые пошлем далее
					_, err = conn.Write(utils.DecToBin(len(dataToBeSent), 4))
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
						return
					}
					// далее шлем сами данные
					logger.Debug("dataToBeSent: %x", dataToBeSent)
					_, err = conn.Write(dataToBeSent)
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
						return
					}

					/*
					 * Получаем тр-ии, которые есть у юзера, в ответ выдаем те, что недостают и
					 * их порядок следования, чтобы получить валидный блок
					 */
					buf := make([]byte, 4)
					_, err = conn.Read(buf)
					if err != nil {
						logger.Error("%v", utils.ErrInfo(err))
						return
					}
					dataSize := utils.BinToDec(buf)
					// и если данных менее 10мб, то получаем их
					if dataSize < 10485760 {

						data, err := d.OneRow("SELECT * FROM candidateBlock").String()
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}

						responseBinaryData := utils.DecToBin(utils.StrToInt64(data["block_id"]), 4)
						responseBinaryData = append(responseBinaryData, utils.DecToBin(utils.StrToInt64(data["time"]), 4)...)
						responseBinaryData = append(responseBinaryData, utils.DecToBin(utils.StrToInt64(data["user_id"]), 5)...)
						responseBinaryData = append(responseBinaryData, utils.EncodeLengthPlusData(data["signature"])...)

						addSql := ""
						if dataSize > 0 {
							binaryData := make([]byte, dataSize)
							_, err := conn.Read(binaryData)
							if err != nil {
								logger.Error("%v", utils.ErrInfo(err))
								return
							}

							// разбираем присланные данные
							// получим хэши тр-ий, которые надо исключить
							for {
								if len(binaryData) < 16 {
									break
								}
								txHex := utils.BinToHex(utils.BytesShift(&binaryData, 16))
								// проверим
								addSql += string(txHex) + ","
								if len(binaryData) == 0 {
									break
								}
							}
							addSql = addSql[:len(addSql)-1]
							addSql = "WHERE id NOT IN (" + addSql + ")"
						}
						// сами тр-ии
						var transactions []byte
						transactions_candidate_block, err := d.GetList(`SELECT data FROM transactions_candidate_block ` + addSql).String()
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}
						for _, txData := range transactions_candidate_block {
							transactions = append(transactions, utils.EncodeLengthPlusData(txData)...)
						}

						responseBinaryData = append(responseBinaryData, utils.EncodeLengthPlusData(transactions)...)

						// порядок тр-ий
						transactions_candidate_block, err = d.GetList(`SELECT hash FROM transactions_candidate_block ORDER BY id ASC`).String()
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}
						for _, txHash := range transactions_candidate_block {
							responseBinaryData = append(responseBinaryData, []byte(txHash)...)
						}

						// в 4-х байтах пишем размер данных, которые пошлем далее
						_, err = conn.Write(utils.DecToBin(len(responseBinaryData), 4))
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}

						// далее шлем сами данные
						_, err = conn.Write(responseBinaryData)
						if err != nil {
							logger.Error("%v", utils.ErrInfo(err))
							return
						}
					}

				}(host)
			}
		}

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)

}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
