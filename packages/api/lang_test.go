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
)

func TestLang(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`lng`)
	value := `{"en": "My test", "fr": "French string" }`

	form := url.Values{"Name": {name}, "Trans": {value}}
	err := postTx(`NewLang`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	forutf := randName(`lng`)
	err = postTx(`NewLang`, &url.Values{"Name": {forutf}, "Trans": {`{"en": "тест" }`}})
	if err != nil {
		t.Error(err)
		return
	}

	input := fmt.Sprintf(`
		Div(){
			Button(Body: $%[1]s$ $,  Page:test).Alert(Text: $%[1]s$, ConfirmButton: $confirm$, CancelButton: $cancel$)
			Button(Body: LangRes(%[1]s) LangRes, PageParams: "test", ).Alert(Text: $%[1]s$, CancelButton: $cancel$)
	}`, forutf)
	var ret contentResult
	err = sendPost(`content`, &url.Values{`template`: {input}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if RawToString(ret.Tree) != `[{"tag":"div","children":[{"tag":"button","attr":{"alert":{"cancelbutton":"$cancel$","confirmbutton":"$confirm$","text":"тест"},"page":"test"},"children":[{"tag":"text","text":"тест $"}]},{"tag":"button","attr":{"alert":{"cancelbutton":"$cancel$","text":"тест"},"pageparams":{"test":{"text":"test","type":"text"}}},"children":[{"tag":"text","text":"тест"},{"tag":"text","text":" LangRes"}]}]}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}
	input = fmt.Sprintf(`Span(Text LangRes(%s)+LangRes(%[1]s,fr))`, name)
	err = sendPost(`content`, &url.Values{`template`: {input}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if RawToString(ret.Tree) != `[{"tag":"span","children":[{"tag":"text","text":"Text My test"},{"tag":"text","text":"+French string"}]}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}
	err = sendPost(`content`, &url.Values{`template`: {input}, `lang`: {`fr`}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if RawToString(ret.Tree) != `[{"tag":"span","children":[{"tag":"text","text":"Text French string"},{"tag":"text","text":"+French string"}]}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}

	value = `{"en": "My test", "fr": "French string", "es": "Spanish text"}`

	form = url.Values{"Name": {name}, "Trans": {value}}
	err = postTx(`EditLang`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	input = fmt.Sprintf(`Table(mysrc,"$%[1]s$=name")Span(Text LangRes(%[1]s,es) $%[1]s$) Input(Class: form-control, Placeholder: $%[1]s$, Type: text, Name: Name)`, name)
	err = sendPost(`content`, &url.Values{`template`: {input}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if RawToString(ret.Tree) != `[{"tag":"table","attr":{"columns":[{"Name":"name","Title":"My test"}],"source":"mysrc"}},{"tag":"span","children":[{"tag":"text","text":"Text Spanish text"},{"tag":"text","text":" My test"}]},{"tag":"input","attr":{"class":"form-control","name":"Name","placeholder":"My test","type":"text"}}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}
	input = fmt.Sprintf(`MenuGroup($%s$){MenuItem(Ooops, ooops)}MenuGroup(nolang){MenuItem(no, no)}`, name)
	err = sendPost(`content`, &url.Values{`template`: {input}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if RawToString(ret.Tree) != fmt.Sprintf(`[{"tag":"menugroup","attr":{"name":"$%s$","title":"My test"},"children":[{"tag":"menuitem","attr":{"page":"ooops","title":"Ooops"}}]},{"tag":"menugroup","attr":{"name":"nolang","title":"nolang"},"children":[{"tag":"menuitem","attr":{"page":"no","title":"no"}}]}]`, name) {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}

}
