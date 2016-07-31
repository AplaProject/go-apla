package menu

/*
Package menu:
Provides common objects and methods for working with menus in
Thrust. Unfortunately, we have had to use GoRoutines in this and other packages to ensure that information is passed to GoLang in the proper order.
Please be aware of this when using this package in your own library.
The reason this is unfortunate is that as library writers it is best to leave GoRoutines out of your library and let the user decide when to use them. However with I/O you typically need to use GoRoutines in some way or another.
*/
import (
	"time"

	"github.com/go-thrust/lib/bindings/window"
	. "github.com/go-thrust/lib/commands"
	. "github.com/go-thrust/lib/common"
	"github.com/go-thrust/lib/connection"
	"github.com/go-thrust/lib/dispatcher"
)

/*
Menu is the basic object for creating and working with Menu's
Provides all the necessary attributes and methods to work with asynchronous calls to the menu API.
The TargetID is assigned by ThrustCore, so on init of this object, there is no TargetID. A Goroutine is dispatched to get the targetID.
*/
type Menu struct {
	TargetID         uint                                                 `json:"target_id,omitempty"`
	WaitingResponses []*Command                                           `json:"awaiting_responses,omitempty"`
	CommandQueue     []*Command                                           `json:"command_queue,omitempty"`
	Ready            bool                                                 `json:"ready"`
	Displayed        bool                                                 `json:"displayed"`
	Parent           *Menu                                                `json:"-"`
	Children         []*Menu                                              `json:"-"`
	Items            []*MenuItem                                          `json:"items,omitempty"`
	EventRegistry    []uint                                               `json:"events,omitempty"`
	SendChannel      *connection.In                                       `json:"-"`
	Sync             MenuSync                                             `jons:"-"`
	ReplyHandlers    map[uint]func(reply CommandResponse, item *MenuItem) `json:"_"`
}

/*
Create a new menu object.
Dispatches a call to ThrustCore to generate the object and return the new
TargetID in a reply.
*/
func NewMenu() *Menu {
	menu := Menu{}
	menuCreate := Command{
		Action:     "create",
		ObjectType: "menu",
	}
	menu.Sync = MenuSync{
		ReadyChan:        make(chan bool),
		DisplayedChan:    make(chan bool),
		ChildStableChan:  make(chan uint),
		TreeStableChan:   make(chan bool),
		ReadyQueue:       make([]*Command, 0),
		DisplayedQueue:   make([]*Command, 0),
		ChildStableQueue: make([]*ChildCommand, 0),
		TreeStableQueue:  make([]*Command, 0),
	}
	menu.ReplyHandlers = make(map[uint]func(reply CommandResponse, item *MenuItem))
	menu.SetSendChannel(connection.GetInputChannels())
	menu.WaitingResponses = append(menu.WaitingResponses, &menuCreate)
	dispatcher.RegisterHandler(menu.DispatchResponse)

	go menu.SendThread()
	menu.Send(&menuCreate)

	return &menu
}

/*
SetSendChannel is a helper Setter for SendChannel, in case we make it private in the future.
Use this for full forwards compatibility.
*/
func (menu *Menu) SetSendChannel(sendChannel *connection.In) {
	menu.SendChannel = sendChannel
}

/*
IsTarget checks if the current menu is the menu we are looking for.
*/
func (menu *Menu) IsTarget(targetId uint) bool {
	return targetId == menu.TargetID
}

/*
HandleError is a handler for Error responses from ThrustCore
This should be changed to private as soon as API stabilizes.
*/
func (menu *Menu) HandleError(reply CommandResponse) {

}

/*
HandleEvent is a handler for Event responses from ThrustCore
This should be changed to private as soon as API stabilizes.
*/
func (menu *Menu) HandleEvent(reply CommandResponse) {
	for _, item := range menu.Items {
		if reply.Event.CommandID == item.CommandID {
			Log.Print("Menu(", menu.TargetID, "):: Handling Event", item.CommandID, "::Handled With Flags", reply.Event.EventFlags, "With Type", item.Type)
			handler, ok := menu.ReplyHandlers[item.CommandID]
			if ok {
				handler(reply, item)
			} else {
				item.HandleEvent()
			}
			return
		}
	}
}

