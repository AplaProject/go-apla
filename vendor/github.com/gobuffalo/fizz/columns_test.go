package fizz_test

import (
	"testing"

	"github.com/gobuffalo/fizz"
	"github.com/stretchr/testify/require"
)

func Test_Column_Stringer(t *testing.T) {
	t.Run("primary column", func(tt *testing.T) {
		r := require.New(tt)
		c := fizz.Column{
			Name:    "pk",
			ColType: "int",
			Primary: true,
		}

		r.Equal(`t.Column("pk", "int", {primary: true})`, c.String())
	})

	t.Run("primary column with raw default", func(tt *testing.T) {
		r := require.New(tt)
		c := fizz.Column{
			Name:    "pk",
			ColType: "int",
			Primary: true,
			Options: map[string]interface{}{
				"default_raw": "uuid_generate_v4()",
			},
		}

		r.Equal(`t.Column("pk", "int", {default_raw: "uuid_generate_v4()", primary: true})`, c.String())
	})

	t.Run("simple column", func(tt *testing.T) {
		r := require.New(tt)
		c := fizz.Column{
			Name:    "name",
			ColType: "string",
		}

		r.Equal(`t.Column("name", "string")`, c.String())
	})

	t.Run("with option", func(tt *testing.T) {
		r := require.New(tt)
		c := fizz.Column{
			Name:    "alive",
			ColType: "boolean",
			Options: map[string]interface{}{
				"null": true,
			},
		}

		r.Equal(`t.Column("alive", "boolean", {null: true})`, c.String())
	})

	t.Run("with string option", func(tt *testing.T) {
		r := require.New(tt)
		c := fizz.Column{
			Name:    "price",
			ColType: "numeric",
			Options: map[string]interface{}{
				"default": "1.00",
			},
		}

		r.Equal(`t.Column("price", "numeric", {default: "1.00"})`, c.String())
	})
}
