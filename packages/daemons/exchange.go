package daemons

import (
	//"fmt"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/DayLightProject/go-daylight/packages/parser"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
)

func Exchange(chBreaker chan bool, chAnswer chan string) {

	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "Exchange"
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
		d.sleepTime = 60
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

		blockId, err := d.GetConfirmedBlockId()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		var myPrefix string
		community, err := d.GetCommunityUsers()
		if len(community) > 0 {
			adminUserId, err := d.GetPoolAdminUserId()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			myPrefix = utils.Int64ToStr(adminUserId) + "_"
		} else {
			myPrefix = ""
		}

		eConfig, err := d.GetMap(`SELECT * FROM e_config`, "name", "value")
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		confirmations := utils.StrToInt64(eConfig["confirmations"])
		mainDcAccount := utils.StrToInt64(eConfig["main_dc_account"])
		logger.Debug("confirmations", confirmations)

		// все валюты, с которыми работаем
		currencyList, err := utils.EGetCurrencyList()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// ++++++++++++ reduction ++++++++++++

		// максимальный номер блока для процентов. Чтобы брать только новые
		maxReductionBlock, err := d.Single(`SELECT max(block_id) FROM e_reduction`).Int64()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// если есть свежая reduction, то нужно остановить торги
		reduction, err := d.Single(`SELECT block_id FROM reduction WHERE block_id > ? and pct > 0`, maxReductionBlock).Int64()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if reduction > 0 {
			err = d.ExecSql(`INSERT INTO e_reduction_lock (time) VALUES (?)`, utils.Time())
			if err != nil {
				logger.Error("%v", utils.ErrInfo(err))
			}
		}

		// если уже прошло 10 блоков с момента обнаружения reduction, то производим сокращение объема монет
		rows, err := d.Query(d.FormatQuery(`SELECT pct, currency_id, time, block_id FROM reduction WHERE block_id > ? AND
	block_id < ?`), maxReductionBlock, blockId-confirmations)
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		for rows.Next() {
			var pct float64
			var currencyId, rTime, blockId int64
			err = rows.Scan(&pct, &currencyId, &rTime, &blockId)
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			//	$k = (100-$row['pct'])/100;
			k := (100 - pct) / 100
			// уменьшаем все средства на счетах
			err = d.ExecSql(`UPDATE e_wallets SET amount = amount * ? WHERE currency_id = ?`, k, currencyId)

			// уменьшаем все средства на вывод
			err = d.ExecSql(`UPDATE e_withdraw SET amount = amount * ? WHERE currency_id = ? AND close_time = 0`, k, currencyId)

			// уменьшаем все ордеры на продажу
			err = d.ExecSql(`UPDATE e_orders SET amount = amount * ?, begin_amount = begin_amount * ? WHERE sell_currency_id = ?`, k, k, currencyId)

			err = d.ExecSql(`INSERT INTO e_reduction (
					time,
					block_id,
					currency_id,
					pct
				)
				VALUES (
					?,
					?,
					?,
					?
				)`, rTime, blockId, currencyId, pct)
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}
		rows.Close()

		// ++++++++++++ DC ++++++++++++
		/*
		 * Важно! отключать в кроне при обнулении данных в БД
		 *
		 * 1. Получаем инфу о входящих переводах и начисляем их на счета юзеров
		 * 2. Обновляем проценты
		 * 3. Чистим кол-во отправленных смс-ок
		 * */
		nodePrivateKey, err := d.GetNodePrivateKey()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		/*
		 * Получаем инфу о входящих переводах и начисляем их на счета юзеров
		 * */

		// если всё остановлено из-за найденного блока с reduction, то входящие переводы не обрабатываем
		reductionLock, err := utils.EGetReductionLock()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if reductionLock > 0 {
			if d.dPrintSleep(utils.ErrInfo("reductionLock"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		logger.Debug("blockId-confirmations", blockId-confirmations)
		logger.Debug("mainDcAccount", mainDcAccount)
		rows, err = d.Query(d.FormatQuery(`
				SELECT amount, id, block_id, type_id, currency_id, to_user_id, time, comment, comment_status
				FROM my_dc_transactions
				WHERE type = 'from_user' AND
					  block_id < ? AND
					  exchange_checked = 0 AND
					  status = 'approved' AND
					  to_user_id = ?
				ORDER BY id DESC`), blockId-confirmations, mainDcAccount)
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		for rows.Next() {
			var amount float64
			var id, blockId, typeId, currencyId, toUserId, txTime int64
			var comment, commentStatus string
			err = rows.Scan(&amount, &id, &blockId, &typeId, &currencyId, &toUserId, &txTime, &comment, &commentStatus)
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			// отметим exchange_checked=1, чтобы больше не брать эту тр-ию
			err = d.ExecSql(`UPDATE my_dc_transactions SET exchange_checked = 1 WHERE id = ?`, id)
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// вначале нужно проверить, точно ли есть такой перевод в блоке
			binaryData, err := d.Single(`SELECT data FROM block_chain WHERE id = ?`, blockId).Bytes()
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			p := new(parser.Parser)
			p.DCDB = d.DCDB
			p.BinaryData = binaryData
			p.ParseDataLite()
			for _, txMap := range p.TxMapArr {
				// пропускаем все ненужные тр-ии
				if utils.BytesToInt64(txMap["type"]) != utils.TypeInt("SendDc") {
					continue
				}
				logger.Debug("md5hash %s", txMap["md5hash"])

				// если что-то случится с таблой my_dc_transactions, то все ввода на биржу будут зачислены по новой
				// поэтому нужно проверять e_adding_funds
				exists, err := d.Single(`SELECT id FROM e_adding_funds WHERE hex(tx_hash) = ?`, string(txMap["md5hash"])).Int64()
				if err != nil {
					rows.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				logger.Debug("exists %d", exists)
				if exists != 0 {
					continue
				}

				logger.Debug("user_id = %d / typeId = %d / currency_id = %d / currencyId = %d / amount = %f / amount = %f / comment = %s / comment = %s / to_user_id = %d / toUserId = %d ", utils.BytesToInt64(txMap["user_id"]), typeId, utils.BytesToInt64(txMap["currency_id"]), currencyId, utils.BytesToFloat64(txMap["amount"]), amount, string(utils.BinToHex(txMap["comment"])), comment, utils.BytesToInt64(txMap["to_user_id"]), toUserId)
				// сравнение данных из таблы my_dc_transactions с тем, что в блоке
				if utils.BytesToInt64(txMap["user_id"]) == typeId && utils.BytesToInt64(txMap["currency_id"]) == currencyId && utils.BytesToFloat64(txMap["amount"]) == amount && string(utils.BinToHex(txMap["comment"])) == comment && utils.BytesToInt64(txMap["to_user_id"]) == toUserId {

					decryptedComment := comment
					if commentStatus == "encrypted" {
						// расшифруем коммент
						block, _ := pem.Decode([]byte(nodePrivateKey))
						if block == nil || block.Type != "RSA PRIVATE KEY" {
							rows.Close()
							if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
						private_key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
						if err != nil {
							rows.Close()
							if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
						decryptedComment_, err := rsa.DecryptPKCS1v15(rand.Reader, private_key, utils.HexToBin(comment))
						if err != nil {
							rows.Close()
							if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
						decryptedComment = string(decryptedComment_)
						// запишем расшифрованный коммент, чтобы потом можно было найти перевод в ручном режиме
						err = d.ExecSql("UPDATE "+myPrefix+"my_dc_transactions SET comment = ?, comment_status = 'decrypted' WHERE id = ?", decryptedComment, id)
						if err != nil {
							rows.Close()
							if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
					}

					// возможно юзер сделал перевод в той валюте, которая у нас на бирже еще не торгуется
					if len(currencyList[currencyId]) == 0 {
						logger.Error("currencyId %d not trading", currencyId)
						continue
					}

					// возможно, что чуть раньше было reduction, а это значит, что все тр-ии,
					// которые мы ещё не обработали и которые были До блока с reduction нужно принимать с учетом reduction
					// т.к. средства на нашем счете уже урезались, а  вот те, что после reduction - остались в том виде, в котором пришли
					lastReduction, err := d.OneRow("SELECT block_id, pct FROM reduction WHERE currency_id  = ? ORDER BY block_id", currencyId).Int64()
					if err != nil {
						rows.Close()
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					if blockId <= lastReduction["block_id"] {
						// сумму с учетом reduction
						k0 := (100 - lastReduction["pct"]) / 100
						amount = amount * float64(k0)
					}

					// начисляем средства на счет того, чей id указан в комменте
					r, _ := regexp.Compile(`(?i)\s*#\s*([0-9]+)\s*`)
					user := r.FindStringSubmatch(decryptedComment)
					if len(user) == 0 {
						logger.Error("len(user) == 0")
						continue
					}
					logger.Debug("user %s", user[1])
					// user_id с биржевой таблы
					uid := utils.StrToInt64(user[1])
					userAmountAndProfit := utils.EUserAmountAndProfit(uid, currencyId)
					newAmount_ := userAmountAndProfit + amount

					utils.UpdEWallet(uid, currencyId, utils.Time(), newAmount_, true)

					// для отчетности запишем в историю
					err = d.ExecSql(`
						INSERT INTO e_adding_funds (
							user_id,
							currency_id,
							time,
							amount,
							tx_hash
						)
						VALUES (
							?,
							?,
							?,
							?,
							?
						)`, uid, currencyId, txTime, amount, string(txMap["md5hash"]))
					if err != nil {
						rows.Close()
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
				}
			}
		}
		rows.Close()

		/*
		 * Обновляем проценты
		 * */
		// максимальный номер блока для процентов. Чтобы брать только новые
		maxPctBlock, err := d.Single(`SELECT max(block_id) FROM e_user_pct`).Int64()

		logger.Debug(`SELECT time, block_id, currency_id, user FROM pct WHERE  block_id < ` + utils.Int64ToStr(blockId-confirmations) + ` AND block_id > ` + utils.Int64ToStr(maxPctBlock))
		rows, err = d.Query(d.FormatQuery(`SELECT time, block_id, currency_id, user FROM pct WHERE  block_id < ? AND block_id > ?`), blockId-confirmations, maxPctBlock)
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		for rows.Next() {
			var pct float64
			var pTime, blockId, currencyId int64
			err = rows.Scan(&pTime, &blockId, &currencyId, &pct)
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			d.ExecSql(`
				INSERT INTO e_user_pct (
					time,
					block_id,
					currency_id,
					pct
				)
				VALUES (
					?,
					?,
					?,
					?
				)`, pTime, blockId, currencyId, pct)
		}
		rows.Close()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
