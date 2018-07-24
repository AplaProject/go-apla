// Copyright 2013 Ivan Danyliuk
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/rpc"
)

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// NewCodec returns a new XML-RPC Codec.
func NewCodec() *Codec {
	return &Codec{
		aliases: make(map[string]string),
	}
}

// Codec creates a CodecRequest to process each request.
type Codec struct {
	aliases map[string]string
}

// RegisterAlias creates a method alias
func (c *Codec) RegisterAlias(alias, method string) {
	c.aliases[alias] = method
}

// NewRequest returns a CodecRequest.
func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest {
	rawxml, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &CodecRequest{err: err}
	}
	defer r.Body.Close()

	var request ServerRequest
	if err := xml.Unmarshal(rawxml, &request); err != nil {
		return &CodecRequest{err: err}
	}
	request.rawxml = string(rawxml)
	if method, ok := c.aliases[request.Method]; ok {
		request.Method = method
	}
	return &CodecRequest{request: &request}
}

// ----------------------------------------------------------------------------
// CodecRequest
// ----------------------------------------------------------------------------

type ServerRequest struct {
	Name   xml.Name `xml:"methodCall"`
	Method string   `xml:"methodName"`
	rawxml string
}

// CodecRequest decodes and encodes a single request.
type CodecRequest struct {
	request *ServerRequest
	err     error
}

// Method returns the RPC method for the current request.
//
// The method uses a dotted notation as in "Service.Method".
func (c *CodecRequest) Method() (string, error) {
	if c.err == nil {
		return c.request.Method, nil
	}
	return "", c.err
}

// ReadRequest fills the request object for the RPC method.
//
// args is the pointer to the Service.Args structure
// it gets populated from temporary XML structure
func (c *CodecRequest) ReadRequest(args interface{}) error {
	c.err = xml2RPC(c.request.rawxml, args)
	return nil
}

// WriteResponse encodes the response and writes it to the ResponseWriter.
//
// response is the pointer to the Service.Response structure
// it gets encoded into the XML-RPC xml string
func (c *CodecRequest) WriteResponse(w http.ResponseWriter, response interface{}, methodErr error) error {
	if c.err == nil {
		c.err = methodErr
	}
	buffer := bytes.NewBuffer(make([]byte, 0))
	if c.err != nil {
		var fault Fault
		switch c.err.(type) {
		case Fault:
			fault = c.err.(Fault)
		default:
			fault = FaultApplicationError
			fault.String += fmt.Sprintf(": %v", c.err)
		}
		fault2XML(fault, buffer)
	} else {
		rpcResponse2XML(response, buffer)
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	buffer.WriteTo(w)
	return nil
}
