package daemons

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"net/http"
	"net/url"
	"regexp"
)

/*
 * Важно! отключать демона при обнулении данных в БД
 */

func Shop(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "Shop"
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
		d.sleepTime = 120
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

		myBlockId, err := d.GetMyBlockId()
		blockId, err := d.GetBlockId()
		if myBlockId > blockId {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		currencyList, err := d.GetCurrencyList(false)
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		// нужно знать текущий блок, который есть у большинства нодов
		blockId, err = d.GetConfirmedBlockId()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}

		// сколько должно быть подтверждений, т.е. кол-во блоков сверху
		confirmations := int64(5)

		// берем всех юзеров по порядку
		community, err := d.GetCommunityUsers()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue
		}
		for _, userId := range community {
			privateKey := ""
			myPrefix := utils.Int64ToStr(userId) + "_"
			allTables, err := d.GetAllTables()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			if !utils.InSliceString(myPrefix+"my_keys", allTables) {
				continue
			}
			// проверим, майнер ли
			minerId, err := d.GetMinerId(userId)
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			if minerId > 0 {
				// наш приватный ключ нода, которым будем расшифровывать комменты
				privateKey, err = d.GetNodePrivateKey(myPrefix)
				if err != nil {
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
			// возможно, что комменты будут зашифрованы юзерским ключем
			if len(privateKey) == 0 {
				privateKey, err = d.GetMyPrivateKey(myPrefix)
				if err != nil {
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
			// если это еще не майнер и админ ноды не указал его приватный ключ в табле my_keys, то $private_key будет пуст
			if len(privateKey) == 0 {
				continue
			}
			myData, err := d.OneRow("SELECT shop_secret_key, shop_callback_url FROM " + myPrefix + "my_table").String()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// Получаем инфу о входящих переводах и начисляем их на счета юзеров
			dq := d.GetQuotes()
			rows, err := d.Query(d.FormatQuery(`
					SELECT id, block_id, type_id, currency_id, amount, to_user_id, comment_status, comment
					FROM `+dq+myPrefix+`my_dc_transactions`+dq+`
					WHERE type = 'from_user' AND
								 block_id < ? AND
								 merchant_checked = 0 AND
								 status = 'approved'
					ORDER BY id DESC
					`), blockId-confirmations)
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			for rows.Next() {
				var id, block_id, type_id, currency_id, to_user_id int64
				var comment_status, comment string
				var amount float64
				err = rows.Scan(&id, &block_id, &type_id, &currency_id, &amount, &to_user_id, &comment_status, &comment)
				if err != nil {
					rows.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				if len(myData["shop_callback_url"]) == 0 {
					// отметим merchant_checked=1, чтобы больше не брать эту тр-ию
					err = d.ExecSql("UPDATE "+myPrefix+"my_dc_transactions SET merchant_checked = 1 WHERE id = ?", id)
					if err != nil {
						rows.Close()
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					continue
				}

				// вначале нужно проверить, точно ли есть такой перевод в блоке
				binaryData, err := d.Single("SELECT data FROM block_chain WHERE id  =  ?", blockId).Bytes()
				if err != nil {
					rows.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				p := new(dcparser.Parser)
				p.DCDB = d.DCDB
				p.BinaryData = binaryData
				p.ParseDataLite()
				for _, txMap := range p.TxMapArr {

					// пропускаем все ненужные тр-ии
					if utils.BytesToInt64(txMap["type"]) != utils.TypeInt("SendDc") {
						continue
					}

					// сравнение данных из таблы my_dc_transactions с тем, что в блоке
					if utils.BytesToInt64(txMap["user_id"]) == userId && utils.BytesToInt64(txMap["currency_id"]) == currency_id && utils.BytesToFloat64(txMap["amount"]) == amount && string(utils.BinToHex(txMap["comment"])) == comment && utils.BytesToInt64(txMap["to_user_id"]) == to_user_id {
						decryptedComment := comment
						// расшифруем коммент
						if comment_status == "encrypted" {
							block, _ := pem.Decode([]byte(privateKey))
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
							decryptedComment_, err := rsa.DecryptPKCS1v15(rand.Reader, private_key, []byte(comment))
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

						// возможно, что чуть раньше было reduction, а это значит, что все тр-ии,
						// которые мы ещё не обработали и которые были До блока с reduction нужно принимать с учетом reduction
						// т.к. средства на нашем счете уже урезались, а  вот те, что после reduction - остались в том виде, в котором пришли
						lastReduction, err := d.OneRow("SELECT block_id, pct FROM reduction WHERE currency_id  = ? ORDER BY block_id", currency_id).Int64()
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

						// делаем запрос к callback скрипту
						r, _ := regexp.Compile(`(?i)\s*#\s*([0-9]+)\s*`)
						order := r.FindStringSubmatch(decryptedComment)
						orderId := 0
						if len(order) > 0 {
							orderId = utils.StrToInt(order[1])
						}
						txId := id
						sign := fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v:%v", amount, currencyList[currency_id], orderId, decryptedComment, txMap["user_id"], blockId, txId, myData["shop_secret_key"])
						data := url.Values{}
						data.Add("amount", utils.Float64ToStrPct(amount))
						data.Add("currency", currencyList[currency_id])
						data.Add("order_id", utils.IntToStr(orderId))
						data.Add("message", decryptedComment)
						data.Add("user_id", string(txMap["user_id"]))
						data.Add("block_id", string(txMap["block_id"]))
						data.Add("tx_id", utils.Int64ToStr(txId))
						data.Add("sign", sign)

						client := &http.Client{}
						req, err := http.NewRequest("POST", myData["shop_callback_url"], bytes.NewBufferString(data.Encode()))
						if err != nil {
							rows.Close()
							if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
						req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
						req.Header.Add("Content-Length", utils.IntToStr(len(data.Encode())))

						resp, err := client.Do(req)
						if err != nil {
							rows.Close()
							if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
						//contents, _ := ioutil.ReadAll(resp.Body)
						if resp.StatusCode == 200 {
							// отметим merchant_checked=1, чтобы больше не брать эту тр-ию
							err = d.ExecSql("UPDATE "+myPrefix+"my_dc_transactions SET merchant_checked = 1 WHERE id = ?", id)
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
			}
			rows.Close()
		}

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
