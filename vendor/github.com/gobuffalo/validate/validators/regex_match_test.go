package validators_test

import (
	"testing"

	"github.com/gobuffalo/validate"
	. "github.com/gobuffalo/validate/validators"
	"github.com/stretchr/testify/require"
)

func Test_RegexMatch(t *testing.T) {
	r := require.New(t)

	v := RegexMatch{Name: "Phone", Field: "555-555-5555", Expr: "^([0-9]{3}-[0-9]{3}-[0-9]{4})$"}
	errors := validate.NewErrors()
	v.IsValid(errors)
	r.Equal(errors.Count(), 0)

	v = RegexMatch{Name: "Phone", Field: "123-ab1-1424", Expr: "^([0-9]{3}-[0-9]{3}-[0-9]{4})$"}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("phone"), []string{"Phone does not match the expected format."})

	errors = validate.NewErrors()
	v = RegexMatch{Name: "Phone", Field: "123-ab1-1424", Expr: "^([0-9]{3}-[0-9]{3}-[0-9]{4})$", Message: "Phone number does not match the expected format."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("phone"), []string{"Phone number does not match the expected format."})

	errors = validate.NewErrors()
	v = RegexMatch{"Phone", "123-ab1-1424", "^([0-9]{3}-[0-9]{3}-[0-9]{4})$", "Phone number does not match the expected format."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("phone"), []string{"Phone number does not match the expected format."})
}
