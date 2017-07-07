package sql

import (
	"context"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// WaitDB waits for the end of the installation
func WaitDB(ctx context.Context) (*DCDB, error) {
	// there could be the situation when installation is not over yet. Database could be created but tables are not inserted yet
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-tick.C:
			db := GetCurrentDB()
			if db != nil {
				if db.CheckDB() {
					return db, nil
				}
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (db *DCDB) CheckDB() bool {
	progress, err := db.Single("SELECT progress FROM install").String()
	if err != nil || progress != "complete" {
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		return false
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
func (db *DCDB) DbLock(ctx context.Context, goRoutineName string) (bool, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()

	ok, err := db.tryLock(goRoutineName)

	ticker := time.NewTicker(1 * time.Second)

	for ok || err != nil {
		select {
		case <-ticker.C:
			ok, err = db.tryLock(goRoutineName)
		case <-ctx.Done():
			return false, ctx.Err()
		}

	}
	return ok, err
}

func (db *DCDB) tryLock(goRoutineName string) (bool, error) {
	Mutex.Lock()
	defer Mutex.Unlock()

	exists, err := db.OneRow("SELECT lock_time, script_name FROM main_lock").String()
	if err != nil {
		return false, utils.ErrInfo(err)
	}

	if len(exists["script_name"]) == 0 {
		err = db.ExecSQL(`INSERT INTO main_lock(lock_time, script_name, info) VALUES(?, ?, ?)`, time.Now().Unix(), goRoutineName, utils.Caller(2))
		if err != nil {
			return false, utils.ErrInfo(err)
		}
		return true, nil

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
