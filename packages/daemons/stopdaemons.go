//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package daemons

import (
	"time"

	"github.com/GenesisCommunity/go-genesis/packages/daylight/system"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

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
		dExists, err := model.Single(`SELECT stop_time FROM stop_daemons`).Int64()
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
