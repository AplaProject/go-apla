package session

/*
Pacage Session contains an API Binding, and
interfaces that will assist in accessing the
default Session or implementing your own Custom Session.
*/
import (
	"encoding/json"

	. "github.com/go-thrust/lib/commands"
	. "github.com/go-thrust/lib/common"
	"github.com/go-thrust/lib/connection"
	"github.com/go-thrust/lib/dispatcher"
)

/*
Session is the core API Binding object used to communicate with Thrust.
*/
type Session struct {
	TargetID                 uint
	CookieStore              bool
	OffTheRecord             bool
	Path                     string
	Ready                    bool
	CommandHistory           []*Command
	ResponseHistory          []*CommandResponse
	WaitingResponses         []*Command
	CommandQueue             []*Command
	SendChannel              *connection.In
	SessionOverrideInterface SessionInvokable
}

/*
NewSession is a constructor that takes 3 arguments,
incognito which is a boolean, meaning dont persist session state
after close.
overrideDefaultSession which is a boolean that till tell thrust core
to try to invoke session methods from us.
path string is the path to store session data.
*/
func NewSession(incognito, overrideDefaultSession bool, path string) *Session {
	session := Session{
		CookieStore:  overrideDefaultSession,
		OffTheRecord: incognito,
		Path:         path,
	}
	if overrideDefaultSession == true {
		session.SetInvokable(*NewDummySession())
	}
	command := Command{
		Action:     "create",
		ObjectType: "session",
		Args: CommandArguments{
			CookieStore:  session.CookieStore,
			OffTheRecord: session.OffTheRecord,
			Path:         session.Path,
		},
	}
	session.SendChannel = connection.GetInputChannels()
	session.WaitingResponses = append(session.WaitingResponses, &command)
	session.Send(&command)
	dispatcher.RegisterHandler(session.DispatchResponse)
	return &session
}

func (session *Session) HandleInvoke(reply CommandResponse) {
	if reply.TargetID == session.TargetID {
		response := &CommandResponse{
			Action: "reply",
			ID:     reply.ID,
			Result: ReplyResult{},
		}
		switch reply.Method {
		case "cookies_load":
			cookies := session.SessionOverrideInterface.InvokeCookiesLoad(&reply.Args, session)
			marshaledCookies, err := json.Marshal(cookies)
			if err != nil {
				Log.Print(err)
			} else {
				response.Result.Cookies = marshaledCookies
			}
		case "cookies_load_for_key":
			cookies := session.SessionOverrideInterface.InvokeCookiesLoadForKey(&reply.Args, session)
			marshaledCookies, err := json.Marshal(cookies)
			if err != nil {
				Log.Print(err)
			} else {
				response.Result.Cookies = marshaledCookies
			}
		case "cookies_flush":
			session.SessionOverrideInterface.InvokeCookiesFlush(&reply.Args, session)
		case "cookies_add":
			session.SessionOverrideInterface.InvokeCookiesAdd(&reply.Args, session)
		case "cookies_delete":
			session.SessionOverrideInterface.InvokeCookiesDelete(&reply.Args, session)
		case "cookies_update_access_time":
			session.SessionOverrideInterface.InvokeCookiesUpdateAccessTime(&reply.Args, session)
		case "cookies_force_keep_session_state":
			session.SessionOverrideInterface.InvokeCookieForceKeepSessionState(&reply.Args, session)
		}
		Log.Print("Sending Response to Invoke")
		session.SendChannel.CommandResponses <- response
	}
}

func (session *Session) HandleReply(reply CommandResponse) {
	Log.Print(reply)
	for k, command := range session.WaitingResponses {
		if command.ID != reply.ID {
			continue
		}
		if command.ID == reply.ID {
			Log.Print("Window(", session.TargetID, ")::Handling Reply::", reply)
			if len(session.WaitingResponses) > 1 {
				// Remove the element at index k
				session.WaitingResponses = session.WaitingResponses[:k+copy(session.WaitingResponses[k:], session.WaitingResponses[k+1:])]
			} else {
				// Just initialize to empty splice literal
				session.WaitingResponses = []*Command{}
			}
			Log.Print("session", session.TargetID, command.Action, reply.Result.TargetID)
			if session.TargetID == 0 && command.Action == "create" {
				if reply.Result.TargetID != 0 {
					session.TargetID = reply.Result.TargetID
					Log.Print("Session:: Received TargetID(", session.TargetID, ") :: Setting Ready State")
					session.Ready = true
				}
			}
		}
	}
}

func (session *Session) DispatchResponse(response CommandResponse) {
	switch response.Action {
	case "invoke":
		session.HandleInvoke(response)
	case "reply":
		session.HandleReply(response)

	}
}

func (session *Session) Send(command *Command) {
	session.SendChannel.Commands <- command
}

func (session *Session) SetInvokable(si SessionInvokable) {
	session.SessionOverrideInterface = si
}
