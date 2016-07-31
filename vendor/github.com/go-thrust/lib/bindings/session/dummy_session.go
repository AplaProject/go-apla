package session

import (
	"github.com/go-thrust/lib/commands"
	"github.com/go-thrust/lib/common"
)

type DummySession struct{}

func NewDummySession() (dummy *DummySession) {
	return &DummySession{}
}

/*
For Simplicity type declarations
*/
func (ds DummySession) InvokeCookiesLoad(args *commands.CommandResponseArguments, session *Session) (cookies []Cookie) {
	common.Log.Print("InvokeCookiesLoad")
	cookies = make([]Cookie, 0)

	return cookies
}

func (ds DummySession) InvokeCookiesLoadForKey(args *commands.CommandResponseArguments, session *Session) (cookies []Cookie) {
	common.Log.Print("InvokeCookiesLoadForKey")
	cookies = make([]Cookie, 0)

	return cookies
}

func (ds DummySession) InvokeCookiesFlush(args *commands.CommandResponseArguments, session *Session) bool {
	common.Log.Print("InvokeCookiesFlush")
	return false
}

func (ds DummySession) InvokeCookiesAdd(args *commands.CommandResponseArguments, session *Session) bool {
	common.Log.Print("InvokeCookiesAdd")
	return false
}

func (ds DummySession) InvokeCookiesUpdateAccessTime(args *commands.CommandResponseArguments, session *Session) bool {
	common.Log.Print("InvokeCookiesUpdateAccessTime")
	return false
}

func (ds DummySession) InvokeCookiesDelete(args *commands.CommandResponseArguments, session *Session) bool {
	common.Log.Print("InvokeCookiesDelete")
	return false
}

func (ds DummySession) InvokeCookieForceKeepSessionState(args *commands.CommandResponseArguments, session *Session) {
	common.Log.Print("InvokeCookieForceKeepSessionState")
}
