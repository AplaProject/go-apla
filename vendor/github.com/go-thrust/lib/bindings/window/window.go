package window

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/go-thrust/lib/bindings/session"
	. "github.com/go-thrust/lib/commands"
	. "github.com/go-thrust/lib/common"
	"github.com/go-thrust/lib/connection"
	"github.com/go-thrust/lib/dispatcher"
	"github.com/go-thrust/lib/events"
	"github.com/go-thrust/lib/spawn"
)

type Window struct {
	TargetID         uint
	CommandHistory   []*Command
	ResponseHistory  []*CommandResponse
	WaitingResponses []*Command
	CommandQueue     []*Command
	Url              string
	Title            string
	Ready            bool
	Displayed        bool
	SendChannel      *connection.In `json:"-"`
}

type Options struct {
	RootUrl  string
	Size     SizeHW
	Title    string
	IconPath string
	HasFrame bool
	Session  *session.Session
}

func checkUrl(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}
	if u.Scheme == "" {
		p, err := filepath.Abs(s)
		if err != nil {
			return s, err
		}
		u = &url.URL{
			Scheme: "file",
			Path:   p,
		}
	}
	return u.String(), err
}

func NewWindow(options Options) *Window {
	w := Window{}
	w.setOptions(options)
	_, sendChannel := connection.GetCommunicationChannels()

	size := options.Size
	if options.Size == (SizeHW{}) {
		size = SizeHW{
			Width:  1024,
			Height: 768,
		}
	}

	windowCreate := Command{
		Action:     "create",
		ObjectType: "window",
		Args: CommandArguments{
			RootUrl:  w.Url,
			Title:    spawn.ApplicationName,
			Size:     size,
			HasFrame: !options.HasFrame,
		},
	}
	dispatcher.RegisterHandler(w.DispatchResponse)
	if options.Session == nil {
		w.SetSendChannel(sendChannel)
		w.WaitingResponses = append(w.WaitingResponses, &windowCreate)
		w.Send(&windowCreate)
	} else {
		go func() {
			for {
				if options.Session.TargetID != 0 {
					fmt.Println("sess", options.Session.TargetID)
					windowCreate.Args.SessionID = options.Session.TargetID
					w.SetSendChannel(sendChannel)
					w.WaitingResponses = append(w.WaitingResponses, &windowCreate)
					w.Send(&windowCreate)
					return
				}
				time.Sleep(time.Microsecond * 10)
			}
		}()
	}
	return &w
}

func (w *Window) setOptions(options Options) {
	u, _ := checkUrl(options.RootUrl)
	w.Url = u
	if len(w.Url) == 0 {
		w.Url = "http://google.com"
	}

	w.Title = options.Title
	if len(w.Title) == 0 {
		w.Title = spawn.ApplicationName
	}
}

func (w *Window) SetSendChannel(sendChannel *connection.In) {
	w.SendChannel = sendChannel
}

func (w *Window) IsTarget(targetId uint) bool {
	return targetId == w.TargetID
}

func (w *Window) HandleError(reply CommandResponse) {

}

func (w *Window) HandleReply(reply CommandResponse) {
	for k, v := range w.WaitingResponses {
		if v.ID != reply.ID {
			continue
		}
		Log.Print("Window(", w.TargetID, ")::Handling Reply::", reply)
		if len(w.WaitingResponses) > 1 {
			// Remove the element at index k
			w.WaitingResponses = w.WaitingResponses[:k+copy(w.WaitingResponses[k:], w.WaitingResponses[k+1:])]
		} else {
			// Just initialize to empty splice literal
			w.WaitingResponses = []*Command{}
		}

		// If we dont already have a TargetID then we accept a create action
		if w.TargetID == 0 && v.Action == "create" {
			if reply.Result.TargetID != 0 {
				w.TargetID = reply.Result.TargetID
				Log.Print("Received TargetID", "\nSetting Ready State")
				w.Ready = true
			}

			for i, _ := range w.CommandQueue {
				w.CommandQueue[i].TargetID = w.TargetID
				w.Send(w.CommandQueue[i])
			}
			// Reinitialize empty command queue, and allow gc.
			w.CommandQueue = []*Command{}

			return
		}

		if v.Action == "call" && v.Method == "show" {
			w.Displayed = true
		}

	}
}

func (w *Window) DispatchResponse(reply CommandResponse) {

	switch reply.Action {
	case "reply":
		w.HandleReply(reply)
	}

}
func (w *Window) Send(command *Command) {

	w.SendChannel.Commands <- command
}

func (w *Window) Call(command *Command) {
	command.Action = "call"
	command.TargetID = w.TargetID
	if w.Ready == false {
		w.CommandQueue = append(w.CommandQueue, command)
		return
	}
	w.Send(command)
}

