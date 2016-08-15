package daemons

import (
	"fmt"
	"github.com/democratic-coin/dcoin-go/packages/dcparser"
	"github.com/democratic-coin/dcoin-go/packages/utils"
)

func ElectionsAdmin(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "ElectionsAdmin"
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
			if d.unlockPrintSleep(utils.ErrInfo("blockId == 0"), d.sleepTime) {
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

		// проверим, прошло ли 2 недели с момента последнего обновления
		adminTime, err := d.Single("SELECT time FROM admin").Int64()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if curTime-adminTime <= variables.Int64["new_pct_period"] {
			if d.unlockPrintSleep(utils.ErrInfo("14 day error"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// сколько всего майнеров
		countMiners, err := d.Single("SELECT count(miner_id) FROM miners WHERE active  =  1").Int64()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		if countMiners < 1000 {
			if d.unlockPrintSleep(utils.ErrInfo("countMiners < 1000"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// берем все голоса
		var newAdmin int64
		votes_admin, err := d.GetMap(`
				SELECT	 admin_user_id,
							  count(user_id) as votes
				FROM votes_admin
				WHERE time > ?
				GROUP BY  admin_user_id
				`, "admin_user_id", "votes", curTime-variables.Int64["new_pct_period"])
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		for admin_user_id, votes := range votes_admin {
			// если более 50% майнеров проголосовали
			if utils.StrToInt64(votes) > countMiners/2 {
				newAdmin = utils.StrToInt64(admin_user_id)
			}
		}
		if newAdmin == 0 {
			if d.unlockPrintSleep(utils.ErrInfo("newAdmin == 0"), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		_, myUserId, _, _, _, _, err := d.TestBlock()
		forSign := fmt.Sprintf("%v,%v,%v,%v", utils.TypeInt("NewAdmin"), curTime, myUserId, newAdmin)
		binSign, err := d.GetBinSign(forSign, myUserId)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		data := utils.DecToBin(utils.TypeInt("NewAdmin"), 1)
		data = append(data, utils.DecToBin(curTime, 4)...)
		data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(myUserId))...)
		data = append(data, utils.EncodeLengthPlusData(utils.Int64ToByte(newAdmin))...)
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
			if d.unlockPrintSleep(err, d.sleepTime) {
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
