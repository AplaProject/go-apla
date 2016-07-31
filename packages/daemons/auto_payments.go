package daemons

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func AutoPayments(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "AutoPayments"
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
		d.sleepTime = 600 // нужно время, чтобы предыдущий автоплатеж успел попасть в блокчейн
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

		myUsersIds, err := d.GetMyUsersIds(false, true)
		if len(myUsersIds) == 0 {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		// берем автоплатежи, которые пора провести
		rows, err := d.Query(d.FormatQuery(`
				SELECT id, sender
				FROM auto_payments
				WHERE sender IN (`+utils.JoinInt64Slice(myUsersIds, ",")+`) AND last_payment_time < ? - period
				LIMIT 1
				`), utils.Time())
		if err != nil {
			if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if ok := rows.Next(); ok {
			var auto_payment_id, sender int64
			err = rows.Scan(&auto_payment_id, &sender)
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// проверим, не нужно ли нам выйти, т.к. обновилась версия софта
			if CheckDaemonsRestart(chBreaker, chAnswer, GoroutineName) {
				rows.Close()
				utils.Sleep(1)
				break
			}

			myUsersIds, err := d.GetMyUsersIds(true, true)
			if err != nil {
				rows.Close()
				if d.dPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}

			// не наш ли это автоплатеж
			if utils.InSliceInt64(sender, myUsersIds) {
				curTime := utils.Time()
				forSign := fmt.Sprintf("%v,%v,%v,%v", utils.TypeInt("AutoPayment"), curTime, sender, auto_payment_id)
				binSign, err := d.GetBinSign(forSign, sender)
				if err != nil {
					rows.Close()
					if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
						break BEGIN
					}
					continue BEGIN
				}
				data := utils.DecToBin(utils.TypeInt("AutoPayment"), 1)
				data = append(data, utils.DecToBin(curTime, 4)...)
				data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(sender))...)
				data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(auto_payment_id))...)
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
		rows.Close()
		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)

}
