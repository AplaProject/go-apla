package trayhost_test

import (
	"fmt"
	"github.com/cratonica/trayhost"
	"runtime"
)

// Refer to documentation at http://github.com/cratonica/trayhost for generating this
var iconData []byte

func main() {
	// EnterLoop must be called on the OS's main thread
	runtime.LockOSThread()

	go func() {
		// Run your application/server code in here. Most likely you will
		// want to start an HTTP server that the user can hit with a browser
		// by clicking the tray icon.

		// Be sure to call this to link the tray icon to the target url
		trayhost.SetUrl("http://github.com/cratonica/trayhost")
	}()

	// Enter the host system's event loop
	trayhost.EnterLoop("My Go App", iconData)

	// This is only reached once the user chooses the Exit menu item
	fmt.Println("Exiting")
}
