package validators_test

import (
	"testing"
	"time"

	"github.com/gobuffalo/validate"
	. "github.com/gobuffalo/validate/validators"
	"github.com/stretchr/testify/require"
)

func Test_TimeIsPresent(t *testing.T) {
	r := require.New(t)
	v := TimeIsPresent{Name: "Created At", Field: time.Now()}
	errors := validate.NewErrors()
	v.IsValid(errors)
	r.Equal(0, errors.Count())

	v = TimeIsPresent{Name: "Created At", Field: time.Time{}}
	v.IsValid(errors)
	r.Equal(1, errors.Count())
	r.Equal(errors.Get("created_at"), []string{"Created At can not be blank."})

	errors = validate.NewErrors()
	v = TimeIsPresent{Name: "Created At", Field: time.Time{}, Message: "Field can't be blank."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("created_at"), []string{"Field can't be blank."})

	errors = validate.NewErrors()
	v = TimeIsPresent{"Created At", time.Time{}, "Field can't be blank."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("created_at"), []string{"Field can't be blank."})
}
