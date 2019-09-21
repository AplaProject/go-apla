package validators_test

import (
	"testing"

	"github.com/gobuffalo/validate"
	. "github.com/gobuffalo/validate/validators"
	uuid "github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func Test_UUIDIsPresent(t *testing.T) {
	r := require.New(t)

	id, err := uuid.NewV4()
	r.NoError(err)
	v := UUIDIsPresent{Name: "Name", Field: id}
	errors := validate.NewErrors()
	v.IsValid(errors)
	r.Equal(errors.Count(), 0)

	v = UUIDIsPresent{Name: "Name", Field: uuid.UUID{}}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Name can not be blank."})

	errors = validate.NewErrors()
	v = UUIDIsPresent{Name: "Name", Field: uuid.UUID{}, Message: "Field can't be blank."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Field can't be blank."})

	errors = validate.NewErrors()
	v = UUIDIsPresent{"Name", uuid.UUID{}, "Field can't be blank."}
	v.IsValid(errors)
	r.Equal(errors.Count(), 1)
	r.Equal(errors.Get("name"), []string{"Field can't be blank."})
}
