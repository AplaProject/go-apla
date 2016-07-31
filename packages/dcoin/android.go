// +build android

package dcoin

import (
	"net/http"
)

func IosLog(text string) {
}

func KillPid(pid string) error {
	return nil
}

func httpListener(ListenHttpHost, BrowserHttpHost string) {
	go func() {
		http.ListenAndServe(ListenHttpHost, nil)
	}()
}

func tcpListener() {

}

func tray() {

}