/*
HandleReply is a handler for Reply responses from ThrustCore
This should be changed to private as soon as API stabilizes.
*/
func (menu *Menu) HandleReply(reply CommandResponse) {

	for k, v := range menu.WaitingResponses {
		if v.ID != reply.ID {
			continue
		}
		Log.Print("MENU(", menu.TargetID, ")::Handling Reply", reply)
		removeAt := func(k int) {
			if len(menu.WaitingResponses) > 1 {
				menu.WaitingResponses = menu.WaitingResponses[:k+copy(menu.WaitingResponses[k:], menu.WaitingResponses[k+1:])]
			} else {
				menu.WaitingResponses = []*Command{}
			}
		}
		defer removeAt(k)

		if menu.TargetID == 0 && v.Action == "create" {
			//Assume we have a reply to action:create
			if reply.Result.TargetID != 0 {
				menu.TargetID = reply.Result.TargetID
				Log.Print("Received TargetID", "\nSetting Ready State")
				menu.Ready = true
			}
			for i, _ := range menu.CommandQueue {
				menu.CommandQueue[i].TargetID = menu.TargetID
				menu.Send(menu.CommandQueue[i])
			}
			// Reinitialize empty command queue, and allow gc.
			menu.CommandQueue = []*Command{}
			return
		}

		if v.Action == "call" && v.Method == "set_application_menu" {
			Log.Print("Received reply to set_application_menu", "Setting Menu Displayed to True")
			menu.setDisplayed(true)
		}

	}
}

/*
setDisplayed sets the menu Displayed attribute, this is a tracking attribute and has no effect on ThrustCore or UI Layer. Thus is is private.
*/
func (menu *Menu) setDisplayed(displayed bool) {
	menu.Displayed = displayed

	for _, child := range menu.Items {
		if child.IsSubMenu() {
			child.SubMenu.setDisplayed(displayed)
		}
	}
}

/*
DispatchResponse dispatches CommandResponses to the proper delegates (Error, Event, Reply)
*/
func (menu *Menu) DispatchResponse(reply CommandResponse) {
	switch reply.Action {
	case "event":
		menu.HandleEvent(reply)
	case "reply":
		menu.HandleReply(reply)
	}

	// for _, child := range menu.Items {
	// 	if child.IsSubMenu() {
	// 		child.SubMenu.DispatchResponse(reply)
	// 	}
	// }
}

