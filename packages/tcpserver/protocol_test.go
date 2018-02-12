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
