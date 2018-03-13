/*
XML-RPC implementation for the Gorilla/RPC toolkit.

It's built on top of gorilla/rpc package in Go(Golang) language and implements XML-RPC, according to it's specification. Unlike net/rpc from Go strlib, gorilla/rpc allows usage of HTTP POST requests for RPC.

XML-RPC spec: http://xmlrpc.scripting.com/spec.html

Installation

Assuming you already imported gorilla/rpc, use the following command:

    go get github.com/divan/gorilla-xmlrpc/xml

Implementation details

The main objective was to use standard encoding/xml package for XML marshalling/unmarshalling. Unfortunately, in current implementation there is no graceful way to implement common structre for marshal and unmarshal functions - marshalling doesn't handle interface{} types so far (though, it could be changed in the future). So, marshalling is implemented manually.

Unmarshalling code first creates temporary structure for unmarshalling XML into, then converts it into the passed variable using reflect package. If XML struct member's name is lowercased, it's first letter will be uppercased, as in Go/Gorilla field name must be exported(first-letter uppercased).

Marshalling code converts rpc directly to the string XML representation.

For the better understanding, I use terms 'rpc2xml' and 'xml2rpc' instead of 'marshal' and 'unmarshall'.

Types

The following types are supported:

    XML-RPC             Golang
    -------             ------
    int, i4             int
    double              float64
    boolean             bool
    stringi             string
    dateTime.iso8601    time.Time
    base64              []byte
    struct              struct
    array               []interface{}
    nil                 nil

TODO

TODO list:
 * Add more corner cases tests

Examples

Checkout examples in examples/ directory.

*/
package xml
