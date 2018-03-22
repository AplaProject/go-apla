package api

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInterfaceRow(t *testing.T) {
	cases := []struct {
		url        string
		contract   string
		equalAttrs []string
	}{
		{"interface/page/", "NewPage", []string{"Name", "Value", "Menu", "Conditions"}},
		{"interface/menu/", "NewMenu", []string{"Name", "Value", "Title", "Conditions"}},
		{"interface/block/", "NewBlock", []string{"Name", "Value", "Conditions"}},
	}

	checkEqualAttrs := func(form url.Values, result map[string]interface{}, equalKeys []string) {
		for _, key := range equalKeys {
			v := result[strings.ToLower(key)]
			assert.EqualValues(t, form.Get(key), v)
		}
	}

	errUnauthorized := `401 {"error": "E_UNAUTHORIZED", "msg": "Unauthorized" }`
	for _, c := range cases {
		assert.EqualError(t, sendGet(c.url+"-", &url.Values{}, nil), errUnauthorized)
	}

	assert.NoError(t, keyLogin(1))

	for _, c := range cases {
		name := randName("component")
		form := url.Values{
			"Name": {name}, "Value": {"value"}, "Menu": {"default_menu"}, "Title": {"title"},
			"Conditions": {"true"},
		}
		assert.NoError(t, postTx(c.contract, &form))
		result := map[string]interface{}{}
		assert.NoError(t, sendGet(c.url+name, &url.Values{}, &result))
		checkEqualAttrs(form, result, c.equalAttrs)
	}
}
