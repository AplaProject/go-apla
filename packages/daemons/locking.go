package daemons

import (
	"context"
	"time"

	"sync"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

var mutex = sync.Mutex{}

// WaitDB waits for the end of the installation
func WaitDB(ctx context.Context) error {
	// There is could be the situation when installation is not over yet.
	// Database could be created but tables are not inserted yet

	if model.DBConn != nil && CheckDB() {
		return nil
	}

	// poll a base with period
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-tick.C:
			if model.DBConn != nil && CheckDB() {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// CheckDB check if installation complete or not
func CheckDB() bool {
	install := &model.Install{}

	err := install.Get()
	if err != nil {
		log.Errorf("%v", utils.ErrInfo(err))
	}

	if install.Progress == "complete" {
		return true
	}

	return false
}

// UpdMainLock updates the lock time
func UpdMainLock() error {
	return model.MainLockUpdate()
}

// DbLock locks daemons
func DbLock(ctx context.Context, goRoutineName string) (bool, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()

	ok, err := tryLock(goRoutineName)

	ticker := time.NewTicker(1 * time.Second)

	for !ok && err == nil {
		select {
		case <-ticker.C:
			ok, err = tryLock(goRoutineName)
		case <-ctx.Done():
			return false, ctx.Err()
		}

	}
	return ok, err
}

const MaxLockTime = 600

func tryLock(goRoutineName string) (bool, error) {
	mutex.Lock()
	defer mutex.Unlock()

	ml := model.MainLock{}
	err := ml.Get()

	// check for lock record and lock period
	if ml.LockTime == 0 {
		ml.LockTime = int32(time.Now().Unix())
		ml.ScriptName = goRoutineName
		ml.Info = utils.Caller(2)
		if err = ml.Save(); err != nil {
			return false, err
		}
		return true, nil

	} else {
		lockPeriod := time.Now().Unix() - int64(ml.LockTime)
		if lockPeriod > MaxLockTime {
			log.Error("%d %s %d", ml.LockTime, ml.ScriptName, lockPeriod)
			if utils.Mobile() {
				err = model.MainLockDelete(ml.ScriptName)
			}
		}
	}

	return false, err
}

// DbUnlock unlocks database
func DbUnlock(goRoutineName string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered", r)
			panic(r)
		}
	}()
	log.Debug("DbUnlock %v %v", utils.Caller(2), goRoutineName)
	if err := model.MainLockDelete(goRoutineName); err != nil {
		return utils.ErrInfo(err)
	}
	return nil
}
