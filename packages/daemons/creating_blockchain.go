package daemons

import (
	"os"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// CreatingBlockchain writes blockchain
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
		MonitorDaemonCh <- []string{GoroutineName, converter.Int64ToStr(time.Now().Unix())}

		// проверим, не нужно ли нам выйти из цикла
		// check if we have to break the cycle
		if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
			break BEGIN
		}

		curBlockID, err := d.GetBlockID()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// пишем свежие блоки в резервный блокчейн
		// record the newest blocks in reserve blockchain
		endBlockID, err := utils.GetEndBlockID()
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
		}
		logger.Debug("curBlockID: %v / endBlockID: %v", curBlockID, endBlockID)
		if curBlockID-consts.COUNT_BLOCK_BEFORE_SAVE > endBlockID {
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
					`), endBlockID, curBlockID-consts.COUNT_BLOCK_BEFORE_SAVE)
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
				blockData := append(converter.DecToBin(id, 5), converter.EncodeLengthPlusData(data)...)
				sizeAndData := append(converter.DecToBin(len(blockData), 5), blockData...)
				//err := ioutil.WriteFile(*utils.Dir+"/public/blockchain", append(sizeAndData, utils.DecToBin(len(sizeAndData), 5)...), 0644)
				if _, err = file.Write(append(sizeAndData, converter.DecToBin(len(sizeAndData), 5)...)); err != nil {
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
