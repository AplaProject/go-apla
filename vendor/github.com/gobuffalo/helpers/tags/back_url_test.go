package tags

import (
	"net/http"
	"testing"

	"github.com/gobuffalo/helpers/helptest"
	"github.com/stretchr/testify/require"
)

func TestBackURL(t *testing.T) {

	req, _ := http.NewRequest("GET", "https://wawand.co/contact", nil)
	req.Header.Add("Referer", "https://gobuffalo.io")

	req2, _ := http.NewRequest("GET", "https://wawand.co/contact", nil)

	testCases := []struct {
		name        string
		request     interface{}
		expectedURL string
	}{
		{name: "RefererIncluded", request: req, expectedURL: "https://gobuffalo.io"},
		{name: "RequestNotRequest", request: "not-request", expectedURL: "javascript:history.back()"},
		{name: "RequestNotRequest", request: req2, expectedURL: "javascript:history.back()"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(st *testing.T) {
			c := helptest.NewContext()
			c.Set("request", testCase.request)

			r := require.New(st)
			r.Equal(testCase.expectedURL, BackURL(c))
		})
	}
}