func (w *Window) CallWhenReady(command *Command) {
	w.WaitingResponses = append(w.WaitingResponses, command)
	go func() {
		for {
			if w.Ready {
				w.Call(command)
				return
			}
			time.Sleep(time.Microsecond * 100)
		}
	}()
}

func (w *Window) CallWhenDisplayed(command *Command) {
	w.WaitingResponses = append(w.WaitingResponses, command)
	go func() {
		for {
			if w.Displayed {
				w.Call(command)
				return
			}
			time.Sleep(time.Microsecond * 100)
		}
	}()
}

func (w *Window) Show() {
	command := Command{
		Method: "show",
	}

	w.CallWhenReady(&command)
}

func (w *Window) SetTitle(title string) {
	command := Command{
		Method: "set_title",
		Args: CommandArguments{
			Title: title,
		},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) Maximize() {
	command := Command{
		Method: "maximize",
	}
	w.CallWhenDisplayed(&command)
}

func (w *Window) UnMaximize() {
	command := Command{
		Method: "unmaximize",
	}
	w.CallWhenDisplayed(&command)
}

func (w *Window) Minimize() {
	command := Command{
		Method: "minmize",
	}
	w.CallWhenDisplayed(&command)
}

func (w *Window) Restore() {
	command := Command{
		Method: "restore",
	}
	w.CallWhenDisplayed(&command)
}

func (w *Window) Focus() {
	command := Command{
		Method: "focus",
		Args: CommandArguments{
			Focus: true,
		},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) UnFocus() {
	command := Command{
		Method: "show",
		Args: CommandArguments{
			Focus: false,
		},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) Fullscreen(fullscreen bool) {
	command := Command{
		Method: "set_fullscreen",
		Args: CommandArguments{
			Fullscreen: fullscreen,
		},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) Kiosk(kiosk bool) {
	command := Command{
		Method: "set_kiosk",
		Args: CommandArguments{
			Kiosk: kiosk,
		},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) Close() {
	command := Command{
		Method: "close",
		Args:   CommandArguments{},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) OpenDevtools() {
	command := Command{
		Method: "open_devtools",
		Args:   CommandArguments{},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) CloseDevtools() {
	command := Command{
		Method: "close_devtools",
		Args:   CommandArguments{},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) Move(x, y int) {
	command := Command{
		Method: "move",
		Args: CommandArguments{
			X: x,
			Y: y,
		},
	}

	w.CallWhenDisplayed(&command)
}

func (w *Window) Resize(width, height uint) {
	command := Command{
		Method: "resize",
		Args: CommandArguments{
			Width:  width,
			Height: height,
		},
	}

	w.CallWhenReady(&command)
}

func (w *Window) Position(x, y int) {
	command := Command{
		Method: "position",
		Args: CommandArguments{
			Position: PositionXY{
				X: x,
				Y: y,
			},
		},
	}

	w.CallWhenReady(&command)
}

func (w *Window) SendRemoteMessage(msg string) {
	command := Command{
		Method: "remote",
		Args: CommandArguments{
			Message: RemoteMessage{
				Payload: msg,
			},
		},
	}

	// We dont use call, because messages are of variable size, and we definitely,
	// do not want to store a reference to them. So we use .Send
	go func() {
		for {
			if w.Displayed {
				command.Action = "call"
				command.TargetID = w.TargetID
				w.Send(&command)
				return
			}
			time.Sleep(time.Microsecond * 100)
		}
	}()
}

/*
CRorER means commands.CommandResponse or commands.EventResult
*/
type WindowEventHandler func(CRorER interface{}, window *Window)

/*
Binding Event Handlers are a bit different than global thrust handlers.
The Signature of the function you pass in is WindowEventHandler
*/
func (w *Window) HandleEvent(event string, fn interface{}) (events.ThrustEventHandler, error) {
	if fn, ok := fn.(func(CommandResponse, *Window)); ok == true {
		return events.NewHandler(event, func(cr CommandResponse) {
			fn(cr, w)
		})
	}
	if fn, ok := fn.(func(EventResult, *Window)); ok == true {
		return events.NewHandler(event, func(er EventResult) {
			fn(er, w)
		})
	}
	return events.ThrustEventHandler{}, errors.New("Function Signature Invalid")
}

func (w *Window) HandleBlur(fn interface{}) (events.ThrustEventHandler, error) {
	return w.HandleEvent("blur", fn)
}

func (w *Window) HandleRemote(fn interface{}) (events.ThrustEventHandler, error) {
	return w.HandleEvent("remote", fn)

}
