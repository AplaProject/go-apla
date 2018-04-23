// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLang(t *testing.T) {
	requireLogin(t, 1)

	name := randName("lng")
	utfName := randName("lngutf")

	cases := []struct {
		url    string
		form   url.Values
		expect string
	}{
		{
			"NewLang",
			url.Values{
				"Name":  {name},
				"Trans": {`{"en": "My test", "fr": "French string", "en-US": "US locale"}`},
			},
			"",
		},
		{
			"NewLang",
			url.Values{
				"Name":  {utfName},
				"Trans": {`{"en": "тест"}`},
			},
			"",
		},
		{
			"NewPage",
			url.Values{
				"Name":       {name},
				"Value":      {fmt.Sprintf("Span($%s$)", name)},
				"Menu":       {"default_menu"},
				"Conditions": {"ContractConditions(`MainCondition`)"},
			},
			"",
		},
		{
			"content/page/" + name,
			url.Values{`lang`: {`fr`}},
			`[{"tag":"span","children":[{"tag":"text","text":"French string"}]}]`,
		},
		{
			"content/page/" + name,
			url.Values{`lang`: {`en-GB`}},
			`[{"tag":"span","children":[{"tag":"text","text":"My test"}]}]`,
		},
		{
			"content/page/" + name,
			url.Values{`lang`: {`en-US`}},
			`[{"tag":"span","children":[{"tag":"text","text":"US locale"}]}]`,
		},
		{
			"content",
			url.Values{
				`template`: {
					fmt.Sprintf(`Div(){
						Button(Body: $%[1]s$ $,  Page:test).Alert(Text: $%[1]s$, ConfirmButton: $confirm$, CancelButton: $cancel$)
						Button(Body: LangRes(%[1]s) LangRes, PageParams: "test", ).Alert(Text: $%[1]s$, CancelButton: $cancel$)
					}`, utfName),
				},
			},
			`[{"tag":"div","children":[{"tag":"button","attr":{"alert":{"cancelbutton":"$cancel$","confirmbutton":"$confirm$","text":"тест"},"page":"test"},"children":[{"tag":"text","text":"тест $"}]},{"tag":"button","attr":{"alert":{"cancelbutton":"$cancel$","text":"тест"},"pageparams":{"test":{"text":"test","type":"text"}}},"children":[{"tag":"text","text":"тест"},{"tag":"text","text":" LangRes"}]}]}]`,
		},
		{
			"content",
			url.Values{`template`: {fmt.Sprintf(`Span(Text LangRes(%s)+LangRes(%[1]s,fr))`, name)}},
			`[{"tag":"span","children":[{"tag":"text","text":"Text My test"},{"tag":"text","text":"+French string"}]}]`,
		},
		{
			"content",
			url.Values{
				`template`: {fmt.Sprintf(`Span(Text LangRes(%s)+LangRes(%[1]s,fr))`, name)},
				`lang`:     {`fr`},
			},
			`[{"tag":"span","children":[{"tag":"text","text":"Text French string"},{"tag":"text","text":"+French string"}]}]`,
		},
		{
			"EditLang",
			url.Values{"Name": {name}, "Trans": {`{"en": "My test", "fr": "French string", "es": "Spanish text"}`}},
			"",
		},
		{
			"content",
			url.Values{`template`: {fmt.Sprintf(`Table(mysrc,"$%[1]s$=name")Span(Text LangRes(%[1]s,es) $%[1]s$) Input(Class: form-control, Placeholder: $%[1]s$, Type: text, Name: Name)`, name)}},
			`[{"tag":"table","attr":{"columns":[{"Name":"name","Title":"My test"}],"source":"mysrc"}},{"tag":"span","children":[{"tag":"text","text":"Text Spanish text"},{"tag":"text","text":" My test"}]},{"tag":"input","attr":{"class":"form-control","name":"Name","placeholder":"My test","type":"text"}}]`,
		},
		{
			"content",
			url.Values{`template`: {fmt.Sprintf(`MenuGroup($%s$){MenuItem(Ooops, ooops)}MenuGroup(nolang){MenuItem(no, no)}`, name)}},
			fmt.Sprintf(`[{"tag":"menugroup","attr":{"name":"$%s$","title":"My test"},"children":[{"tag":"menuitem","attr":{"page":"ooops","title":"Ooops"}}]},{"tag":"menugroup","attr":{"name":"nolang","title":"nolang"},"children":[{"tag":"menuitem","attr":{"page":"no","title":"no"}}]}]`, name),
		},
	}

	for _, v := range cases {
		var ret contentResult

		if len(v.expect) == 0 {
			assert.NoError(t, postTx(v.url, &v.form))
			continue
		}

		require.NoError(t, sendPost(v.url, &v.form, &ret))
		assert.Equal(t, v.expect, RawToString(ret.Tree))
	}
}
