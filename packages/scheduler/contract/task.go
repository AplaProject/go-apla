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

package contract

import (
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/scheduler"

	log "github.com/sirupsen/logrus"
)

// ContractHandler represents contract handler
type ContractHandler struct {
	Contract string
}

// Run executes task
func (ch *ContractHandler) Run(t *scheduler.Task) {
	_, err := NodeContract(ch.Contract)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ContractError, "error": err, "task": t.String(), "contract": ch.Contract}).Error("run contract task")
		return
	}

	log.WithFields(log.Fields{"task": t.String(), "contract": ch.Contract}).Info("run contract task")
}
