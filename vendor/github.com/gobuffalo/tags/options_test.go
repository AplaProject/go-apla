package tags_test

import (
	"testing"

	"github.com/gobuffalo/tags"
	"github.com/stretchr/testify/require"
)

func Test_Options_String(t *testing.T) {
	r := require.New(t)
	o := tags.Options{
		"value": "Mark",
		"id":    "foo-bar",
		"class": "foo bar baz",
	}
	s := o.String()
	r.Equal(`class="foo bar baz" id="foo-bar" value="Mark"`, s)
}

func Test_Options_String_Escaped(t *testing.T) {
	r := require.New(t)
	o := tags.Options{
		"<b>": "<p>",
	}
	s := o.String()
	r.Equal(`&lt;b&gt;="&lt;p&gt;"`, s)
}

func Test_Options_String_Empty_Attribute(t *testing.T) {
	r := require.New(t)
	o := tags.Options{
		"value":   "Mark",
		"checked": nil,
	}
	s := o.String()
	r.Equal(`checked value="Mark"`, s)
}

func Test_Options_Data_Map(t *testing.T) {
	r := require.New(t)
	o := tags.Options{
		"value": "Mark",
		"id":    "foo-bar",
		"class": "foo bar baz",
		"data": map[string]interface{}{
			"remote": true,
			"method": "PUT",
		},
	}
	s := o.String()
	r.Equal(`class="foo bar baz" data-method="PUT" data-remote="true" id="foo-bar" value="Mark"`, s)
}
