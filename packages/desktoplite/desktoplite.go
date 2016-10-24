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
	"github.com/EGaaS/go-mvp/packages/system"
	"github.com/EGaaS/go-mvp/packages/utils"
	"github.com/go-bindata-assetfs"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/thrust"
	"os"
	"runtime"
	"path/filepath"
	"io/ioutil"
	"encoding/json"
	"github.com/EGaaS/go-mvp/packages/static"
)

//const GETPOOLURL = `http://localhost:8089/getpool/`
const GETPOOLURL = `http://getpool.daylight.world/`

type Pool struct {
	Pool string `json:"pool"`
	UserId int64 `json:"user_id"`
}

func main() {
	var ( thrustWindow *window.Window
		pool Pool
		mainWin bool 
	)

	runtime.LockOSThread()
	//	if utils.Desktop() && (winVer() >= 6 || winVer() == 0) {
	thrust.Start()

	dir,_ := filepath.Abs(filepath.Dir(os.Args[0]))
	userfile := filepath.Join(dir, `iduser.txt`)
	txtUser,_ := ioutil.ReadFile(userfile)
	idUser := utils.StrToInt64(string(txtUser))

	chIntro := make(chan bool)
	var  isClosed bool
	
	thrust.NewEventHandler("*", func(cr commands.CommandResponse) {
		if cr.Type == "closed" {
			if mainWin || !isClosed {
				system.FinishThrust(0)
				os.Exit(0)
			} else {
				close( chIntro )
				mainWin = true
			}
		}
	})

	if idUser == 0 {
		introWindow := thrust.NewWindow(thrust.WindowOptions{
			RootUrl: `http://localhost:8990`,
			Title : "DayLight Lite",
			Size:    commands.SizeHW{Width: 1024, Height: 600},
		})
		introWindow.HandleRemote(func(er commands.EventResult, this *window.Window) {
			if  len(er.Message.Payload) > 7 && er.Message.Payload[:7]==`PUBLIC=` {
				json.Unmarshal( []byte(er.Message.Payload[7:]), &pool )
				if pool.UserId != 0 {
					err := ioutil.WriteFile( userfile, []byte(utils.Int64ToStr(pool.UserId)), 0644 )
					if err != nil {
						fmt.Println( `Error`, err )
					}
				}
				fmt.Println(`Answer`, pool )
				chIntro <- true
			} else if er.Message.Payload == "next" {
				chIntro <- true
			}
		})
		introWindow.Show()
		introWindow.Focus()
//		introWindow.OpenDevtools()
		
		go func() {
			http.HandleFunc("/", introLoader )
			http.Handle("/static/", http.FileServer(&assetfs.AssetFS{Asset: static.Asset, AssetDir: static.AssetDir, Prefix: ""}))
			http.ListenAndServe(":8990", nil)
		}()
		<- chIntro
		isClosed = true
		introWindow.Close()
	} else {
		mainWin = true
	}
	if len(pool.Pool) == 0 {
		resp, err := http.Get( GETPOOLURL + `?user_id=` + utils.Int64ToStr(idUser))
		if err!=nil {
			os.Exit(1)
		}
		jsonPool, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err!=nil {
			os.Exit(1)
		}
		json.Unmarshal(jsonPool, &pool)
	}
	if pool.Pool == `0` || len(pool.Pool) == 0 {
		pool.Pool = `http://pool.daylight.world`
	}
	
	fmt.Println( pool.Pool )

	thrustWindow = thrust.NewWindow(thrust.WindowOptions{
		Title : "DayLight Lite",
		RootUrl: pool.Pool,
		Size:    commands.SizeHW{Width: 1024, Height: 600},
	})
/*	thrustWindow.HandleEvent("*", func(cr commands.EventResult) {
		fmt.Println("HandleEvent", cr)
	})*/
	thrustWindow.HandleRemote(func(er commands.EventResult, this *window.Window) {
		fmt.Println("RemoteMessage Recieved:", er.Message.Payload)
		if len(er.Message.Payload) > 7 && er.Message.Payload[:7]==`USERID=` {
			err := ioutil.WriteFile( userfile, []byte(er.Message.Payload[7:]), 0644 )
			if err != nil {
				fmt.Println( `Error`, err )
			}
		} else {
			utils.ShellExecute(er.Message.Payload)
		}
	})
	thrustWindow.Show()
	thrustWindow.Focus()
//	thrustWindow.OpenDevtools()
	for {
		utils.Sleep(3600)
	}
	system.FinishThrust(0)
}
