// +build !ios,!android

package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/daylight"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/system"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
	"net/http"
	"os"
	"runtime"
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
		thrust.Start()
		thrustWindow = thrust.NewWindow(thrust.WindowOptions{
			RootUrl:  "http://localhost:8989/loader.html",
			HasFrame: winVer() != 6,
			Title:    "DayLight",
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
			http.ListenAndServe(":8989", nil)
		}()
	}
	tray()

	go daylight.Start("", thrustWindow)

	enterLoop()
	system.Finish(0)
}
