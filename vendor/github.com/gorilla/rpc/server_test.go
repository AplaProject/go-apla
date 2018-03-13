// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"net/http"
	"strconv"
	"testing"
)

type Service1Request struct {
	A int
	B int
}

type Service1Response struct {
	Result int
}

type Service1 struct {
}

func (t *Service1) Multiply(r *http.Request, req *Service1Request, res *Service1Response) error {
	res.Result = req.A * req.B
	return nil
}

func (t *Service1) Add(req *Service1Request, res *Service1Response) error {
	res.Result = req.A + req.B
	return nil
}

type Service2 struct {
}

func TestRegisterService(t *testing.T) {
	var err error
	s := NewServer()
	service1 := new(Service1)
	service2 := new(Service2)

	// Inferred name.
	err = s.RegisterService(service1, "")
	if err != nil || !s.HasMethod("Service1.Multiply") {
		t.Errorf("Expected to be registered: Service1.Multiply")
	}
	// Provided name.
	err = s.RegisterService(service1, "Foo")
	if err != nil || !s.HasMethod("Foo.Multiply") {
		t.Errorf("Expected to be registered: Foo.Multiply")
	}
	// No methods.
	err = s.RegisterService(service2, "")
	if err == nil {
		t.Errorf("Expected error on service2")
	}
}

func TestRegisterTCPService(t *testing.T) {
	var err error
	s := NewServer()
	service1 := new(Service1)
	service2 := new(Service2)

	// Inferred name.
	err = s.RegisterTCPService(service1, "")
	if err != nil || !s.HasMethod("Service1.Add") {
		t.Errorf("Expected to be registered: Service1.Add")
	}
	// Provided name.
	err = s.RegisterTCPService(service1, "Foo")
	if err != nil || !s.HasMethod("Foo.Add") {
		t.Errorf("Expected to be registered: Foo.Add")
	}
	// No methods.
	err = s.RegisterTCPService(service2, "")
	if err == nil {
		t.Errorf("Expected error on service2")
	}
}

// MockCodec decodes to Service1.Multiply.
type MockCodec struct {
	A, B int
}

func (c MockCodec) NewRequest(*http.Request) CodecRequest {
	return MockCodecRequest{c.A, c.B}
}

type MockCodecRequest struct {
	A, B int
}

func (r MockCodecRequest) Method() (string, error) {
	return "Service1.Multiply", nil
}

func (r MockCodecRequest) ReadRequest(args interface{}) error {
	req := args.(*Service1Request)
	req.A, req.B = r.A, r.B
	return nil
}

func (r MockCodecRequest) WriteResponse(w http.ResponseWriter, reply interface{}, methodErr error) error {
	if methodErr != nil {
		w.Write([]byte(methodErr.Error()))
	} else {
		res := reply.(*Service1Response)
		w.Write([]byte(strconv.Itoa(res.Result)))
	}
	return nil
}

type MockResponseWriter struct {
	header http.Header
	Status int
	Body   string
}

func NewMockResponseWriter() *MockResponseWriter {
	header := make(http.Header)
	return &MockResponseWriter{header: header}
}

func (w *MockResponseWriter) Header() http.Header {
	return w.header
}

func (w *MockResponseWriter) Write(p []byte) (int, error) {
	w.Body = string(p)
	if w.Status == 0 {
		w.Status = 200
	}
	return len(p), nil
}

func (w *MockResponseWriter) WriteHeader(status int) {
	w.Status = status
}

func TestServeHTTP(t *testing.T) {
	const (
		A = 2
		B = 3
	)
	expected := A * B

	s := NewServer()
	s.RegisterService(new(Service1), "")
	s.RegisterCodec(MockCodec{A, B}, "mock")

	r, err := http.NewRequest("POST", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "mock; dummy")
	w := NewMockResponseWriter()
	s.ServeHTTP(w, r)
	if w.Status != 200 {
		t.Errorf("Status was %d, should be 200.", w.Status)
	}
	if w.Body != strconv.Itoa(expected) {
		t.Errorf("Response body was %s, should be %s.", w.Body, strconv.Itoa(expected))
	}

	// Test wrong Content-Type
	r.Header.Set("Content-Type", "invalid")
	w = NewMockResponseWriter()
	s.ServeHTTP(w, r)
	if w.Status != 415 {
		t.Errorf("Status was %d, should be 415.", w.Status)
	}
	if w.Body != "rpc: unrecognized Content-Type: invalid" {
		t.Errorf("Wrong response body.")
	}

	// Test omitted Content-Type; codec should default to the sole registered one.
	r.Header.Del("Content-Type")
	w = NewMockResponseWriter()
	s.ServeHTTP(w, r)
	if w.Status != 200 {
		t.Errorf("Status was %d, should be 200.", w.Status)
	}
	if w.Body != strconv.Itoa(expected) {
		t.Errorf("Response body was %s, should be %s.", w.Body, strconv.Itoa(expected))
	}
}

func TestInterception(t *testing.T) {
	const (
		A = 2
		B = 3
	)
	expected := A * B

	r2, err := http.NewRequest("POST", "mocked/request", nil)
	if err != nil {
		t.Fatal(err)
	}

	s := NewServer()
	s.RegisterService(new(Service1), "")
	s.RegisterCodec(MockCodec{A, B}, "mock")
	s.RegisterInterceptFunc(func(i *RequestInfo) *http.Request {
		return r2
	})
	s.RegisterAfterFunc(func(i *RequestInfo) {
		if i.Request != r2 {
			t.Errorf("Request was %v, should be %v.", i.Request, r2)
		}
	})

	r, err := http.NewRequest("POST", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "mock; dummy")
	w := NewMockResponseWriter()
	s.ServeHTTP(w, r)
	if w.Status != 200 {
		t.Errorf("Status was %d, should be 200.", w.Status)
	}
	if w.Body != strconv.Itoa(expected) {
		t.Errorf("Response body was %s, should be %s.", w.Body, strconv.Itoa(expected))
	}
}