/*
SendThread is a Thread for Sending Commands based on current state of the Menu.
Some commands require other events in the system to have already taken place.
This thread ensures that you can run almost any command at anytime, and have it take place in the correct order. This further insures that the underlying ThrustCore api does not crash, do to improper api knowledge.
*/
func (menu *Menu) SendThread() {
	//removeItemAt for []ChildCommand
	CCremoveItemAt := func(a []*ChildCommand, i int) []*ChildCommand {
		copy(a[i:], a[i+1:])
		a[len(a)-1] = nil // or the zero value of T
		a = a[:len(a)-1]
		return a
	}
	//removeItemAt for []*Command
	CremoveItemAt := func(a []*Command, i int) []*Command {
		copy(a[i:], a[i+1:])
		a[len(a)-1] = nil // or the zero value of T
		a = a[:len(a)-1]
		return a
	}
	go func() {
		for {
			if menu.Ready {
				menu.Sync.ReadyChan <- true
			}
			if menu.Displayed {
				menu.Sync.DisplayedChan <- true
			}
			for _, child := range menu.Items {
				if child.IsSubMenu() {
					if child.SubMenu.IsStable() {
						menu.Sync.ChildStableChan <- child.SubMenu.TargetID
					}
				}
			}
			if menu.IsTreeStable() {
				menu.Sync.TreeStableChan <- true
			}
			time.Sleep(time.Microsecond * 100)
		}
	}()

	go func() {
		for {
			select {
			case ready := <-menu.Sync.ReadyChan:
				if len(menu.Sync.ReadyQueue) == 0 || ready == false {
					break
				}
				command := menu.Sync.ReadyQueue[0]
				menu.Sync.ReadyQueue = CremoveItemAt(menu.Sync.ReadyQueue, 0)
				menu.Call(command)
			case displayed := <-menu.Sync.DisplayedChan:
				if len(menu.Sync.DisplayedQueue) == 0 || displayed == false {
					break
				}
				command := menu.Sync.DisplayedQueue[0]
				menu.Sync.DisplayedQueue = CremoveItemAt(menu.Sync.DisplayedQueue, 0)
				menu.WaitingResponses = append(menu.WaitingResponses, command)
				menu.Call(command)
			case stableChildId := <-menu.Sync.ChildStableChan:
				if len(menu.Sync.ChildStableQueue) == 0 {
					break
				}
				for i, childCommand := range menu.Sync.ChildStableQueue {
					if childCommand.Child.TargetID != stableChildId {
						continue
					}

					childCommand.Command.Args.MenuID = childCommand.Child.TargetID
					menu.Sync.ChildStableQueue = CCremoveItemAt(menu.Sync.ChildStableQueue, i)
					menu.Call(childCommand.Command)
					break

				}

			case treeStable := <-menu.Sync.TreeStableChan:
				if len(menu.Sync.TreeStableQueue) == 0 || treeStable == false {
					break
				}
				command := menu.Sync.TreeStableQueue[0]
				command.Args.MenuID = menu.TargetID
				menu.WaitingResponses = append(menu.WaitingResponses, command)
				menu.Sync.TreeStableQueue = CremoveItemAt(menu.Sync.TreeStableQueue, 0)
				menu.Call(command)
			}
		}
	}()
}

/*
Send emits a Command over the Command SendChannel to be delivered to ThrustCore
*/
func (menu *Menu) Send(command *Command) {
	menu.SendChannel.Commands <- command
}

/*
Call turns a Command into an action:call, there are two main types of Actions for outgoing commands, create/call. There may be more added later.
*/
func (menu *Menu) Call(command *Command) {
	command.Action = "call"
	command.TargetID = menu.TargetID

	if menu.Ready == false {
		menu.CommandQueue = append(menu.CommandQueue, command)
		return
	}
	menu.Send(command)
}

/*
CallWhenReady queues up "Calls" to go out only when the Menu State is "Ready"
*/
func (menu *Menu) CallWhenReady(command *Command) {
	menu.WaitingResponses = append(menu.WaitingResponses, command)
	menu.Sync.ReadyQueue = append(menu.Sync.ReadyQueue, command)
}

/*
CallWhenChildStable queues up "Calls" to go out only when the state of the Child is Stable. Stable means that the child is Ready and has no AwaitingResponses
*/
func (menu *Menu) CallWhenChildStable(command *Command, child *Menu) {
	menu.WaitingResponses = append(menu.WaitingResponses, command)
	menu.Sync.ChildStableQueue = append(menu.Sync.ChildStableQueue, &ChildCommand{
		Command: command,
		Child:   child,
	})
}

/*
CallWhenTreeStable queues up "Calls" to go out only when the state of the menu is Stable. Stable means that the menu is Ready and has no AwaitingResponses
*/
func (menu *Menu) CallWhenTreeStable(command *Command) {
	menu.Sync.TreeStableQueue = append(menu.Sync.TreeStableQueue, command)
}

/*
CallWhenDisplayed queues up "Calls" to go out only when
the menu is Displayed
*/
func (menu *Menu) CallWhenDisplayed(command *Command) {
	menu.Sync.DisplayedQueue = append(menu.Sync.DisplayedQueue, command)
}

