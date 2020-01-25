package forms

import (
	"html/template"
	"reflect"
	"testing"

	"github.com/gobuffalo/helpers/hctx"
	"github.com/gobuffalo/helpers/helptest"
	"github.com/gobuffalo/tags"
)

type Car struct {
	ID int
}

type boat struct {
	Slug string
}

type plane struct{}

func (plane) ToParam() string {
	return "aeroplane"
}

type truck struct{}

func (truck) ToPath() string {
	return "/a/truck"
}

type BadCar struct {
	IDs int
}

func TestFormFor(t *testing.T) {
	type args struct {
	}

	hcNoBlock := helptest.NewContext()
	hc1 := helptest.NewContext()
	hc1.BlockFn = func() (string, error) {
		return "good", nil
	}
	hc1.Set("errors", map[string][]string{})
	hc2 := helptest.NewContext()
	hc2.BlockFn = func() (string, error) {
		return "good", nil
	}
	hc2.Set("authenticity_token", "myToken")

	tests := []struct {
		name    string
		model   interface{}
		opts    tags.Options
		help    hctx.HelperContext
		want    template.HTML
		wantErr bool
		remote  bool
	}{
		{"No Block Given", Car{1}, tags.Options{}, hcNoBlock, template.HTML(""), true, false},
		{"Normal form for individual", Car{1}, nil, hc1, template.HTML(`<form action="/cars/1" id="car-form" method="POST">good</form>`), false, false},
		{"Normal form for group", Car{}, tags.Options{}, hc1, template.HTML(`<form action="/cars" id="car-form" method="POST">good</form>`), false, false},
		{"Normal form for pointer group", &Car{}, tags.Options{}, hc1, template.HTML(`<form action="/cars" id="car-form" method="POST">good</form>`), false, false},
		{"Normal form for model", boat{"titanic"}, tags.Options{"var": "foo", "errors": map[string][]string{}}, hc1, template.HTML(`<form action="/boats/titanic" id="boat-form" method="POST">good</form>`), false, false},
		{"Remote form for model", plane{}, tags.Options{"var": "foo"}, hc1, template.HTML(`<form action="/planes/aeroplane" data-remote="true" id="plane-form" method="POST">good</form>`), false, true},
		{"Remote form for complex path", truck{}, nil, hc1, template.HTML(`<form action="/a/truck" data-remote="true" id="truck-form" method="POST">good</form>`), false, true},
		{"Remote form for complex path", []interface{}{truck{}, plane{}}, nil, hc2, template.HTML(`<form action="/a/truck/planes/aeroplane" data-remote="true" id="-form" method="POST"><input name="authenticity_token" type="hidden" value="myToken" />good</form>`), false, true},
		{"Remote form for complex path", "foo", nil, hc2, template.HTML(`<form action="/foo" id="string-form" method="POST"><input name="authenticity_token" type="hidden" value="myToken" />good</form>`), false, false},
		{"Remote form for complex path", template.HTML("foo"), nil, hc2, template.HTML(`<form action="/foo" id="html-form" method="POST"><input name="authenticity_token" type="hidden" value="myToken" />good</form>`), false, false},
		{"Bad model for pathing", map[int]int{}, nil, hc1, template.HTML(``), true, false},
		{"Bad model for pathing", nil, nil, hc1, template.HTML(``), true, false},
		{"Bad model for pathing", []interface{}{truck{}, nil}, tags.Options{}, hc1, template.HTML(``), true, true},
		{"Bad model for pathing", BadCar{}, tags.Options{}, hc1, template.HTML(``), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.remote {
				got, err := RemoteFormFor(tt.model, tt.opts, tt.help)
				if (err != nil) != tt.wantErr {
					t.Errorf("RemoteFormFor() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RemoteFormFor() = %v, want %v", got, tt.want)
				}
			} else {
				got, err := FormFor(tt.model, tt.opts, tt.help)
				if (err != nil) != tt.wantErr {
					t.Errorf("FormFor() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("FormFor() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
