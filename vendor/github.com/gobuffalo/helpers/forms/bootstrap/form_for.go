package bootstrap

import (
	"html/template"

	"github.com/gobuffalo/helpers/hctx"
	"github.com/gobuffalo/helpers/paths"
	"github.com/gobuffalo/tags"
	"github.com/gobuffalo/tags/form/bootstrap"
)

// FormFor implements a `github.com/gobuffalo/plush` helper around the
// bootstrap.NewFormFor function in the `github.com/gobuffalo/tags/form/bootstrap` package
func FormFor(model interface{}, opts tags.Options, help hctx.HelperContext) (template.HTML, error) {
	if opts == nil {
		opts = tags.Options{}
	}
	if _, ok := opts["action"]; !ok {
		act, err := paths.PathFor(model)
		if err != nil {
			return "", err
		}
		opts["action"] = act
	}
	return helper(opts, help, func(opts tags.Options) helperable {
		return bootstrap.NewFormFor(model, opts)
	})
}

// FormFor implements a `github.com/gobuffalo/plush` helper around the
// bootstrap.NewFormFor function in the `github.com/gobuffalo/tags/form/bootstrap` package
//	<%= remoteFormFor(widget, {method: "POST"}) { %>
//	<% } %>
//	<form action="/widgets/b6b0ab24-19ae-4cdd-ad73-c5ecbddd6f91" id="widget-form" method="POST"><input name="_method" type="hidden" value="PUT"></form>
func RemoteFormFor(model interface{}, opts tags.Options, help hctx.HelperContext) (template.HTML, error) {
	if opts == nil {
		opts = tags.Options{}
	}
	opts["data-remote"] = true
	return FormFor(model, opts, help)
}
