package network

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyBodyResponse(t *testing.T) {
	buf := []byte{}
	w := bytes.NewBuffer(buf)
	empty := &GetBodyResponse{}
	require.NoError(t, empty.Write(w))

	r := bytes.NewReader(w.Bytes())
	emptyRes := &GetBodyResponse{}
	require.NoError(t, emptyRes.Read(r))
}
