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

package tcpserver

import (
	"bytes"

	"testing"

	"reflect"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

func TestReadRequest(t *testing.T) {
	type testStruct struct {
		Id   uint32
		Data []byte
	}

	request := &bytes.Buffer{}
	request.Write(converter.DecToBin(10, 4))
	request.Write(converter.DecToBin(len("test"), 4))
	request.Write([]byte("test"))

	test := &testStruct{}
	err := ReadRequest(test, request)
	if err != nil {
		t.Errorf("read request return err: %s", err)
	}
	if test.Id != 10 {
		t.Errorf("bad id value")
	}
	if string(test.Data) != "test" {
		t.Errorf("bad data value: %+v", string(test.Data))
	}
}

func TestReadRequestTag(t *testing.T) {
	type testStruct2 struct {
		Id   uint32
		Data []byte `size:"4"`
	}

	request := &bytes.Buffer{}
	request.Write(converter.DecToBin(10, 4))
	request.Write([]byte("test"))

	test := &testStruct2{}
	err := ReadRequest(test, request)
	if err != nil {
		t.Errorf("read request return err: %s", err)
	}
	if test.Id != 10 {
		t.Errorf("bad id value")
	}
	if string(test.Data) != "test" {
		t.Errorf("bad data value: %+v", string(test.Data))
	}
}

func TestSendRequest(t *testing.T) {
	type testStruct2 struct {
		Id   uint32
		Id2  int64
		Test []byte
		Text []byte `size:"4"`
	}

	test := testStruct2{
		Id:   15,
		Id2:  0x1BCDEF0010203040,
		Test: []byte("test"),
		Text: []byte("text"),
	}

	bin := bytes.Buffer{}
	err := SendRequest(&test, &bin)
	if err != nil {
		t.Fatalf("send request failed: %s", err)
	}

	test2 := testStruct2{}
	err = ReadRequest(&test2, &bin)
	if err != nil {
		t.Fatalf("read request failed: %s", err)
	}

	if !reflect.DeepEqual(test, test2) {
		t.Errorf("different values: %+v and %+v", test, test2)
	}
}
