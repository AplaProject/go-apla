package daemons

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func Stats(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	GoroutineName := "Stats"
	d := new(daemon)
	d.DCDB = DbConnect(chBreaker, chAnswer, GoroutineName)
	if d.DCDB == nil {
		return
	}
	d.goRoutineName = GoroutineName
	d.chAnswer = chAnswer
	d.chBreaker = chBreaker
	d.sleepTime = 86400
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

		curTime, err := d.Single(`SELECT time FROM info_block`).Int64()
		if utils.Time() - curTime > 86400 {
			// идет сбор БД из блокчейна
			d.sleepTime = 60
		} else {
			d.sleepTime = 86400
		}
		t := time.Unix(curTime, 0)

		variables, err := d.GetAllVariables()
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		CurrencyList, err := d.GetCurrencyList(false)
		if err != nil {
			if d.dPrintSleep(err, d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		for currencyId, _ := range CurrencyList {
			sumPromisedAmount, err := d.Single(`
				SELECT sum(amount) as sum_amount
				FROM promised_amount
				WHERE status = 'mining' AND
							 del_block_id = 0 AND
							(cash_request_out_time = 0 OR cash_request_out_time > ?) AND
							currency_id = ?`, utils.Time()-variables.Int64["cash_request_time"], currencyId).Float64()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			// получаем кол-во DC на кошельках
			sumWallets, err := d.Single("SELECT sum(amount) as sum_amount FROM dlt_wallets WHERE currency_id = ?", currencyId).Float64()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			// получаем кол-во TDC на обещанных суммах, плюсуем к тому, что на кошельках
			sumTdc, err := d.Single("SELECT sum(tdc_amount) as sum_amount FROM promised_amount WHERE currency_id = ?", currencyId).Float64()
			if err != nil {
				if d.dPrintSleep(err, d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
			sumWallets += sumTdc

			err = d.ExecSql(`INSERT INTO stats (day, month, year, currency_id, dc, promised_amount) VALUES (?, ?, ?, ?, ?, ?)`, t.Day(), int(t.Month()), t.Year(), currencyId, int64(sumWallets), int64(sumPromisedAmount))
			if err != nil {
				logger.Debug("%v", err)
			}
		}
		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
