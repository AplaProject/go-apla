package validators

import (
	"testing"

	"github.com/gobuffalo/validate"
	"github.com/stretchr/testify/require"
)

func Test_URLIsPresent(t *testing.T) {
	r := require.New(t)

	var tests = []struct {
		url   string
		valid bool
	}{
		{"", false},
		{"http://", false},
		{"https://", false},
		{"http", false},
		{"google.com", false},
		{"http://www.google.com", true},
		{"http://google.com", true},
		{"google.com", false},
		{"https://www.google.cOM", true},
		{"ht123tps://www.google.cOM", false},
		{"https://golang.Org", true},
		{"https://invalid#$%#$@.Org", false},
	}
	for _, test := range tests {
		v := URLIsPresent{Name: "URL", Field: test.url}
		errors := validate.NewErrors()
		v.IsValid(errors)
		r.Equal(test.valid, !errors.HasAny(), test.url, errors.Error())
	}
	v := URLIsPresent{Name: "URL", Field: "http://", Message: "URL isn't valid."}
	errors := validate.NewErrors()
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("url"), []string{"URL isn't valid."})
}
