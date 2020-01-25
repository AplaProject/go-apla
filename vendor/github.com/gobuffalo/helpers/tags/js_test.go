package tags

import (
	"html/template"
	"reflect"
	"testing"

	"github.com/gobuffalo/tags"
)

func TestJS(t *testing.T) {
	type args struct {
		src     string
		options tags.Options
	}
	tests := []struct {
		name string
		args args
		want template.HTML
	}{
		{"normal empty js", args{"", tags.Options{}}, template.HTML(`<script src="" type="text/javascript"></script>`)},
		{"normal js", args{"app.js", tags.Options{}}, template.HTML(`<script src="app.js" type="text/javascript"></script>`)},
		{"normal js with overrides", args{"app.js", tags.Options{"type": "real/real"}}, template.HTML(`<script src="app.js" type="real/real"></script>`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JS(tt.args.src, tt.args.options); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JS() = %v, want %v", got, tt.want)
			}
		})
	}
}
