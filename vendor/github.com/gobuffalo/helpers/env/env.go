package env

import (
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/helpers/hctx"
)

// Keys to be used in templates for the functions in this package.
const (
	EnvKey   = "env"
	EnvOrKey = "envOr"
)

// New returns a map of the helpers within this package.
func New() hctx.Map {
	return hctx.Map{
		EnvKey:   Env,
		EnvOrKey: EnvOr,
	}
}

// Env will return the specified environment variable,
// or an error if it can not be found
//	<%= env("GOPATH") %>
var Env = envy.MustGet

// Env will return the specified environment variable,
// or the second argument, if not found
//	<%= envOr("GOPATH", "~/go) %>
var EnvOr = envy.Get
