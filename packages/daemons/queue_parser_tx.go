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
	"context"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/parser"

	log "github.com/sirupsen/logrus"
)

// QueueParserTx parses transaction from the queue
func QueueParserTx(ctx context.Context, d *daemon) error {
	DBLock()
	defer DBUnlock()

	infoBlock := &model.InfoBlock{}
	_, err := infoBlock.Get()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting info block")
		return err
	}
	if infoBlock.BlockID == 0 {
		d.logger.Debug("no blocks for parsing")
		return nil
	}

	// delete looped transactions
	_, err = model.DeleteLoopedTransactions()
	if err != nil {
		d.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting looped transactions")
		return err
	}

	p := new(parser.Parser)
	err = p.AllTxParser()
	if err != nil {
		d.logger.WithFields(log.Fields{"error": err}).Error("parsing transactions")
		return err
	}

	return nil
}
