package text

import (
	"testing"

	"github.com/gobuffalo/helpers/helptest"
	"github.com/stretchr/testify/require"
)

func Test_MarkdownHelper(t *testing.T) {
	r := require.New(t)

	hc := helptest.NewContext()
	s, err := Markdown("# H1", hc)
	r.NoError(err)
	r.Contains(s, "H1</h1>")
}

func Test_MarkdownHelper_WithBlock(t *testing.T) {
	r := require.New(t)

	hc := helptest.NewContext()
	hc.BlockFn = func() (string, error) {
		return "# H2", nil
	}

	s, err := Markdown("", hc)
	r.NoError(err)
	r.Contains(s, "H2</h1>")
}
