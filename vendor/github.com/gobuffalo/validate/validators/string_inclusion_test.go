package validators_test

import (
	"testing"

	"github.com/gobuffalo/validate"
	. "github.com/gobuffalo/validate/validators"
	"github.com/stretchr/testify/require"
)

func Test_StringInclusion(t *testing.T) {
	r := require.New(t)

	l := []string{"Mark", "Bates"}

	v := StringInclusion{Name: "Name", Field: "Mark", List: l}
	errors := validate.NewErrors()
	v.IsValid(errors)
	r.Equal(errors.Count(), 0)

	v = StringInclusion{Name: "Name", Field: "Foo", List: l}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Name is not in the list [Mark, Bates]."})

	errors = validate.NewErrors()
	v = StringInclusion{Name: "Name", Field: "Foo", Message: "Name is not in the list (Mark, Bates).", List: l}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Name is not in the list (Mark, Bates)."})

	errors = validate.NewErrors()
	v = StringInclusion{"Name", "Foo", l, "Name is not in the list (Mark, Bates)."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Name is not in the list (Mark, Bates)."})
}
