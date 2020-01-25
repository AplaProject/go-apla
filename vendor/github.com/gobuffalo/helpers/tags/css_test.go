package tags

import (
	"html/template"
	"reflect"
	"testing"

	"github.com/gobuffalo/tags"
)

func TestCSS(t *testing.T) {
	type args struct {
		href    string
		options tags.Options
	}
	tests := []struct {
		name string
		args args
		want template.HTML
	}{
		{"normal empty css", args{"", tags.Options{}}, template.HTML(`<link href="" media="screen" rel="stylesheet" />`)},
		{"normal css", args{"yes.css", tags.Options{}}, template.HTML(`<link href="yes.css" media="screen" rel="stylesheet" />`)},
		{"normal css with overrides", args{"yes.css", tags.Options{"media": "foo", "rel": "bar"}}, template.HTML(`<link href="yes.css" media="foo" rel="bar" />`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CSS(tt.args.href, tt.args.options); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CSS() = %v, want %v", got, tt.want)
			}
		})
	}
}
