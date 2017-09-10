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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func openBrowser(BrowserHTTPHost string) {
	logger.LogDebug(consts.FuncStarted, fmt.Sprintf("runtime.GOOS: %v", runtime.GOOS))
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
		logger.LogError(consts.CommandError, err)
	}
}

// GetHTTPHost returns program's hosts
func GetHTTPHost() (string, string, string) {
	logger.LogDebug(consts.FuncStarted, "")
	BrowserHTTPHost := "http://localhost:" + *utils.ListenHTTPPort
	HandleHTTPHost := ""
	ListenHTTPHost := ":" + *utils.ListenHTTPPort
	if len(*utils.TCPHost) > 0 {
		ListenHTTPHost = *utils.TCPHost + ":" + *utils.ListenHTTPPort
		BrowserHTTPHost = "http://" + *utils.TCPHost + ":" + *utils.ListenHTTPPort
	}
	return BrowserHTTPHost, HandleHTTPHost, ListenHTTPHost
}
