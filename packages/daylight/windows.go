// +build windows

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
	//"os/exec"
	//"fmt"
	//"os"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
)

// KillPid kills the process with the specified pid
func KillPid(pid string) error {
	if sql.DB != nil && sql.DB.DB != nil {
		err := sql.DB.ExecSQL(`INSERT INTO stop_daemons(stop_time) VALUES (?)`, utils.Time())
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
			return err
		}
	}
	//var rez []byte
	/*file, err := os.OpenFile("kill", os.O_APPEND|os.O_WRONLY|os.O_CREATE,0600)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString("1")
	*/
	/*err := exec.Command("taskkill","/pid", pid).Start()
	if err!=nil {
		return err
	}*/
	rez, err := exec.Command("tasklist", "/fi", "PID eq "+pid).Output()
	if err != nil {
		return err
	}
	if string(rez) == "" {
		return fmt.Errorf("null")
	}
	log.Debug("%rez s", string(rez))
	fmt.Println("rez", string(rez))
	if ok, _ := regexp.MatchString(`(?i)PID`, string(rez)); !ok {
		return fmt.Errorf("null")
	}
	//fmt.Printf("taskkill /pid %s: %s\n", pid, rez)
	return nil
}

func tray() {

}
