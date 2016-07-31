package daemons

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"os"
	"fmt"
)

func CleaningDb(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "CleaningDb"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	if utils.Mobile() {
		d.sleepTime = 1800
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

		curBlockId, err := d.GetBlockId()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// пишем свежие блоки в резервный блокчейн
		endBlockId, err := utils.GetEndBlockId()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			// чтобы не стопориться тут, а дойти до пересборки БД
			endBlockId = 4294967295
		}
		logger.Debug("curBlockId: %v / endBlockId: %v", curBlockId, endBlockId)
		if curBlockId-30 > endBlockId {
			file, err := os.OpenFile(*utils.Dir+"/public/blockchain", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			rows, err := d.Query(d.FormatQuery(`
					SELECT id, data
					FROM block_chain
					WHERE id > ? AND id <= ?
					ORDER BY id
					`), endBlockId, curBlockId-30)
			if err != nil {
				file.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			for rows.Next() {
				var id, data string
				err = rows.Scan(&id, &data)
				if err != nil {
					rows.Close()
					file.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				blockData := append(utils.DecToBin(id, 5), utils.EncodeLengthPlusData(data)...)
				sizeAndData := append(utils.DecToBin(len(blockData), 5), blockData...)
				//err := ioutil.WriteFile(*utils.Dir+"/public/blockchain", append(sizeAndData, utils.DecToBin(len(sizeAndData), 5)...), 0644)
				if _, err = file.Write(append(sizeAndData, utils.DecToBin(len(sizeAndData), 5)...)); err != nil {
					rows.Close()
					file.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				if err != nil {
					rows.Close()
					file.Close()
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
			}
			rows.Close()
			file.Close()
		}

		autoReload, err := d.Single("SELECT auto_reload FROM config").Int64()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		logger.Debug("autoReload: %v", autoReload)
		if autoReload < 60 {
			if d.dPrintSleep(utils.ErrInfo("autoReload < 60"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// если main_lock висит более x минут, значит был какой-то сбой
		mainLock, err := d.Single("SELECT lock_time FROM main_lock WHERE script_name NOT IN ('my_lock', 'cleaning_db')").Int64()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		var infoBlockRestart bool
		// если с main_lock всё норм, то возможно, что новые блоки не собираются из-за бана нодов
		if mainLock == 0 || utils.Time()-autoReload < mainLock {
			timeInfoBlock, err := d.Single(`SELECT time FROM info_block`).Int64()
			if err != nil {
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			if utils.Time()-timeInfoBlock > autoReload {
				// подождем 5 минут и проверим еще раз
				if d.dSleep(300) {
					break BEGIN
				}
				newTimeInfoBlock, err := d.OneRow(`SELECT block_id, time FROM info_block`).Int64()
				if err != nil {
					if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				// Если за 5 минут info_block тот же, значит обновление блокчейна не идет
				if newTimeInfoBlock["time"] == timeInfoBlock {
					infoBlockRestart = true
					logger.Error("infoBlockRestart %d / %d", newTimeInfoBlock["block_id"], newTimeInfoBlock["time"])
				}
			}
		}
		logger.Debug("mainLock: %v", mainLock)
		logger.Debug("utils.Time(): %v", utils.Time())
		if (mainLock > 0 && utils.Time()-autoReload > mainLock) || infoBlockRestart {

			// ClearDb - убивает демонов, чистит БД, а потом заново запускает демонов
			// не забываем, что это тоже демон и он должен отчитаться о завершении
			err = ClearDb(d.chAnswer, GoroutineName)
			if err != nil {
				fmt.Println(utils.ErrInfo(err))
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
