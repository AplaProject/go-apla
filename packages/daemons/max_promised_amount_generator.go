package daemons

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/*
 * Каждые 2 недели собираем инфу о голосах за max_promised_amount и создаем тр-ию, которая
 * попадет в DC сеть только, если мы окажемся генератором блока
 * */

func MaxPromisedAmountGenerator(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "MaxPromisedAmountGenerator"
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

		blockId, err := d.GetBlockId()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if blockId == 0 {
			if d.unlockPrintSleep(utils.ErrInfo("blockId == 0"), 1) {
				break BEGIN
			}
			continue BEGIN
		}

		_, _, myMinerId, _, _, _, err := d.TestBlock()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		// а майнер ли я ?
		if myMinerId == 0 {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		variables, err := d.GetAllVariables()
		curTime := utils.Time()

		totalCountCurrencies, err := d.GetCountCurrencies()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		// проверим, прошло ли 2 недели с момента последнего обновления
		pctTime, err := d.Single("SELECT max(time) FROM max_promised_amounts").Int64()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if curTime-pctTime <= variables.Int64["new_max_promised_amount"] {
			if d.unlockPrintSleep(utils.ErrInfo("14 day error"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// берем все голоса
		maxPromisedAmountVotes := make(map[int64][]map[int64]int64)
		rows, err := d.Query("SELECT currency_id, amount, count(user_id) as votes FROM votes_max_promised_amount GROUP BY currency_id, amount ORDER BY currency_id, amount ASC")
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		for rows.Next() {
			var currency_id, amount, votes int64
			err = rows.Scan(&currency_id, &amount, &votes)
			if err != nil {
				rows.Close()
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			maxPromisedAmountVotes[currency_id] = append(maxPromisedAmountVotes[currency_id], map[int64]int64{amount: votes})
			//fmt.Println("currency_id", currency_id)
		}
		rows.Close()

		NewMaxPromisedAmountsVotes := make(map[string]int64)
		for currencyId, amountsAndVotes := range maxPromisedAmountVotes {
			NewMaxPromisedAmountsVotes[utils.Int64ToStr(currencyId)] = utils.GetMaxVote(amountsAndVotes, 0, totalCountCurrencies, 10)
		}

		jsonData, err := json.Marshal(NewMaxPromisedAmountsVotes)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		_, myUserId, _, _, _, _, err := d.TestBlock()
		forSign := fmt.Sprintf("%v,%v,%v,%s", utils.TypeInt("NewMaxPromisedAmounts"), curTime, myUserId, jsonData)
		logger.Debug("forSign = %v", forSign)
		binSign, err := d.GetBinSign(forSign, myUserId)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		data := utils.DecToBin(utils.TypeInt("NewMaxPromisedAmounts"), 1)
		data = append(data, utils.DecToBin(curTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(myUserId))...)
		data = append(data, utils.EncodeLengthPlusData(jsonData)...)
		data = append(data, utils.EncodeLengthPlusData([]byte(binSign))...)

		err = d.InsertReplaceTxInQueue(data)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		p := new(dcparser.Parser)
		p.DCDB = d.DCDB
		err = p.TxParser(utils.HexToBin(utils.Md5(data)), data, true)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)

}
