package form_test

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/gobuffalo/tags"
	"github.com/gobuffalo/tags/form"
	"github.com/stretchr/testify/require"
)

type Talk struct {
	Date time.Time `format:"01-02-2006"`
}

func Test_NewFormFor(t *testing.T) {
	r := require.New(t)

	f := form.NewFormFor(Talk{}, tags.Options{
		"action": "/users/1",
	})
	r.Equal("form", f.Name)
	r.Equal(`<form action="/users/1" id="talk-form" method="POST"></form>`, f.String())
}

func Test_FormFor_InputValue(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{
		"action": "/users/1",
	})

	l := f.InputTag("Name", tags.Options{"value": "Something"})

	r.Equal(`<input id="talk-Name" name="Name" type="text" value="Something" />`, l.String())
}

func Test_FormFor_InputHiddenValue(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{
		"action": "/users/1",
	})

	l := f.HiddenTag("Name", tags.Options{"value": "Something"})

	r.Equal(`<input id="talk-Name" name="Name" type="hidden" value="Something" />`, l.String())
}

func Test_FormFor_Input_BeforeTag_Opt(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{})

	s := `<span>Content</span>`
	l := f.InputTag("Test", tags.Options{"before_tag": s})

	r.Equal(`<span>Content</span><input id="talk-Test" name="Test" type="text" value="" />`, l.String())
}

func Test_FormFor_Input_AfterTag_Opt(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{})

	b := `<button>Button</button>`
	l := f.InputTag("Test", tags.Options{"after_tag": b})

	r.Equal(`<input id="talk-Test" name="Test" type="text" value="" /><button>Button</button>`, l.String())
}

func Test_FormFor_File(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{
		"action": "/users/1",
	})

	l := f.FileTag("Name", tags.Options{"value": "Something"})

	r.Equal(`<input id="talk-Name" name="Name" type="file" value="Something" />`, l.String())
}

func Test_FormFor_InputValueFormat(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{
		"action": "/users/1",
	})

	l := f.InputTag("Date", tags.Options{})
	r.Equal(`<input id="talk-Date" name="Date" type="text" value="01-01-0001" />`, l.String())

	l = f.InputTag("Date", tags.Options{"format": "01/02"})
	r.Equal(`<input id="talk-Date" name="Date" type="text" value="01/01" />`, l.String())
}

func Test_NewFormFor_With_AuthenticityToken(t *testing.T) {
	r := require.New(t)

	f := form.NewFormFor(Talk{}, tags.Options{
		"action": "/users/1",
	})
	f.SetAuthenticityToken("12345")
	r.Equal("form", f.Name)
	r.Equal(`<form action="/users/1" id="talk-form" method="POST"><input name="authenticity_token" type="hidden" value="12345" /></form>`, f.String())
}

func Test_NewFormFor_With_NotPostMethod(t *testing.T) {
	r := require.New(t)

	f := form.NewFormFor(Talk{}, tags.Options{
		"action": "/users/1",
		"method": "put",
	})
	r.Equal("form", f.Name)
	r.Equal(`<form action="/users/1" id="talk-form" method="POST"><input name="_method" type="hidden" value="PUT" /></form>`, f.String())
}

func Test_FormFor_Label(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{})
	l := f.Label("Name", tags.Options{})
	r.Equal(`<label>Name</label>`, l.String())
}

func Test_FormFor_FieldDoesntExist(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{})
	l := f.InputTag("IDontExist", tags.Options{})
	r.Equal(`<input id="talk-IDontExist" name="IDontExist" type="text" value="" />`, l.String())
}

func Test_FormFor_HiddenTag(t *testing.T) {
	r := require.New(t)
	f := form.NewFormFor(Talk{}, tags.Options{})
	l := f.HiddenTag("Name", tags.Options{})
	r.Equal(`<input id="talk-Name" name="Name" type="hidden" value="" />`, l.String())
}

func Test_FormFor_NullableField(t *testing.T) {
	r := require.New(t)
	model := struct {
		Name       string
		CreditCard nullString
		Floater    sql.NullFloat64
		Other      sql.NullBool
	}{
		CreditCard: NewNullString("Hello"),
	}

	f := form.NewFormFor(model, tags.Options{})

	cases := map[string][]string{
		"CreditCard": {`<input id="-CreditCard" name="CreditCard" type="text" value="Hello" />`},
		"Floater":    {`<input id="-Floater" name="Floater" type="text" value="" />`},
		"Other":      {`<input id="-Other" name="Other" type="text" value="" />`},
	}

	for field, html := range cases {
		l := f.InputTag(field, tags.Options{})
		r.Equal(html[0], l.String())
	}
}

