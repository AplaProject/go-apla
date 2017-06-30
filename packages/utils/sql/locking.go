package sql

import (
	"fmt"
	"regexp"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// CheckInstall waits for the end of the installation
func (db *DCDB) CheckInstall(DaemonCh chan bool, AnswerDaemonCh chan string, GoroutineName string) bool {
	// Возможна ситуация, когда инсталяция еще не завершена. База данных может быть создана, а таблицы еще не занесены
	// there could be the situation when installation is not over yet. Database could be created but tables are not inserted yet
	for {
		select {
		case <-DaemonCh:
			log.Debug("Restart from CheckInstall")
			AnswerDaemonCh <- GoroutineName
			return false
		default:
		}
		progress, err := db.Single("SELECT progress FROM install").String()
		if err != nil || progress != "complete" {
			// возможно попасть на тот момент, когда БД закрыта и идет скачивание готовой БД с сервера
			// the moment could happen when the database is closed and there is a download of the completed database from the server
			if ok, _ := regexp.MatchString(`database is closed`, fmt.Sprintf("%s", err)); ok {
				if DB != nil {
					db = DB
				}
			}
			//log.Debug("%v", `progress != "complete"`, db.GoroutineName)
			if err != nil {
				log.Error("%v", utils.ErrInfo(err))
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return true
}

// UpdMainLock updates the lock time
func (db *DCDB) UpdMainLock() error {
	return db.ExecSQL("UPDATE main_lock SET lock_time = ?", time.Now().Unix())
}

// CheckDaemonsRestart is reserved
func (db *DCDB) CheckDaemonsRestart() bool {
	return false
}

// DbLock locks deamons
func (db *DCDB) DbLock(DaemonCh chan bool, AnswerDaemonCh chan string, goRoutineName string) (bool, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()

	log.Debug("DbLock")
	var ok bool
	for {
		select {
		case <-DaemonCh:
			log.Debug("Restart from DbLock")
			AnswerDaemonCh <- goRoutineName
			return true, utils.ErrInfo("Restart from DbLock")
		default:
		}

		Mutex.Lock()

		log.Debug("DbLock Mutex.Lock()")

		exists, err := db.OneRow("SELECT lock_time, script_name FROM main_lock").String()
		if err != nil {
			Mutex.Unlock()
			return false, utils.ErrInfo(err)
		}
		if len(exists["script_name"]) == 0 {
			err = db.ExecSQL(`INSERT INTO main_lock(lock_time, script_name, info) VALUES(?, ?, ?)`, time.Now().Unix(), goRoutineName, utils.Caller(2))
			if err != nil {
				Mutex.Unlock()
				return false, utils.ErrInfo(err)
			}
			ok = true
		} else {
			t := converter.StrToInt64(exists["lock_time"])
			now := time.Now().Unix()
			if now-t > 600 {
				log.Error("%d %s %d", t, exists["script_name"], now-t)
				if utils.Mobile() {
					db.ExecSQL(`DELETE FROM main_lock`)
				}
			}
		}
		Mutex.Unlock()
		if !ok {
			time.Sleep(time.Duration(crypto.RandInt(300, 400)) * time.Millisecond)
		} else {
			break
		}
	}
	return false, nil
}

// DbUnlock unlocks database
func (db *DCDB) DbUnlock(goRoutineName string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()
	log.Debug("DbUnlock %v %v", utils.Caller(2), goRoutineName)
	affect, err := db.ExecSQLGetAffect("DELETE FROM main_lock WHERE script_name = ?", goRoutineName)
	log.Debug("main_lock affect: %d, goRoutineName: %s", affect, goRoutineName)
	if err != nil {
		log.Error("%s", utils.ErrInfo(err))
		return utils.ErrInfo(err)
	}
	return nil
}

// UpdDaemonTime is reserved
func (db *DCDB) UpdDaemonTime(name string) {

}
