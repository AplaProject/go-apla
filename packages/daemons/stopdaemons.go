// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package daemons

import (
	"time"

	"github.com/AplaProject/go-apla/packages/daylight/system"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
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
			err := model.Delete(nil, "stop_daemons", "")
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting from stop daemons")
			}
			first = true
		}
		dExists, err := model.Single(nil, `SELECT stop_time FROM stop_daemons`).Int64()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting stop_time from StopDaemons")
		}
		if dExists > 0 {
			utils.CancelFunc()
			for i := 0; i < utils.DaemonsCount; i++ {
				name := <-utils.ReturnCh
				log.WithFields(log.Fields{"daemon_name": name}).Debug("daemon stopped")
			}

			err := model.GormClose()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("gorm close")
			}
			err = system.RemovePidFile()
			if err != nil {
				log.WithFields(log.Fields{
					"type": consts.IOError, "error": err,
				}).Error("removing pid file")
				panic(err)
			}
		}
		time.Sleep(time.Second)
	}
}
