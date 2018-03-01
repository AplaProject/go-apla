// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
	value := `{"en": "My test", "fr": "French string", "en-US": "US locale" }`

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
	form = url.Values{"Name": {name}, "Value": {fmt.Sprintf(`Span($%s$)`, name)},
		"Menu": {`default_menu`}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	var ret contentResult
	err = sendPost(`content/page/`+name, &url.Values{`lang`: {`fr`}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if RawToString(ret.Tree) != `[{"tag":"span","children":[{"tag":"text","text":"French string"}]}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}
	err = sendPost(`content/page/`+name, &url.Values{`lang`: {`en-GB`}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if RawToString(ret.Tree) != `[{"tag":"span","children":[{"tag":"text","text":"My test"}]}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}
	err = sendPost(`content/page/`+name, &url.Values{`lang`: {`en-US`}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if RawToString(ret.Tree) != `[{"tag":"span","children":[{"tag":"text","text":"US locale"}]}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}
	input := fmt.Sprintf(`
		Div(){
			Button(Body: $%[1]s$ $,  Page:test).Alert(Text: $%[1]s$, ConfirmButton: $confirm$, CancelButton: $cancel$)
			Button(Body: LangRes(%[1]s) LangRes, PageParams: "test", ).Alert(Text: $%[1]s$, CancelButton: $cancel$)
	}`, forutf)
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
