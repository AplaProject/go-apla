// +build !android,!ios

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
	//"fmt"
	//"github.com/AplaProject/go-apla/packages/consts"
	"net"
	"net/http"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/tcpserver"
	"github.com/AplaProject/go-apla/packages/utils"
	//"os"
	//"regexp"
	//	"time"

	"fmt"

	"time"

	"github.com/AplaProject/go-apla/packages/consts"
)

// IosLog is reserved
func IosLog(text string) {
}

/*
func NewBoundListener(maxActive int, l net.Listener) net.Listener {
	return &boundListener{l, make(chan bool, maxActive)}
}

type boundListener struct {
	net.Listener
	active chan bool
}

type boundConn struct {
	net.Conn
	active chan bool
}

func (l *boundListener) Accept() (net.Conn, error) {
	l.active <- true
	c, err := l.Listener.Accept()
	if err != nil {
		<-l.active
		return nil, err
	}
	return &boundConn{c, l.active}, err
}

func (l *boundConn) Close() error {
	err := l.Conn.Close()
	<-l.active
	return err
}
*/
func httpListener(ListenHTTPHost string, BrowserHTTPHost *string, route http.Handler) {

	i := 0
	host := ListenHTTPHost
	var l net.Listener
	var err error
	for {
		i++
		if i > 7 {
			log.Error("Error listening %d", host)
			panic("Error listening ")
		}
		if i > 1 {
			host = ":7" + converter.IntToStr(i) + "79"
			*BrowserHTTPHost = "http://" + host
		}
		log.Debug("host", host)
		l, err = net.Listen("tcp4", host)
		log.Debug("l", l)
		if err == nil {
			// Если это повторный запуск и он не из консоли, то открываем окно браузера, т.к. скорее всего юзер тыкнул по иконке
			// If this is a restart and it is made not from the console, then open the browser window, because user most likely pressed the icon
			/*if *utils.Console == 0 {
				openBrowser(browser)
			}*/
			fmt.Println("BrowserHTTPHost", host)
			break
		} else {
			log.Error(utils.ErrInfo(err).Error())
		}
	}

	go func() {
		srv := &http.Server{Handler: route} //Handler: http.TimeoutHandler(http.DefaultServeMux, time.Duration(120*time.Second), "Your request has timed out")}
		//		srv.SetKeepAlivesEnabled(false)
		err = srv.Serve(l)
		//		err = http.Serve( NewBoundListener(100, l), http.TimeoutHandler(http.DefaultServeMux, time.Duration(600*time.Second), "Your request has timed out"))
		if err != nil {
			log.Error("Error listening:", err, ListenHTTPHost)
			panic(err)
			//os.Exit(1)
		}
	}()
}

// For ipv6 on the server
func httpListenerV6(route http.Handler) {
	i := 0
	port := *utils.ListenHTTPPort
	var l net.Listener
	var err error
	for {
		if i > 7 {
			log.Error("Error listening ipv6 %d", port)
			panic("Error listening ")
		}
		if i > 0 {
			port = "7" + converter.IntToStr(i) + "79"
		}
		i++
		l, err = net.Listen("tcp6", ":"+port)
		if err == nil {
			break
		} else {
			log.Error(utils.ErrInfo(err).Error())
		}
	}

	go func() {
		srv := &http.Server{Handler: route} //Handler: http.TimeoutHandler(http.DefaultServeMux, time.Duration(120*time.Second), "Your request has timed out")}
		//		srv.SetKeepAlivesEnabled(false)
		err = srv.Serve(l)
		//		err = http.Serve(NewBoundListener(100, l), http.TimeoutHandler(http.DefaultServeMux, time.Duration(600*time.Second), "Your request has timed out"))
		if err != nil {
			log.Error("Error listening:", err)
			panic(err)
			//os.Exit(1)
		}
	}()
}

func tcpListener() {
	log.Debug("tcp")
	go func() {
		log.Debug("*utils.tcpHost: %v", *utils.TCPHost+":"+consts.TCP_PORT)
		//if len(*utils.TCPHost) > 0 {
		// включаем листинг TCP-сервером и обработку входящих запросов
		// switch on the listing by TCP-server and the processing of incoming requests
		l, err := net.Listen("tcp4", *utils.TCPHost+":"+consts.TCP_PORT)
		if err != nil {
			log.Error("Error listening:", err)
			//panic(err)
		} else {
			//defer l.Close()
			go func() {
				for {
					conn, err := l.Accept()
					if err != nil {
						log.Error("Error accepting:", err)
						time.Sleep(time.Second)
					} else {
						go func(conn net.Conn) {
							tcpserver.HandleTCPRequest(conn)
							conn.Close()
						}(conn)
					}
				}
			}()
		}
		//}
	}()
}
