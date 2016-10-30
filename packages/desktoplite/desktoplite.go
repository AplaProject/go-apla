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

package main

import (
	//	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/EGaaS/go-egaas-mvp/packages/system"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
)

const GETPOOLURL = `http://node0.egaas.org/`

func main() {
	var (
		thrustWindow *window.Window
		//		mainWin      bool
	)

	runtime.LockOSThread()
	//	if utils.Desktop() && (winVer() >= 6 || winVer() == 0) {
	thrust.Start()

	thrust.NewEventHandler("*", func(cr commands.CommandResponse) {
		if cr.Type == "closed" {
			//			if mainWin || !isClosed {
			system.FinishThrust(0)
			os.Exit(0)
			/*			} else {
						close(chIntro)
						mainWin = true
					}*/
		}
	})

	thrustWindow = thrust.NewWindow(thrust.WindowOptions{
		Title:   "EGaaS Lite",
		RootUrl: GETPOOLURL,
		Size:    commands.SizeHW{Width: 1024, Height: 800},
	})
	/*	thrustWindow.HandleEvent("*", func(cr commands.EventResult) {
		fmt.Println("HandleEvent", cr)
	})*/
	if *utils.DevTools != 0 {
		thrustWindow.OpenDevtools()
	}
	thrustWindow.HandleRemote(func(er commands.EventResult, this *window.Window) {
		//		fmt.Println("RemoteMessage Recieved:", er.Message.Payload)
		if len(er.Message.Payload) > 7 && er.Message.Payload[:2] == `[{` {
			ioutil.WriteFile(filepath.Join(*utils.Dir, `accounts.txt`), []byte(er.Message.Payload), 0644)
		} else if er.Message.Payload == `ACCOUNTS` {
			accounts, _ := ioutil.ReadFile(filepath.Join(*utils.Dir, `accounts.txt`))
			this.SendRemoteMessage(string(accounts))
		} else {
			utils.ShellExecute(er.Message.Payload)
		}
	})
	thrustWindow.Show()
	thrustWindow.Focus()
	for {
		utils.Sleep(3600)
	}
	system.FinishThrust(0)
}
