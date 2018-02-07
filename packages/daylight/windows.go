// +build windows

// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package daylight

import (
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

// KillPid kills the process with the specified pid
func KillPid(pid string) error {
	if model.DBConn != nil {
		sd := &model.StopDaemon{StopTime: time.Now().Unix()}
		err := sd.Create()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Error creating StopDaemon")
			return err
		}
	}
	rez, err := exec.Command("tasklist", "/fi", "PID eq "+pid).Output()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CommandExecutionError, "err": err, "cmd": "tasklist /fi PID eq" + pid}).Error("Error executing command")
		return err
	}
	if string(rez) == "" {
		return fmt.Errorf("null")
	}
	log.WithFields(log.Fields{"cmd": "tasklist /fi PID eq " + pid}).Debug("command execution result")
	if ok, _ := regexp.MatchString(`(?i)PID`, string(rez)); !ok {
		return fmt.Errorf("null")
	}
	return nil
}
