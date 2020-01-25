package tags

import (
	"errors"
	"html/template"
	"testing"

	"github.com/gobuffalo/helpers/helptest"
	"github.com/gobuffalo/tags"
	"github.com/stretchr/testify/require"
)

func Test_LinkTo(t *testing.T) {
	table := []struct {
		in       interface{}
		out      string
		opts     tags.Options
		body     string
		errBlock bool
		err      bool
	}{
		{"foo", `<a href="/foo"></a>`, tags.Options{}, "", false, false},
		{"foo", `<a class="btn" href="/foo"></a>`, tags.Options{"class": "btn"}, "", false, false},
		{[]string{"foo", "bar"}, `<a href="/foo/bar">baz</a>`, tags.Options{"body": "baz"}, "", false, false},
		{"foo", `<a href="/foo">my body</a>`, tags.Options{}, "my body", false, false},
		{"foo", `<a href="/foo"></a>`, nil, "", false, false},
		{"foo", ``, nil, "", true, true},
		{nil, ``, nil, "", false, true},
	}

	r1 := require.New(t)
	s1, err1 := LinkTo("this should fail", nil, nil)
	r1.Empty(s1)
	r1.Error(err1)

	s1, err1 = RemoteLinkTo("this should work", nil, helptest.NewContext())
	r1.Equal(template.HTML(`<a data-remote="true" href="/this should work"></a>`), s1)
	r1.NoError(err1)

	for _, tt := range table {
		t.Run(tt.out, func(st *testing.T) {
			r := require.New(st)
			c := helptest.NewContext()
			if len(tt.body) != 0 {
				c.BlockFn = func() (string, error) {
					return tt.body, nil
				}
			}
			if tt.errBlock {
				c.BlockFn = func() (string, error) {
					return "", errors.New("nope")
				}
			}
			s, err := LinkTo(tt.in, tt.opts, c)
			if tt.err {
				r.Error(err)
				return
			}
			r.NoError(err)
			r.Equal(template.HTML(tt.out), s)
		})
	}
}
