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
	"net/http"
	"os"
	"runtime"

	"github.com/AplaProject/go-apla/packages/daylight"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/system"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
)

func main() {
	runtime.LockOSThread()

	var width uint = 900
	var height uint = 600
	var thrustWindow *window.Window
	if runtime.GOOS == "darwin" {
		height = 578
	}
	if utils.Desktop() && (winVer() >= 6 || winVer() == 0) {
		utils.Thrust = true
		thrust.Start()
		thrustWindow = thrust.NewWindow(thrust.WindowOptions{
			RootUrl:  "http://localhost:7979/loader.html",
			HasFrame: winVer() != 6,
			Title:    "AplaProject",
			Size:     commands.SizeHW{Width: width, Height: height},
		})

		thrust.NewEventHandler("*", func(cr commands.CommandResponse) {
			//			fmt.Println(fmt.Sprintf("======Event(%d) - Signaled by Command (%s)", cr.TargetID, cr.Type))
			if cr.TargetID > 1 && cr.Type == "closed" {
				if model.DBConn != nil {
					model.SetStopNow()
				} else {
					thrust.Exit()
					system.FinishThrust()
					os.Exit(0)
				}
			}
		})
		thrustWindow.Show()
		thrustWindow.Focus()
		go func() {
			http.ListenAndServe(":7979", nil)
		}()
	}
	tray()

	go daylight.Start("", thrustWindow)

	enterLoop()
	system.Finish()
}
