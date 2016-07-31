package commands

import "encoding/json"

/*
commands package contains structures for working with JSON RPC
Calls to ThrustCore
*/

/*
Command defines the structure used in send json rpc messages to ThrustCore
*/
type Command struct {
	ID         uint             `json:"_id"`
	Action     string           `json:"_action"`
	ObjectType string           `json:"_type,omitempty"`
	Method     string           `json:"_method"`
	TargetID   uint             `json:"_target,omitempty"`
	Args       CommandArguments `json:"_args"`
}

/*
CommandArguments defines the structure used in providing arguments
to Command's when talking to ThrustCore
Covers all possible argument combinations.
Makes use of omit empty to adapt to different use cases
*/
type CommandArguments struct {
	RootUrl      string        `json:"root_url,omitempty"`
	Title        string        `json:"title,omitempty"`
	Size         SizeHW        `json:"size,omitempty"`
	Position     PositionXY    `json:"position,omitempty"`
	X            int           `json:"x,omitempty"`
	Y            int           `json:"y,omitempty"`
	Width        uint          `json:"width,omitempty"`
	Height       uint          `json:"height,omitempty"`
	CommandID    uint          `json:"command_id,omitempty"`
	Label        string        `json:"label,omitempty"`
	MenuID       uint          `json:"menu_id,omitempty"` // this should never be 0 anyway
	WindowID     uint          `json:"window_id,omitempty"`
	SessionID    uint          `json:"session_id,omitempty"`
	GroupID      uint          `json:"group_id,omitempty"`
	Value        bool          `json:"value"`
	CookieStore  bool          `json:"cookie_store"`
	OffTheRecord bool          `json:"off_the_record"`
	Fullscreen   bool          `json:"fullscreen"`
	Kiosk        bool          `json:"kiosk"`
	Focus        bool          `json:"focus"`
	HasFrame     bool          `json:"has_frame"`
	Path         string        `json:"path,omitempty"`
	Message      RemoteMessage `json:"message,omitempty"`
}

/*
SizeHW is a simple struct defining Height and Width
*/
type SizeHW struct {
	Width  uint `json:"width,omitempty"`
	Height uint `json:"height,omitempty"`
}

type PositionXY struct {
	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}

/*
ReplyResult is used in CommandResponse's of Type Reply
*/
type ReplyResult struct {
	TargetID uint            `json:"_target,omitempty"`
	Cookies  json.RawMessage `json:"cookies"`
}

/*
EventResult is used in CommandResponse's of Type Event
*/
type EventResult struct {
	CommandID  uint          `json:"command_id,omitempty"`
	EventFlags int           `json:"event_flags,omitempty"`
	Message    RemoteMessage `json:"message,omitempty"`
	Type       string        `json:"-"`
}

/*
Remote Message
*/
type RemoteMessage struct {
	Payload string `json:"payload"`
}

/*
CommandResponse defines the structure of a response
from a Command sent to ThrustCore
*/
type CommandResponse struct {
	Action   string                   `json:"_action,omitempty"`
	Error    string                   `json:"_error,omitempty"`
	ID       uint                     `json:"_id,omitempty"`
	Result   ReplyResult              `json:"_result,omitempty"`
	Event    EventResult              `json:"_event,omitempty"`
	Args     CommandResponseArguments `json:"_args,omitempty"`
	TargetID uint                     `json:"_target,omitempty"`
	Method   string                   `json:"_method,omitempty"`
	Type     string                   `json:"_type,omitempty"`
}

/*
Command Response Arguments covers all possible cases of CommandResponses that contain an _args parameter.
This is often the case with methods on the Session Object that are Invoked from the core.
*/
type CommandResponseArguments struct {
	key string `json:"key,omitempty"`
}
