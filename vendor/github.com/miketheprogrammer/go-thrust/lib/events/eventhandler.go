package events

import (
	"errors"

	"github.com/miketheprogrammer/go-thrust/lib/commands"
	"github.com/miketheprogrammer/go-thrust/lib/dispatcher"
)

/*
Create a new EventHandler for a give event.
*/
func NewHandler(event string, fn interface{}) (ThrustEventHandler, error) {
	h := ThrustEventHandler{}
	h.Event = event
	h.Type = "event"
	err := h.SetHandleFunc(fn)
	dispatcher.RegisterHandler(h)
	return h, err
}

/**
Begin Thrust Handler Code.
**/
type Handler interface {
	Handle(cr commands.CommandResponse)
	Register()
	SetHandleFunc(fn interface{})
}

type ThrustEventHandler struct {
	Type    string
	Event   string
	Handler interface{}
}

func (teh ThrustEventHandler) Handle(cr commands.CommandResponse) {
	if cr.Action != "event" {
		return
	}
	if cr.Type != teh.Event && teh.Event != "*" {
		return
	}
	cr.Event.Type = cr.Type
	if fn, ok := teh.Handler.(func(commands.CommandResponse)); ok == true {
		fn(cr)
		return
	}
	if fn, ok := teh.Handler.(func(commands.EventResult)); ok == true {
		fn(cr.Event)
		return
	}
}

func (teh *ThrustEventHandler) SetHandleFunc(fn interface{}) error {
	if fn, ok := fn.(func(commands.CommandResponse)); ok == true {
		teh.Handler = fn
		return nil
	}
	if fn, ok := fn.(func(commands.EventResult)); ok == true {
		teh.Handler = fn
		return nil
	}

	return errors.New("Invalid Handler Definition")
}
