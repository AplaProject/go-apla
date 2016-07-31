package daemons

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
	//"log"
	"os"
)

/*
 * Если наш miner_id есть среди тех, кто должен скачать фото нового майнера к себе, то качаем
 */

func NodeVoting(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "NodeVoting"
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

		err, restart := d.dbLock()
		if restart {
			break BEGIN
		}
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// берем данные, которые находятся на голосовании нодов
		rows, err := d.Query(d.FormatQuery(`
				SELECT miners_data.user_id,
							 http_host as host,
							 pool_user_id,
							 face_hash,
							 profile_hash,
							 photo_block_id,
							 photo_max_miner_id,
							 miners_keepers,
							 id as vote_id,
							 miner_id
				FROM votes_miners
				LEFT JOIN miners_data
						 ON votes_miners.user_id = miners_data.user_id
				WHERE cron_checked_time < ? AND
							 votes_end = 0 AND
							 type = 'node_voting'
				`), utils.Time()-consts.CRON_CHECKED_TIME_SEC)
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if ok := rows.Next(); ok {
			var vote_id, miner_id, pool_user_id int64
			var user_id, host,row_face_hash, row_profile_hash, photo_block_id, photo_max_miner_id, miners_keepers string
			err = rows.Scan(&user_id, &host, &pool_user_id, &row_face_hash, &row_profile_hash, &photo_block_id, &photo_max_miner_id, &miners_keepers, &vote_id, &miner_id)
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// проверим, не нужно нам выйти, т.к. обновилась версия софта
			if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
				rows.Close()
				utils.Sleep(1)
				break
			}
			minersIds := utils.GetMinersKeepers(photo_block_id, photo_max_miner_id, miners_keepers, true)
			myUsersIds, err := d.GetMyUsersIds(true, true)
			myMinersIds, err := d.GetMyMinersIds(myUsersIds)

			// нет ли нас среди тех, кто должен скачать фото к себе и проголосовать
			var intersectMyMiners []int64
			for _, id := range minersIds {
				if utils.InSliceInt64(int64(id), myMinersIds) {
					intersectMyMiners = append(intersectMyMiners, int64(id))
				}
			}
			var vote int64
			if len(intersectMyMiners) > 0 {
				var downloadError bool
				var faceHash, profileHash string
				var faceFile, profileFile []byte

				if pool_user_id > 0 {
					host, err = d.Single(`SELECT http_host FROM miners_data WHERE user_id = ?`, pool_user_id).String()
					if err != nil {
						rows.Close()
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
				}

				// копируем фото  к себе
				profilePath := *utils.Dir + "/public/profile_" + user_id + ".jpg"
				_, err = utils.DownloadToFile(host+"/public/"+user_id+"_user_profile.jpg", profilePath, 60, chBreaker, chAnswer, GoroutineName)
				if err != nil {
					logger.Error("%s", utils.ErrInfo(err))
					downloadError = true
				}
				facePath := *utils.Dir + "/public/face_" + user_id + ".jpg"
				_, err = utils.DownloadToFile(host+"/public/"+user_id+"_user_face.jpg", facePath, 60, chBreaker, chAnswer, GoroutineName)
				if err != nil {
					logger.Error("%s", utils.ErrInfo(err))
					downloadError = true
				}
				if !downloadError {
					// хэши скопированных фото
					profileFile, err = ioutil.ReadFile(profilePath)
					if err != nil {
						rows.Close()
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					profileHash = string(utils.DSha256(profileFile))
					logger.Info("%v", "profileHash", profileHash)
					faceFile, err = ioutil.ReadFile(facePath)
					if err != nil {
						rows.Close()
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					faceHash = string(utils.DSha256(faceFile))
					logger.Info("%v", "faceHash", faceHash)
				}
				// проверяем хэш. Если сходится, то голосуем за, если нет - против и размер не должен быть более 200 Kb.
				if profileHash == row_profile_hash && faceHash == row_face_hash && len(profileFile) < 204800 && len(faceFile) < 204800 {
					vote = 1
				} else {
					logger.Error("%s %s %s %s %d %d", profileHash, row_face_hash, faceHash, row_profile_hash, len(profileFile), len(faceFile))
					vote = 0 // если хэш не сходится, то удаляем только что скаченное фото
					os.Remove(profilePath)
					os.Remove(facePath)
				}

				// проходимся по всем нашим майнерам, если это пул и по одному, если это сингл-мод
				for _, myMinerId := range intersectMyMiners {

					myUserId, err := d.Single("SELECT user_id FROM miners_data WHERE miner_id  =  ?", myMinerId).Int64()
					if err != nil {
						rows.Close()
						if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}

					curTime := utils.Time()

					forSign := fmt.Sprintf("%v,%v,%v,%v,%v", utils.TypeInt("VotesNodeNewMiner"), curTime, myUserId, vote_id, vote)
					binSign, err := d.GetBinSign(forSign, myUserId)
					if err != nil {
						rows.Close()
						if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}
					data := utils.DecToBin(utils.TypeInt("VotesNodeNewMiner"), 1)
					data = append(data, utils.DecToBin(curTime, 4)...)
					data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(myUserId))...)
					data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(vote_id))...)
					data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(vote))...)
					data = append(data, utils.EncodeLengthPlusData([]byte(binSign))...)

					err = d.InsertReplaceTxInQueue(data)
					if err != nil {
						rows.Close()
						if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
							break BEGIN
						}
						continue BEGIN
					}

				}
			}

			// отмечаем, чтобы больше не брать эту строку
			err = d.ExecSql("UPDATE votes_miners SET cron_checked_time = ? WHERE id = ?", utils.Time(), vote_id)
			if err != nil {
				rows.Close()
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}
		rows.Close()
		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)

}