/*
AddItem adds a MenuItem to both the internal representation of menu and the external representation of menu
*/
func (menu *Menu) AddItem(commandID uint, label string) {
	command := Command{
		Method: "add_item",
		Args: CommandArguments{
			CommandID: commandID,
			Label:     label,
		},
	}
	menuItem := MenuItem{
		CommandID: commandID,
		Label:     label,
		Parent:    menu,
		Type:      "item",
	}
	menu.Items = append(menu.Items, &menuItem)

	menu.CallWhenReady(&command)
}

/*
AddCheckItem adds a CheckItem to both the internal representation of menu and the external representation of menu
*/
func (menu *Menu) AddCheckItem(commandID uint, label string) {
	command := Command{
		Method: "add_check_item",
		Args: CommandArguments{
			CommandID: commandID,
			Label:     label,
		},
	}
	menuItem := MenuItem{
		CommandID: commandID,
		Label:     label,
		Type:      "check",
		Parent:    menu,
	}
	menu.Items = append(menu.Items, &menuItem)
	menu.CallWhenReady(&command)
}

/*
AddRadioItem adds a RadioItem to both the internal representation of menu and the external representation of menu
*/
func (menu *Menu) AddRadioItem(commandID uint, label string, groupID uint) {
	command := Command{
		Method: "add_radio_item",
		Args: CommandArguments{
			CommandID: commandID,
			Label:     label,
			GroupID:   groupID,
		},
	}
	menuItem := MenuItem{
		CommandID: commandID,
		Label:     label,
		GroupID:   groupID,
		Parent:    menu,
		Type:      "radio",
	}
	menu.Items = append(menu.Items, &menuItem)
	menu.CallWhenReady(&command)
}

/*
AddSubmenu adds a SubMenu to both the internal representation of menu and the external representation of menu
*/
func (menu *Menu) AddSubmenu(commandID uint, label string, child *Menu) {
	command := Command{
		Method: "add_submenu",
		Args: CommandArguments{
			CommandID: commandID,
			Label:     label,
		},
	}

	// Assign Bidirectional navigation elements i.e. DoublyLinkedLists
	child.Parent = menu
	menuItem := MenuItem{
		CommandID: commandID,
		Label:     label,
		SubMenu:   child,
		Parent:    menu,
	}
	menu.Items = append(menu.Items, &menuItem)

	menu.CallWhenChildStable(&command, child)
}

/*
SetChecked Checks or Unchecks a CheckItem in the UI
*/
func (menu *Menu) SetChecked(commandID uint, checked bool) {
	command := Command{
		Method: "set_checked",
		Args: CommandArguments{
			CommandID: commandID,
			Value:     checked,
		},
	}

	for _, item := range menu.Items {
		if item.IsCommandID(commandID) {
			item.Checked = checked
		}
	}
	menu.CallWhenDisplayed(&command)
}

/*
ToggleRadio Checks or Unchecks a RadioItem in the UI.
It is used by the default event handler to turn on the expected item,
and turn of other items in the group.
*/
func (menu *Menu) ToggleRadio(commandID, groupID uint, checked bool) {
	for _, item := range menu.RadioGroupAtGroupID(groupID) {
		command := Command{
			Method: "set_checked",
			Args: CommandArguments{
				CommandID: item.CommandID,
				Value:     checked,
			},
		}
		if item.IsCommandID(commandID) {
			item.Checked = checked
		} else {
			item.Checked = false
			command.Args.Value = false
		}
		menu.CallWhenDisplayed(&command)
	}

}

/*
SetEnabled sets whether or not a given item can receive
actions via the UI.
*/
func (menu *Menu) SetEnabled(commandID uint, enabled bool) {
	command := Command{
		Method: "set_enabled",
		Args: CommandArguments{
			CommandID: commandID,
			Value:     enabled,
		},
	}

	for _, item := range menu.Items {
		if item.IsCommandID(commandID) {
			item.Enabled = enabled
		}
	}
	menu.CallWhenDisplayed(&command)
}

