// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package daemons

import (
	"os"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/system"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

// WaitStopTime closes the database and stop daemons
func WaitStopTime() {
	var first bool
	for {
		if model.DBConn == nil {
			time.Sleep(time.Second * 3)
			continue
		}
		if !first {
			err := model.Delete("stop_daemons", "")
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting from stop daemons")
			}
			first = true
		}
		sd := model.StopDaemon{}
		_, err := sd.Get()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting stop_time from StopDaemons")
		}
		if sd.StopTime > 0 {
			utils.CancelFunc()
			for i := 0; i < utils.DaemonsCount; i++ {
				name := <-utils.ReturnCh
				log.WithFields(log.Fields{"daemon_name": name}).Debug("daemon stopped")
			}
			system.FinishThrust()

			err := model.GormClose()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("gorm close")
			}
			err = os.Remove(*utils.Dir + "/daylight.pid")
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err, "path": *utils.Dir + "/daylight.pid"}).Error("removing file")
				panic(err)
			}
		}
		time.Sleep(time.Second)
	}
}
