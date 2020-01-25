package validators_test

import (
	"testing"

	"github.com/gobuffalo/validate"
	. "github.com/gobuffalo/validate/validators"
	"github.com/stretchr/testify/require"
)

func Test_IntArrayIsPresent(t *testing.T) {
	r := require.New(t)

	v := IntArrayIsPresent{Name: "Name", Field: []int{1}}
	errors := validate.NewErrors()
	v.IsValid(errors)
	r.Equal(errors.Count(), 0)

	v = IntArrayIsPresent{Name: "Name", Field: []int{}}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Name can not be empty."})

	errors = validate.NewErrors()
	v = IntArrayIsPresent{Name: "Name", Field: []int{}, Message: "Field can't be blank."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Field can't be blank."})

	errors = validate.NewErrors()
	v = IntArrayIsPresent{"Name", []int{}, "Field can't be blank."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Field can't be blank."})
}
