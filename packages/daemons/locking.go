package daemons

import (
	"context"
	"sync"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting install")
	}

	if install.Progress == model.ProgressComplete {
		return true
	}

	return false
}

// DBLock locks daemons
func DBLock() {
	mutex.Lock()
}

// DBUnlock unlocks database
func DBUnlock() {
	mutex.Unlock()
}
