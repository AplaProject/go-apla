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
	"os/exec"
	"runtime"
	"strconv"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"

	log "github.com/sirupsen/logrus"
)

func openBrowser(BrowserHTTPHost string) {
	var err error
	cmd := ""
	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		err = exec.Command("xdg-open", BrowserHTTPHost).Start()
	case "windows", "darwin":
		cmd = "open"
		err = exec.Command("open", BrowserHTTPHost).Start()
		if err != nil {
			cmd = "cmd /c start"
			exec.Command("cmd", "/c", "start", BrowserHTTPHost).Start()
		}
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.WithFields(log.Fields{"command": cmd, "type": consts.CommandExecutionError, "error": err}).Error("Error executing command opening browser")
	}
}

// GetHTTPHost returns program's hosts
func GetHTTPHost() (string, string) {
	// !!!
	// BrowserHTTPHost := "http://localhost:" + *utils.ListenHTTPPort
	// HandleHTTPHost := ""
	// ListenHTTPHost := ":" + *utils.ListenHTTPPort
	// if len(*utils.TCPHost) > 0 {
	// 	fmt.Println(*utils.TCPHost)
	// 	ListenHTTPHost = *utils.TCPHost + ":" + *utils.ListenHTTPPort
	// 	BrowserHTTPHost = "http://" + *utils.TCPHost + ":" + *utils.ListenHTTPPort
	// }
	host := "localhost"
	if len(conf.Config.API.Host) > 0 && conf.Config.API.Host != "0.0.0.0" {
		host = conf.Config.API.Host
	}

	BrowserHTTPHost := "http://" + host + ":" + strconv.Itoa(conf.Config.API.Port)
	ListenHTTPHost := conf.Config.API.Str()

	return BrowserHTTPHost, ListenHTTPHost
}
