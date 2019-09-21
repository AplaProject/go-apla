package validators

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/validate"
	"github.com/gofrs/uuid"
)

type UUIDIsPresent struct {
	Name    string
	Field   uuid.UUID
	Message string
}

// IsValid adds an error if the field is not a valid uuid.
func (v *UUIDIsPresent) IsValid(errors *validate.Errors) {
	s := v.Field.String()
	if strings.TrimSpace(s) != "" && v.Field != uuid.Nil {
		return
	}

	if len(v.Message) > 0 {
		errors.Add(GenerateKey(v.Name), v.Message)
		return
	}

	errors.Add(GenerateKey(v.Name), fmt.Sprintf("%s can not be blank.", v.Name))
}
