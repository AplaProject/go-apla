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
	"os/signal"
	"syscall"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

/*
#include <stdio.h>
#include <signal.h>

extern void go_callback_int();
static inline void SigBreak_Handler(int n_signal){
    printf("closed\n");
	go_callback_int();
}
static inline void waitSig() {
    #if (WIN32 || WIN64)
    signal(SIGBREAK, &SigBreak_Handler);
    signal(SIGINT, &SigBreak_Handler);
    #endif
}
*/
import (
	"C"
)

//export go_callback_int
func go_callback_int() {
	SigChan <- syscall.Signal(1)
}

// SigChan is a channel
var SigChan chan os.Signal

func waitSig() {
	C.waitSig()
}

// WaitForSignals waits for Interrupt os.Kill signals
func WaitForSignals() {
	SigChan = make(chan os.Signal, 1)
	waitSig()
	var Term os.Signal = syscall.SIGTERM
	go func() {
		signal.Notify(SigChan, os.Interrupt, os.Kill, Term)
		<-SigChan

		if utils.CancelFunc != nil {
			utils.CancelFunc()
			for i := 0; i < utils.DaemonsCount; i++ {
				name := <-utils.ReturnCh
				log.WithFields(log.Fields{"daemon_name": name}).Debug("daemon stopped")
			}

			log.Debug("Daemons killed")
		}

		if model.DBConn != nil {
			err := model.GormClose()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("closing gorm")
			}
		}

		err := os.Remove(conf.Config.GetPidPath())
		if err != nil {
			log.WithFields(log.Fields{
				"type": consts.IOError, "error": err, "path": conf.Config.GetPidPath(),
			}).Error("removing file")
		}

		os.Exit(1)

	}()
}
