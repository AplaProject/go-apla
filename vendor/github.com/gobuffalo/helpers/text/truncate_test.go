package text

import (
	"testing"

	"github.com/gobuffalo/helpers/hctx"
	"github.com/stretchr/testify/require"
)

func Test_Truncate(t *testing.T) {
	r := require.New(t)
	x := "KEuFHyyImKUMhSkSolLqgqevKQNZUjpSZokrGbZqnUrUnWrTDwi"
	s := Truncate(x, hctx.Map{})
	r.Len(s, 50)
	r.Equal("...", s[47:])

	s = Truncate(x, hctx.Map{
		"size": 10,
	})
	r.Len(s, 10)
	r.Equal("...", s[7:])

	s = Truncate(x, hctx.Map{
		"size":  10,
		"trail": "more",
	})
	r.Len(s, 10)
	r.Equal("more", s[6:])

	// Case size < len(trail)
	s = Truncate(x, hctx.Map{
		"size":  3,
		"trail": "more",
	})
	r.Len(s, 4)
	r.Equal("more", s)

	// Case size >= len(string)
	s = Truncate(x, hctx.Map{
		"size": len(x),
	})
	r.Len(s, len(x))
	r.Equal(x[48:], s[48:])
}
