package dispatcher

import (
	"runtime"

	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/lib/connection"
)

type HandleFunc func(commands.CommandResponse)
type Handler interface {
	Handle(commands.CommandResponse)
}

var registry []interface{}

/*
RegisterHandler registers a HandleFunc f to receive a CommandResponse when one is sent to the system.
*/
func RegisterHandler(h interface{}) {
	registry = append(registry, h)
}

/*
Dispatch dispatches a CommandResponse to every handler in the registry
*/
func Dispatch(command commands.CommandResponse) {
	for _, f := range registry {
		if fn, ok := f.(func(cr commands.CommandResponse)); ok == true {
			go fn(command)
		}
		if handler, ok := f.(Handler); ok == true {
			go handler.Handle(command)
		}
	}
}

/*
RunLoop starts a loop that receives CommandResponses and dispatches them.
This is a helper method, but you could just implement your own, if you only
need this loop to be the blocking loop.
For Instance, in a HTTP Server setting, you might want to run this as a
goroutine and then let the servers blocking handler keep the process open.
As long as there are commands in the channel, this loop will dispatch as fast
as possible
*/
func RunLoop() {
	outChannels := connection.GetOutputChannels()
	defer connection.Clean()

	for {
		Run(outChannels)
		runtime.Gosched()
	}

}

func Run(outChannels *connection.Out) {
	response := <-outChannels.CommandResponses
	Dispatch(response)
}
