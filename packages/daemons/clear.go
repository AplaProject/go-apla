package daemons

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/*
если табла была очищена, то в parser.rollbackAI будет 0, поэтому нелья чистить таблы под нуль
*/
func Clear(chBreaker chan bool, chAnswer chan string) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("daemon Recovered", r)
			panic(r)
		}
	}()

	const GoroutineName = "Clear"
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
		logger.Debug("blockId: %d", blockId)
		variables, err := d.GetAllVariables()
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// чистим log_transactions каждые 15 минут. Удаляем данные, которые старше 36 часов.
		// Можно удалять и те, что старше rollback_blocks_2 + погрешность для времени транзакции (5-15 мин),
		// но пусть будет 36 ч. - с хорошим запасом.

		err = d.ExecSql("DELETE FROM log_transactions WHERE time < ?", utils.Time()-86400*3)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// через rollback_blocks_2 с запасом 14400 блоков чистим таблу log_votes где есть del_block_id
		// при этом, если проверяющих будет мало, то табла может захламиться незаконченными голосованиями
		err = d.ExecSql("DELETE FROM log_votes WHERE del_block_id < ? AND del_block_id > 0", blockId-variables.Int64["rollback_blocks_2"]-14400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// через 14400 блоков чистим таблу wallets_buffer где есть del_block_id
		err = d.ExecSql("DELETE FROM dlt_wallets_buffer WHERE del_block_id < ? AND del_block_id > 0", blockId-variables.Int64["rollback_blocks_2"]-14400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		// чистим все _log_time_
		err = d.ExecSql("DELETE FROM log_time_votes_complex WHERE time < ?", utils.Time()-variables.Int64["limit_votes_complex_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_commission WHERE time < ?", utils.Time()-variables.Int64["limit_commission_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_change_host WHERE time < ?", utils.Time()-variables.Int64["limit_change_host_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_votes_miners WHERE time < ?", utils.Time()-variables.Int64["limit_votes_miners_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_primary_key WHERE time < ?", utils.Time()-variables.Int64["limit_primary_key_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_node_key WHERE time < ?", utils.Time()-variables.Int64["limit_node_key_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_mining WHERE time < ?", utils.Time()-variables.Int64["limit_mining_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_message_to_admin WHERE time < ?", utils.Time()-variables.Int64["limit_message_to_admin_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_holidays WHERE time < ?", utils.Time()-variables.Int64["limit_holidays_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_change_geolocation WHERE time < ?", utils.Time()-variables.Int64["limit_change_geolocation_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_cash_requests WHERE time < ?", utils.Time()-variables.Int64["limit_cash_requests_out_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_promised_amount WHERE time < ?", utils.Time()-variables.Int64["limit_promised_amount_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_abuses WHERE time < ?", utils.Time()-variables.Int64["limit_abuses_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_new_miner WHERE time < ?", utils.Time()-variables.Int64["limit_new_miner_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_votes WHERE time < ?", utils.Time()-86400-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_votes_nodes WHERE time < ?", utils.Time()-variables.Int64["node_voting_period"]-86400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_wallets WHERE block_id < ? AND block_id > 0", blockId-variables.Int64["rollback_blocks_2"]-14400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}
		err = d.ExecSql("DELETE FROM log_time_money_orders WHERE del_block_id < ? AND del_block_id > 0", blockId-variables.Int64["rollback_blocks_2"]-14400)
		if err != nil {
			if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
				break BEGIN
			}
			continue BEGIN
		}

		arr := []string{"log_commission",
			"log_faces",
			"log_forex_orders",
			"log_forex_orders_main",
			"log_miners",
			"log_miners_data",
			"log_points",
			"log_promised_amount",
			"log_recycle_bin",
			"log_spots_compatibility",
			"log_users",
			"log_votes_max_other_currencies",
			"log_votes_max_promised_amount",
			"log_votes_miner_pct",
			"log_votes_reduction",
			"log_votes_user_pct",
			"log_wallets"}
		for _, table := range arr {
			err = d.ExecSql("DELETE FROM "+table+" WHERE block_id < ? AND block_id > 0", blockId-variables.Int64["rollback_blocks_2"]-14400)
			if err != nil {
				if d.unlockPrintSleep(utils.ErrInfo(err), d.sleepTime) {
					break BEGIN
				}
				continue BEGIN
			}
		}
		logger.Debug("variables.Int64[rollback_blocks_2]: %v", variables.Int64["rollback_blocks_2"])

		d.dbUnlock()

		if d.dSleep(d.sleepTime) {
			break BEGIN
		}
	}
	logger.Debug("break BEGIN %v", GoroutineName)
}
