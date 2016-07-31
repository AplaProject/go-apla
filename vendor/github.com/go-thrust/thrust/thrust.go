package thrust

import (
	"runtime"

	"github.com/go-thrust/lib/bindings/menu"
	"github.com/go-thrust/lib/bindings/session"
	"github.com/go-thrust/lib/bindings/window"
	"github.com/go-thrust/lib/common"
	"github.com/go-thrust/lib/connection"
	"github.com/go-thrust/lib/dispatcher"
	"github.com/go-thrust/lib/events"
	"github.com/go-thrust/lib/spawn"
)

/*
Begin Generic Access and Binding Management Section.
*/

/*
Bindings
*/

type WindowOptions window.Options

/* NewWindow creates a new Window Binding */
func NewWindow(options WindowOptions) *window.Window {
	return window.NewWindow(window.Options(options))
}

/* NewSession creates a new Session Binding */
func NewSession(incognito, overrideDefaultSession bool, path string) *session.Session {
	return session.NewSession(incognito, overrideDefaultSession, path)
}

/* NewMenu creates a new Menu Binding */
func NewMenu() *menu.Menu {
	return menu.NewMenu()
}

/*
Start spawns the thrust core executable, and begins the dispatcher loop in a go routine
*/
func Start() {
	spawn.Run()
	go dispatcher.RunLoop()
}

/*
SetProvisioner overrides the default Provisioner, the default provisioner downloads
Thrust-Core if Thrust-Core is not found.
It also does some other nifty things to configure your install (on darwin) for the ApplicationName you choose.
*/
func SetProvisioner(p spawn.Provisioner) {
	spawn.SetProvisioner(p)
}

/*
Use LockThread on the main thread in lieue of a webserver or some other service that holds the thread
This is primarily used when Thrust and just Thrust is what you are using, in that case lock the thread.
Otherwise, why dont you start an http server, and expose some websockets.
*/
func LockThread() {
	for {
		runtime.Gosched()
	}
}

/*
Initialize and Enable the internal *log.Logger.
*/
func InitLogger() {
	common.InitLogger("")
}

/*
Disable the internal *log.Logger instance
*/
func DisableLogger() {
	common.InitLogger("none")
}

/*
ALWAYS use this method instead of os.Exit()
This method will handle destroying the child process, and exiting as cleanly as possible.
*/
func Exit() {
	connection.CleanExit()
}

/*
Sets the Application Name
*/
func SetApplicationName(name string) {
	spawn.ApplicationName = name
}

/*
Create a new EventHandler for a give event.
*/
func NewEventHandler(event string, fn interface{}) (events.ThrustEventHandler, error) {
	return events.NewHandler(event, fn)
}
