package daemons

import (
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"os"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
)

func CreatingBlockchain(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "CreatingBlockchain"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	d.sleepTime = 10
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
		endBlockId, err := utils.GetEndBlockID()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
		}
		logger.Debug("curBlockId: %v / endBlockId: %v", curBlockId, endBlockId)
		if curBlockId-consts.COUNT_BLOCK_BEFORE_SAVE > endBlockId {
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
					`), endBlockId, curBlockId-consts.COUNT_BLOCK_BEFORE_SAVE )
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

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
