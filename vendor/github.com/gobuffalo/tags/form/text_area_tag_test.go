package form_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/gobuffalo/tags"
	"github.com/gobuffalo/tags/form"
	"github.com/stretchr/testify/require"
)

func Test_Form_TextArea(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	ta := f.TextArea(tags.Options{
		"value": "hi",
	})
	r.Equal(`<textarea>hi</textarea>`, ta.String())
}

func Test_Form_TextArea_nullsString(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	ta := f.TextArea(tags.Options{
		"value": NewNullString("hi"),
	})
	r.Equal(`<textarea>hi</textarea>`, ta.String())
}

func Test_Form_TextArea_nullsString_empty(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	ta := f.TextArea(tags.Options{
		"value": nullString{},
	})
	r.Equal(`<textarea></textarea>`, ta.String())
}

func Test_Form_TextArea_Escaped(t *testing.T) {
	r := require.New(t)
	f := form.New(tags.Options{})
	ta := f.TextArea(tags.Options{
		"value": "<b>This should not be bold</b>",
	})
	r.Equal(`<textarea>&lt;b&gt;This should not be bold&lt;/b&gt;</textarea>`, ta.String())
}

type nullString sql.NullString

func (ns nullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

func NewNullString(data string) nullString {
	return nullString{
		String: data,
		Valid:  true,
	}
}

func (ns nullString) Interface() interface{} {
	if !ns.Valid {
		return nil
	}
	return ns.String
}