type Person struct {
	Name       string
	Address    Address
	References []Address
	Contacts   []string
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

	f := form.NewFormFor(p, tags.Options{})
	tag := f.InputTag("Address.State", tags.Options{})

	exp := `<input id="person-Address.State" name="Address.State" type="text" value="MA" />`
	r.Equal(exp, tag.String())
}

func Test_FormFor_Nested_Slice_Struct(t *testing.T) {
	r := require.New(t)
	p := Person{
		Name: "Mark",
		Address: Address{
			City:  "Boston",
			State: "MA",
		},
	}
	p.References = []Address{p.Address}

	f := form.NewFormFor(p, tags.Options{})
	tag := f.InputTag("References[0].City", tags.Options{})

	exp := `<input id="person-References[0].City" name="References[0].City" type="text" value="Boston" />`
	r.Equal(exp, tag.String())
}

func Test_FormFor_Nested_Slice_String(t *testing.T) {
	r := require.New(t)
	p := Person{
		Contacts: []string{
			"Contact 1",
			"Contact 2",
			"Contact 3",
		},
	}

	f := form.NewFormFor(p, tags.Options{})

	for i := 0; i < len(p.Contacts); i++ {
		expectedValue := p.Contacts[i]
		index := strconv.Itoa(i)
		tag := f.InputTag("Contacts["+index+"]", tags.Options{})
		exp := `<input id="person-Contacts[` + index + `]" name="Contacts[` + index + `]" type="text" value="` + expectedValue + `" />`
		r.Equal(exp, tag.String())
	}
}

func Test_FormFor_Nested_Slice_String_Pointer(t *testing.T) {
	r := require.New(t)
	p := struct {
		Contacts *[]string
	}{
		&[]string{"Contact 1", "Contact 2"},
	}

	f := form.NewFormFor(p, tags.Options{})
	tag := f.InputTag("Contacts[0]", tags.Options{})

	exp := `<input id="-Contacts[0]" name="Contacts[0]" type="text" value="Contact 1" />`
	r.Equal(exp, tag.String())
}

func Test_FormFor_Nested_Slice_Pointer(t *testing.T) {
	r := require.New(t)
	p := struct {
		Persons *[]Person
	}{
		&[]Person{{Name: "Mark"}, {Name: "Clayton"}, {Name: "Iain"}},
	}

	f := form.NewFormFor(p, tags.Options{})

	for i := 0; i < len(*p.Persons); i++ {
		expectedValue := (*p.Persons)[i].Name
		index := strconv.Itoa(i)
		tag := f.InputTag("Persons["+index+"].Name", tags.Options{})
		exp := `<input id="-Persons[` + index + `].Name" name="Persons[` + index + `].Name" type="text" value="` + expectedValue + `" />`
		r.Equal(exp, tag.String())
	}
}

func Test_FormFor_Nested_Slice_Pointer_Elements(t *testing.T) {
	r := require.New(t)
	p := struct {
		Persons []*Person
	}{
		[]*Person{
			&Person{Name: "Mark"},
		},
	}

	f := form.NewFormFor(p, tags.Options{})
	tag := f.InputTag("Persons[0].Name", tags.Options{})

	exp := `<input id="-Persons[0].Name" name="Persons[0].Name" type="text" value="Mark" />`
	r.Equal(exp, tag.String())
}

func Test_FormFor_Nested_Slice_With_Sub_Slices(t *testing.T) {
	r := require.New(t)
	p := struct {
		Persons *[]Person
	}{
		&[]Person{
			{
				Name: "Mark",
				References: []Address{
					{City: "Boston"},
				},
			},
		},
	}

	f := form.NewFormFor(p, tags.Options{})
	tag := f.InputTag("Persons[0].References[0].City", tags.Options{})

	exp := `<input id="-Persons[0].References[0].City" name="Persons[0].References[0].City" type="text" value="Boston" />`
	r.Equal(exp, tag.String())
}

func Test_FormFor_DateTimeTag(t *testing.T) {
	r := require.New(t)

	date, err := time.Parse("2006-01-02T03:04", "1976-08-24T06:17")
	r.NoError(err)

	p := struct {
		BirthDate time.Time
	}{
		BirthDate: date,
	}

	f := form.NewFormFor(p, tags.Options{})
	i := f.DateTimeTag("BirthDate", tags.Options{})
	r.Equal(`<input id="-BirthDate" name="BirthDate" type="datetime-local" value="1976-08-24T06:17" />`, i.String())
}
