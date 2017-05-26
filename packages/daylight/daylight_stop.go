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

package daylight

import (
	"fmt"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)


func Stop() {
	log.Debug("Stop()")
	IosLog("Stop()")
	var err error
	utils.DB, err = utils.NewDbConnect(configIni)
	log.Debug("DayLight Stop : %v", utils.DB)
	IosLog("utils.DB:" + fmt.Sprintf("%v", utils.DB))
	if err != nil {
		IosLog("err:" + fmt.Sprintf("%s", utils.ErrInfo(err)))
		log.Error("%v", utils.ErrInfo(err))
		//panic(err)
		//os.Exit(1)
	}
	err = utils.DB.ExecSQL(`INSERT INTO stop_daemons(stop_time) VALUES (?)`, utils.Time())
	if err != nil {
		IosLog("err:" + fmt.Sprintf("%s", utils.ErrInfo(err)))
		log.Error("%v", utils.ErrInfo(err))
	}
	log.Debug("DayLight Stop")
	IosLog("DayLight Stop")
}
