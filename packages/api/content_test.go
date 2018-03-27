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
		t.Error(fmt.Errorf(`wrong tree %s`, ret.Tree))
		return
	}
	err = sendPost("content", &url.Values{
		"template":  {"#test_page# input Div(myclass, #test_page# ok) #test_page#"},
		"test_page": {"7"},
	}, &ret)
	if err != nil {
		t.Error(err)
		return
	}

	if string(ret.Tree) != `[{"tag":"text","text":"7 input "},{"tag":"div","attr":{"class":"myclass"},"children":[{"tag":"text","text":"7 ok"}]},{"tag":"text","text":" 7"}]` {
		t.Error(fmt.Errorf(`wrong tree %s`, ret.Tree))
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
