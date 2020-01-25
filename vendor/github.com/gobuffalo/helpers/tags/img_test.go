package tags

import (
	"html/template"
	"reflect"
	"testing"

	"github.com/gobuffalo/tags"
)

func TestImg(t *testing.T) {
	type args struct {
		src     string
		options tags.Options
	}
	tests := []struct {
		name string
		args args
		want template.HTML
	}{
		{"normal empty img", args{"", tags.Options{}}, template.HTML(`<img src="" />`)},
		{"normal img", args{"testing.png", tags.Options{}}, template.HTML(`<img src="testing.png" />`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Img(tt.args.src, tt.args.options); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Img() = %v, want %v", got, tt.want)
			}
		})
	}
}