/*
SetVisible sets a boolean visibility attribute in the UI for
a menu item with the given commandID.
*/
func (menu *Menu) SetVisible(commandID uint, visible bool) {
	command := Command{
		Method: "set_visible",
		Args: CommandArguments{
			CommandID: commandID,
			Value:     visible,
		},
	}

	for _, item := range menu.Items {
		if item.IsCommandID(commandID) {
			item.Visible = visible
		}
	}
	menu.CallWhenDisplayed(&command)
}

/*
AddSeperator adds a Seperator Item to both the internal representation of menu and the external representation of menu.
*/
func (menu *Menu) AddSeparator() {
	command := Command{
		Method: "add_separator",
	}
	menuItem := MenuItem{
		Type:   "separator",
		Parent: menu,
	}
	menu.Items = append(menu.Items, &menuItem)
	menu.CallWhenReady(&command)
}

/*
SetApplicationMenu sets the Application Menu on system that support global application level menus such as x11, unity, darwin
*/
func (menu *Menu) SetApplicationMenu() {
	command := Command{
		Method: "set_application_menu",
		Args: CommandArguments{
			MenuID: menu.TargetID,
		},
	}

	// Thread to wait for Stable Menu State
	menu.CallWhenTreeStable(&command)
}

/*
Popup creates a popup menu on the given window
*/
func (menu *Menu) Popup(w *window.Window) {
	go func() {
		for {
			if w.TargetID != 0 {
				command := Command{
					Method: "popup",
					Args: CommandArguments{
						WindowID: w.TargetID,
					},
				}

				// Thread to wait for Stable Menu State
				menu.CallWhenTreeStable(&command)
				return
			}
			time.Sleep(time.Microsecond * 10)
		}

	}()
}

/*
IsStable returns the a boolean value indicating that the menu is Ready and has no WaitingResponses
*/
func (menu *Menu) IsStable() bool {
	return menu.Ready && len(menu.WaitingResponses) == 0
}

/*
A Menu Tree is considered stable if and only if its children nodes report that they are stable.
Function is recursive, so factor that in to performance
*/
func (menu *Menu) IsTreeStable() bool {
	if !menu.IsStable() {
		return false
	}
	for _, child := range menu.Items {
		if child.IsSubMenu() {
			if !child.SubMenu.IsTreeStable() {
				return false
			}
		}
	}

	return true
}

/*
ItemAtCommandID recursively searches the Menu Tree for an item with the commandID. Returns the first found match. A proper menu should not reuse
commandID's
*/
func (menu *Menu) ItemAtCommandID(commandID uint) *MenuItem {
	for _, item := range menu.Items {
		if item.IsCommandID(commandID) {
			return item
		}
		if item.IsSubMenu() {
			result := item.SubMenu.ItemAtCommandID(commandID)
			if result != nil {
				return result
			}
		}
	}
	return nil
}

func (menu *Menu) RegisterEventHandlerByCommandID(commandID uint, handler func(reply CommandResponse, item *MenuItem)) {
	menu.ReplyHandlers[commandID] = func(reply CommandResponse, item *MenuItem) {
		handler(reply, item)
	}
}

/*
Find all menu items that belong to group identified by groupID
Not recursive, as a group should be identified at the same level.
Since it is not recursive you can theoretically reuse a groupID but problems
could creep up elsewhere, so please use unique groupID for radio items
*/
func (menu *Menu) RadioGroupAtGroupID(groupID uint) []*MenuItem {
	group := []*MenuItem{}
	for _, item := range menu.Items {
		if item.IsGroupID(groupID) {
			group = append(group, item)
		}
	}

	return group
}

/*
DEBUG Functions
*/
func (menu Menu) PrintRecursiveWaitingResponses() {
	Log.Print("Scanning Menu(", menu.TargetID, ")")
	if len(menu.WaitingResponses) > 0 {
		for _, v := range menu.WaitingResponses {
			Log.Print("Waiting for", v.ID, v.Action, v.Method)
		}
	}

	for _, child := range menu.Items {
		if child.IsSubMenu() {
			child.SubMenu.PrintRecursiveWaitingResponses()
		}
	}
}
