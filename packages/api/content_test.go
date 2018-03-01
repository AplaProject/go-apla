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

func TestContent(t *testing.T) {
	var ret contentResult

	err := sendPost("content", &url.Values{
		"template": {"input Div(myclass, #mytest# Div(mypar) the Div)"},
		"mytest":   {"test value"},
	}, &ret)
	if err != nil {
		t.Error(err)
		return
	}

	if string(ret.Tree) != `[{"tag":"text","text":"input "},{"tag":"div","attr":{"class":"myclass"},"children":[{"tag":"text","text":"test value "},{"tag":"div","attr":{"class":"mypar"}},{"tag":"text","text":" the Div"}]}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, "ret.Tree"))
		return
	}
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`page`)
	form := url.Values{"Name": {name}, "Value": {`If(true){Div(){Span(My text)Address()}}.Else{Div(Body: Hidden text)}`},
		"Menu": {`default_menu`}, "Conditions": {"true"}}
	err = postTx(`NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	err = sendPost("content/source/"+name, &url.Values{}, &ret)
	if err != nil {
		t.Error(err)
		return
	}

	if RawToString(ret.Tree) != `[{"tag":"if","attr":{"condition":"true"},"children":[{"tag":"div","children":[{"tag":"span","children":[{"tag":"text","text":"My text"}]},{"tag":"address"}]}],"tail":[{"tag":"else","children":[{"tag":"div","children":[{"tag":"text","text":"Hidden text"}]}]}]}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, RawToString(ret.Tree)))
		return
	}

}
