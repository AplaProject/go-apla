package daemons

import (
	"bytes"
	"github.com/DayLightProject/go-daylight/packages/availablekey"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func FirstChangePkey(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "FirstChangePkey"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	if utils.Mobile() {
		d.sleepTime = 360
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

		community, err := d.GetCommunityUsers()
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		var uids []int64
		if len(community) > 0 {
			uids = community
		} else {
			myuid, err := d.GetMyUserId("")
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			uids = append(uids, myuid)
		}
		logger.Debug("uids %v", uids)
		var status, myPrefix string
		for _, uid := range uids {
			if len(community) > 0 {
				myPrefix = utils.Int64ToStr(uid) + "_"
			}
			status, err = d.Single(`SELECT status FROM ` + myPrefix + `my_table`).String()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			logger.Debug("status: %v / myPrefix: %v", status, myPrefix)
			if status == "waiting_accept_new_key" {

				// если ключ кто-то сменил
				userPublicKey, err := d.GetUserPublicKey(uid)
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				myUserPublicKey, err := d.GetMyPublicKey(myPrefix)
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				if !bytes.Equal(myUserPublicKey, []byte(userPublicKey)) {
					logger.Debug("myUserPublicKey:%s != userPublicKey:%s", utils.BinToHex(myUserPublicKey), utils.BinToHex(userPublicKey))
					// удаляем старый ключ
					err = d.ExecSql(`DELETE FROM ` + myPrefix + `my_keys`)
					if err != nil {
						if d.dPrintSleep(err, d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					// и user_id
					q := `UPDATE ` + myPrefix + `my_table SET status="my_pending", user_id=0`
					if len(community) > 0 {
						q = `DELETE FROM ` + myPrefix + `my_table`
					}
					err = d.ExecSql(q)
					if err != nil {
						if d.dPrintSleep(err, d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					if len(community) > 0 {
						err = d.ExecSql(`DELETE FROM community WHERE user_id = ?`, uid)
						if err != nil {
							if d.dPrintSleep(err, d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
					}
					// и пробуем взять новый
					availablekey := &availablekey.AvailablekeyStruct{}
					availablekey.DCDB = d.DCDB
					userId, _, err := availablekey.GetAvailableKey()
					if err != nil {
						if d.dPrintSleep(err, d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					if userId > 0 {
						if len(community) > 0 {
							err = d.ExecSql(`INSERT INTO community (user_id) VALUES (?)`, userId)
							if err != nil {
								if d.dPrintSleep(err, d.sleepTime) {
									break BEGIN
								}
								continue BEGIN
							}
						}
						// генерим и шлем новую тр-ию
						err = d.SendTxChangePkey(userId)
						if err != nil {
							if d.dPrintSleep(err, d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
					} else {
						// если userId == 0, значит ключей в паблике больше нет и юзеру придется искать ключ самому
						continue
					}
				}

				// проверим, не прошла ли тр-ия и не сменился ли уже ключ
				userPubKeyCount, err := d.Single(`SELECT count(*) FROM ` + myPrefix + `my_keys WHERE status='approved'`).Int64()
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				logger.Debug("userPubKey: %v", userPubKeyCount)
				if userPubKeyCount > 1 {
					err = d.ExecSql(`UPDATE ` + myPrefix + `my_table SET status='user'`)
					if err != nil {
						if d.dPrintSleep(err, d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					// также, если это пул, то надо удалить приватный ключ из базы данных, т.к. взлом пула будет означать угон всех акков
					// хранение ключа на мобильном - безопасно, хранение ключа на ПК, пока Dcoin не стал популярен, тоже допустимо
					if len(community) > 0 {
						err = d.ExecSql(`DELETE private_key FROM ` + myPrefix + `my_keys`)
						if err != nil {
							if d.dPrintSleep(err, d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
					}

					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}

				lastTx, err := d.GetLastTx(uid, utils.TypesToIds([]string{"ChangePrimaryKey"}), 1, "2006-02-01 15:04:05")
				if err != nil {
					if d.dPrintSleep(err, d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				logger.Debug("lastTx: %v", lastTx)
				if len(lastTx) > 0 {
					if len(lastTx[0]["error"]) > 0 || utils.Time()-utils.StrToInt64(lastTx[0]["time_int"]) > 60 {
						// генерим и шлем новую тр-ию
						err = d.SendTxChangePkey(uid)
						if err != nil {
							if d.dPrintSleep(err, d.sleepTime) {
								break BEGIN
							}
							continue BEGIN
						}
					}
				}
			}
		}

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
