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
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/EGaaS/go-mvp/packages/daylight"
	"github.com/EGaaS/go-mvp/packages/static"
	"github.com/EGaaS/go-mvp/packages/system"
	"github.com/EGaaS/go-mvp/packages/utils"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
)

func main_loader(w http.ResponseWriter, r *http.Request) {
	data, _ := static.Asset("static/img/main_loader.gif")
	fmt.Fprint(w, string(data))
}
func main_loader_html(w http.ResponseWriter, r *http.Request) {
	html := `<html><title>DayLight</title><body style="margin:0;padding:0;overflow:hidden;"><img src="static/img/main_loader.gif"/></body></html>`
	fmt.Fprint(w, html)
}
func main() {
	runtime.LockOSThread()

	var width uint = 800
	var height uint = 600
	var thrustWindow *window.Window
	if runtime.GOOS == "darwin" {
		height = 578
	}
	if utils.Desktop() && (winVer() >= 6 || winVer() == 0) {
		utils.Thrust = true
		thrust.Start()
		thrustWindow = thrust.NewWindow(thrust.WindowOptions{
			RootUrl:  "http://localhost:7079/loader.html",
			HasFrame: winVer() != 6,
			Title:    "EGaaS",
			Size:     commands.SizeHW{Width: width, Height: height},
		})

		thrust.NewEventHandler("*", func(cr commands.CommandResponse) {
			//fmt.Println(fmt.Sprintf("======Event(%s %d) - Signaled by Command (%s)", cr.TargetID, cr.Type))
			if cr.TargetID > 1 && cr.Type == "closed" {
				if utils.DB != nil && utils.DB.DB != nil {
					utils.DB.ExecSql(`INSERT INTO stop_daemons(stop_time) VALUES (?)`, utils.Time())
				} else {
					system.FinishThrust(0)
					os.Exit(0)
				}
			}
		})
		thrustWindow.Show()
		thrustWindow.Focus()
		go func() {
			http.HandleFunc("/static/img/main_loader.gif", main_loader)
			http.HandleFunc("/loader.html", main_loader_html)
			http.ListenAndServe(":7079", nil)
		}()
	}
	tray()

	go daylight.Start("", thrustWindow)

	enterLoop()
	system.Finish(0)
}
