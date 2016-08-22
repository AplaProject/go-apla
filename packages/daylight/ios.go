// +build darwin
// +build arm arm64

package daylight

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#import <Foundation/Foundation.h>

void
logNS(char* text) {
    NSLog(@"golog: %s", text);
}

*/
import "C"

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/stoppableListener"
	"net"
	"net/http"
)

var stop = make(chan bool)

func IosLog(text string) {
	if utils.IOS() {
		C.logNS(C.CString(text))
	}
}

func KillPid(pid string) error {
	return nil
}

func StartHTTPServer(ListenHttpHost string) {
	originalListener, err := net.Listen("tcp", ListenHttpHost)
	if err != nil {
		panic(err)
	}
	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		panic(err)
	}
	server := http.Server{}
	go func() {
		server.Serve(sl)
	}()
	<-stop
	sl.Stop()
}

func StopHTTPServer() {
	log.Debug("StopHTTPServer()")
	IosLog("StopHTTPServer 0")
	go func() { stop <- true }()
	utils.Sleep(1)
	IosLog("StopHTTPServer 1")
}

func tray() {

}

func httpListener(ListenHttpHost, BrowserHttpHost string) {
	go StartHTTPServer(ListenHttpHost)
}

func tcpListener() {

}

func signals(chans []*utils.DaemonsChansType) {

}
