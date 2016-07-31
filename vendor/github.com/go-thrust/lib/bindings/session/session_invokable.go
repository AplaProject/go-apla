package session

import "github.com/go-thrust/lib/commands"

/*
Methods prefixed with Invoke are methods that can be called by ThrustCore, this differs to our
standard call/reply, or event actions, since we are now the responder.
*/
/*
SessionInvokable is an interface designed to allow you to create your own Session Store.
Simple build a structure that supports these methods, and call session.SetInvokable(myInvokable)

?? What is the best and most simple value to return from these methods. Assume most of them work
on the principle of single value return. Do we worry about different types or do we use a CommandResponse object
If we use different types, we will just have to add them to a CommandResponse from the Caller anyway.
I would say different types will keep the user from shooting himself in the foot.
*/

type SessionInvokable interface {
	InvokeCookiesLoad(args *commands.CommandResponseArguments, session *Session) (cookies []Cookie)
	InvokeCookiesLoadForKey(args *commands.CommandResponseArguments, session *Session) (cookies []Cookie)
	InvokeCookiesFlush(args *commands.CommandResponseArguments, session *Session) bool
	InvokeCookiesAdd(args *commands.CommandResponseArguments, session *Session) bool
	InvokeCookiesUpdateAccessTime(args *commands.CommandResponseArguments, session *Session) bool
	InvokeCookiesDelete(args *commands.CommandResponseArguments, session *Session) bool
	InvokeCookieForceKeepSessionState(args *commands.CommandResponseArguments, session *Session)
}
