package validators_test

import (
	"testing"

	"github.com/gobuffalo/validate"
	. "github.com/gobuffalo/validate/validators"
	"github.com/stretchr/testify/require"
)

func Test_IntIsLessThan(t *testing.T) {
	r := require.New(t)

	v := IntIsLessThan{Name: "Number", Field: 1, Compared: 2}
	errors := validate.NewErrors()
	v.IsValid(errors)
	r.Equal(errors.Count(), 0)

	v = IntIsLessThan{Name: "number", Field: 1, Compared: 0}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("number"), []string{"1 is not less than 0."})

	v = IntIsLessThan{Name: "number", Field: 1, Compared: 0, Message: "number is not less than 0."}
	errors = validate.NewErrors()
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("number"), []string{"number is not less than 0."})
}
