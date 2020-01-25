package bootstrap_test

import (
	"fmt"
	"testing"

	"github.com/gobuffalo/tags"
	"github.com/gobuffalo/tags/form/bootstrap"
	"github.com/gobuffalo/validate"
	"github.com/stretchr/testify/require"
)

func Test_InputFieldLabel(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})
	l := f.InputTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label>Custom</label><input class=" form-control" id="-Name" name="Name" type="text" value="" /></div>`, l.String())
}

func Test_InputFieldLabelWithAchronym(t *testing.T) {
	cases := map[string]string{
		"URL":            "URL",
		"MyURL":          "My URL",
		"SimpleURIAdded": "Simple URI Added",
		"GaveAnExample":  "Gave An Example",
	}
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ URL string }{}, tags.Options{})

	for key, expectedLabel := range cases {
		l := f.InputTag(key, tags.Options{})
		r.Equal(`<div class="form-group"><label>`+expectedLabel+`</label><input class=" form-control" id="-`+key+`" name="`+key+`" type="text" value="" /></div>`, l.String())
	}
}

func Test_InputFieldLabel_Humanized(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ LongName string }{}, tags.Options{})
	l := f.InputTag("LongName", tags.Options{})
	r.Equal(`<div class="form-group"><label>Long Name</label><input class=" form-control" id="-LongName" name="LongName" type="text" value="" /></div>`, l.String())
}

func Test_InputFieldSchema(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct {
		Name string `schema:"notName"`
	}{}, tags.Options{})

	l := f.InputTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label>Custom</label><input class=" form-control" id="-notName" name="notName" type="text" value="" /></div>`, l.String())
}

func Test_InputFieldFormInsteadOfSchema(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct {
		Name string `form:"notName"`
	}{}, tags.Options{})

	l := f.InputTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label>Custom</label><input class=" form-control" id="-notName" name="notName" type="text" value="" /></div>`, l.String())
}

func Test_InputFieldFormAndSchema(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct {
		Name string `form:"notName" schema:"name"`
	}{}, tags.Options{})

	l := f.InputTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label>Custom</label><input class=" form-control" id="-notName" name="notName" type="text" value="" /></div>`, l.String())
}

func Test_InputFieldSchema_FieldNotPresent(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct {
		Name string `schema:"notName"`
	}{}, tags.Options{})

	l := f.InputTag("Other", tags.Options{})
	r.Equal(`<div class="form-group"><label>Other</label><input class=" form-control" id="-Other" name="Other" type="text" value="" /></div>`, l.String())
}

func Test_InputFieldSchema_FieldDash(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct {
		Name string `schema:"-"`
	}{}, tags.Options{})

	l := f.InputTag("Name", tags.Options{})
	r.Equal(`<div class="form-group"><label>Name</label><input class=" form-control" id="-Name" name="Name" type="text" value="" /></div>`, l.String())
}

func Test_SelectLabel(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})
	l := f.SelectTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label>Custom</label><select class=" form-control" id="-Name" name="Name"></select></div>`, l.String())
}

func Test_Select_With_String_As_BeforeTag_Opt(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})

	s := `<span>Test</span>`
	l := f.SelectTag("Name", tags.Options{"before_tag": s})

	r.Equal(`<div class="form-group"><label>Name</label><span>Test</span><select class=" form-control" id="-Name" name="Name"></select></div>`, l.String())
}

func Test_Select_With_Nested_Tag_As_BeforeTag_Opt(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})

	s := tags.New("span", tags.Options{"body": "Test"})
	l := f.SelectTag("Name", tags.Options{"before_tag": s})

	r.Equal(`<div class="form-group"><label>Name</label><span>Test</span><select class=" form-control" id="-Name" name="Name"></select></div>`, l.String())
}

func Test_Select_With_String_As_AfterTag_Opt(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})

	b := `<button type="button">Button Name</button>`
	l := f.SelectTag("Name", tags.Options{"after_tag": b})

	r.Equal(`<div class="form-group"><label>Name</label><select class=" form-control" id="-Name" name="Name"></select><button type="button">Button Name</button></div>`, l.String())
}

func Test_Select_With_Nested_Tag_As_AfterTag_Opt(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})

	b := tags.New("button", tags.Options{
		"body": "Button Name",
		"type": "button",
	})
	l := f.SelectTag("Name", tags.Options{"after_tag": b})

	r.Equal(`<div class="form-group"><label>Name</label><select class=" form-control" id="-Name" name="Name"></select><button type="button">Button Name</button></div>`, l.String())
}

func Test_RadioButton(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})
	l := f.RadioButton("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label>Custom</label><label for="-Name"><input class=" form-control" id="-Name" name="Name" type="radio" value="" /> </label></div>`, l.String())
}
func Test_TextArea(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})
	l := f.TextArea("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label>Custom</label><textarea class=" form-control" id="-Name" name="Name"></textarea></div>`, l.String())
}

func Test_CheckBox(t *testing.T) {
	r := require.New(t)
	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{})
	l := f.CheckboxTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label><input class="" id="-Name" name="Name" type="checkbox" value="true" /> Custom</label></div>`, l.String())
}

func Test_InputError(t *testing.T) {
	r := require.New(t)

	errors := validate.NewErrors()
	errors.Add("name", "Name shoud be AJ.")

	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{"errors": errors})
	l := f.InputTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group has-error"><label>Custom</label><input class=" form-control is-invalid" id="-Name" name="Name" type="text" value="" /><div class="invalid-feedback help-block">Name shoud be AJ.</div></div>`, l.String())
}

