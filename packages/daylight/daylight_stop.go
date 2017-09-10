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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

// Stop stops the program
func Stop() {
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("Apla Stop : %v", model.DBConn))
	stopDaemons := &model.StopDaemon{StopTime: time.Now().Unix()}
	err := stopDaemons.Create()
	if err != nil {
		logger.LogError(consts.SystemError, err)
	}
	logger.LogDebug(consts.DebugMessage, "Apla stop")
}
