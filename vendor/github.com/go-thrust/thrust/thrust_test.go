package thrust

import (
	"testing"

	"github.com/go-thrust/lib/commands"
)

func TestNewEventHandler(t *testing.T) {
	handler, err := NewEventHandler("focus", func(cr commands.CommandResponse) {

	})

	if err != nil {
		t.Fail()
	}
	if handler.Type != "event" {
		t.Fail()
	}
	if handler.Event != "focus" {
		t.Fail()
	}

	handler, err = NewEventHandler("focus", func(rr commands.ReplyResult) {

	})

	if err == nil {
		t.Fail()
	}

	handler, err = NewEventHandler("focus", func(er commands.EventResult) {

	})

	if err != nil {
		t.Fail()
	}
	if handler.Type != "event" {
		t.Fail()
	}
	if handler.Event != "focus" {
		t.Fail()
	}
}

func TestHandlerHandle(t *testing.T) {
	var responses int

	handler, err := NewEventHandler("focus", func(cr commands.CommandResponse) {
		t.Log("Woot")
		t.Log(cr)
		responses += 1
	})

	if err != nil {
		t.Fail()
	}

	handler.Handle(commands.CommandResponse{
		Action: "event",
		ID:     1,
		Type:   "focus",
		Event: commands.EventResult{
			Type: "focus",
		},
	})

	handler, err = NewEventHandler("focus", func(er commands.EventResult) {
		t.Log("Woot")
		t.Log(er)
		responses += 1
	})

	if err != nil {
		t.Fail()
	}

	handler.Handle(commands.CommandResponse{
		Action: "event",
		ID:     1,
		Type:   "focus",
		Event: commands.EventResult{
			Type: "focus",
		},
	})

	if responses != 2 {
		t.Fail()
	}
}