func Test_InputHidden(t *testing.T) {
	r := require.New(t)

	errors := validate.NewErrors()
	errors.Add("name", "Name shoud be AJ.")

	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{"errors": errors})
	l := f.InputTag("Name", tags.Options{"type": "hidden"})
	r.Equal(`<input errors="[Name shoud be AJ.]" id="-Name" name="Name" tags-field="Name" type="hidden" value="" />`, l.String())

	l = f.HiddenTag("Name", tags.Options{})
	r.Equal(`<input errors="[Name shoud be AJ.]" id="-Name" name="Name" tags-field="Name" type="hidden" value="" />`, l.String())
}

func Test_InputError_Map(t *testing.T) {
	r := require.New(t)

	errors := map[string][]string{
		"name": {"Name shoud be AJ."},
	}

	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{"errors": errors})
	l := f.InputTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group has-error"><label>Custom</label><input class=" form-control is-invalid" id="-Name" name="Name" type="text" value="" /><div class="invalid-feedback help-block">Name shoud be AJ.</div></div>`, l.String())
}

func Test_InputError_InvalidMap(t *testing.T) {
	r := require.New(t)

	errors := map[string]string{
		"name": "Name shoud be AJ.",
	}

	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{"errors": errors})
	l := f.InputTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group"><label>Custom</label><input class=" form-control" id="-Name" name="Name" type="text" value="" /></div>`, l.String())
}

func Test_InputMultipleError(t *testing.T) {
	r := require.New(t)

	errors := validate.NewErrors()
	errors.Add("name", "Name shoud be AJ.")
	errors.Add("name", "Name shoud start with A.")

	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{"errors": errors})
	l := f.InputTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group has-error"><label>Custom</label><input class=" form-control is-invalid" id="-Name" name="Name" type="text" value="" /><div class="invalid-feedback help-block">Name shoud be AJ.</div><div class="invalid-feedback help-block">Name shoud start with A.</div></div>`, l.String())
}

func Test_CheckBoxError(t *testing.T) {
	r := require.New(t)

	errors := validate.NewErrors()
	errors.Add("name", "Name shoud be AJ.")

	f := bootstrap.NewFormFor(struct{ Name string }{}, tags.Options{"errors": errors})
	l := f.CheckboxTag("Name", tags.Options{"label": "Custom"})
	r.Equal(`<div class="form-group has-error"><label><input class=" is-invalid" id="-Name" name="Name" type="checkbox" value="true" /> Custom</label><div class="invalid-feedback help-block">Name shoud be AJ.</div></div>`, l.String())
}

type Person struct {
	Name    string
	Address Address
}

type Address struct {
	City  string
	State string
}

func Test_FormFor_Nested_Struct(t *testing.T) {
	r := require.New(t)
	p := Person{
		Name: "Mark",
		Address: Address{
			City:  "Boston",
			State: "MA",
		},
	}

	f := bootstrap.NewFormFor(p, tags.Options{})
	tag := f.InputTag("Address.State", tags.Options{})

	exp := `<div class="form-group"><label>Address State</label><input class=" form-control" id="person-Address.State" name="Address.State" type="text" value="MA" /></div>`
	r.Equal(exp, tag.String())
}

func Test_Field_TagOnly(t *testing.T) {
	f := bootstrap.NewFormFor(struct {
		Name string `schema:"-"`
	}{}, tags.Options{})

	cases := []struct {
		f      func(field string, opt tags.Options) *tags.Tag
		name   string
		opts   tags.Options
		output string
	}{
		{
			f:    f.InputTag,
			name: "Name",
			opts: tags.Options{
				"tag_only": true,
				"class":    "custom-input",
			},
			output: `<input class="custom-input" id="-Name" name="Name" type="text" value="" />`,
		},

		{
			f:    f.TextArea,
			name: "Name",
			opts: tags.Options{
				"tag_only": true,
			},
			output: `<textarea class="" id="-Name" name="Name"></textarea>`,
		},

		{
			f:    f.RadioButton,
			name: "Name",
			opts: tags.Options{
				"tag_only": true,
			},
			output: `<input class="" id="-Name" name="Name" type="radio" value="" />`,
		},

		{
			f:    f.CheckboxTag,
			name: "Name",
			opts: tags.Options{
				"tag_only": true,
			},
			output: `<input class="" id="-Name" name="Name" type="checkbox" value="true" />`,
		},

		{
			f:    f.SelectTag,
			name: "Name",
			opts: tags.Options{
				"tag_only": true,
			},
			output: `<select class="" id="-Name" name="Name"></select>`,
		},

		{
			f:    f.FileTag,
			name: "Name",
			opts: tags.Options{
				"tag_only": true,
			},
			output: `<input class="" id="-Name" name="Name" type="file" value="" />`,
		},
	}

	for index, tcase := range cases {
		t.Run(fmt.Sprintf("%v", index), func(tt *testing.T) {
			r := require.New(tt)
			l := tcase.f(tcase.name, tcase.opts)
			r.Equal(tcase.output, l.String())
		})
	}
}
