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
	//	_ "image/png"
	"os/exec"
	"runtime"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/op/go-logging"
)

var (
	log       = logging.MustGetLogger("daylight")
	format    = logging.MustStringFormatter("%{color}%{time:15:04:05.000} %{shortfile} %{shortfunc} [%{level:.4s}] %{color:reset} %{message}[" + consts.VERSION + "]" + string(byte(0)))
	configIni map[string]string
)

func openBrowser(BrowserHTTPHost string) {
	log.Debug("runtime.GOOS: %v", runtime.GOOS)
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", BrowserHTTPHost).Start()
	case "windows", "darwin":
		err = exec.Command("open", BrowserHTTPHost).Start()
		if err != nil {
			exec.Command("cmd", "/c", "start", BrowserHTTPHost).Start()
		}
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Error("%v", err)
	}
}

func GetHTTPHost() (string, string, string) {
	BrowserHTTPHost := "http://localhost:" + *utils.ListenHttpPort
	HandleHTTPHost := ""
	ListenHTTPHost := ":" + *utils.ListenHttpPort
	if len(*utils.TcpHost) > 0 {
		ListenHTTPHost = *utils.TcpHost + ":" + *utils.ListenHttpPort
		BrowserHTTPHost = "http://" + *utils.TcpHost + ":" + *utils.ListenHttpPort
	}
	return BrowserHTTPHost, HandleHTTPHost, ListenHTTPHost
}
