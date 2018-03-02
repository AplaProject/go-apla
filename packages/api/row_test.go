package api

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIRow(t *testing.T) {
	name := randName("page")
	prefixUrl := "rowbycolumn/pages/name/"
	prefixUrlByID := "row/pages/"
	errUnauthorized := `401 {"error": "E_UNAUTHORIZED", "msg": "Unauthorized" }`

	assert.EqualError(t, sendGet(prefixUrl+name, &url.Values{}, nil), errUnauthorized)
	assert.EqualError(t, sendGet(prefixUrlByID+name, &url.Values{}, nil), errUnauthorized)

	assert.NoError(t, keyLogin(1))

	form := url.Values{
		"Name": {name}, "Value": {"P()"}, "Menu": {"default_menu"},
		"Conditions": {"true"},
	}
	assert.NoError(t, postTx("NewPage", &form))

	checkEqualAttrs := func(form url.Values, result *rowResult) {
		equalKeys := []string{"Name", "Value", "Menu", "Conditions"}
		for _, key := range equalKeys {
			assert.Equal(t, form.Get(key), result.Value[strings.ToLower(key)])
		}
	}

	result := &rowResult{}
	assert.NoError(t, sendGet(prefixUrl+name, &url.Values{}, result))
	checkEqualAttrs(form, result)

	id := result.Value["id"]
	result = &rowResult{}
	assert.NoError(t, sendGet(prefixUrlByID+id, &url.Values{}, result))
	checkEqualAttrs(form, result)
}
