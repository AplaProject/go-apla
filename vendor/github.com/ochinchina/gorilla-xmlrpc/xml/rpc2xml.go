// Copyright 2013 Ivan Danyliuk
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

func rpcRequest2XML(method string, rpc interface{}) (string, error) {
	buffer := bytes.NewBuffer(make([]byte, 0))
	fmt.Fprintf(buffer, "<methodCall><methodName>%s</methodName>", method)
	err := rpcParams2XML(rpc, buffer)
	fmt.Fprintf(buffer, "</methodCall>")
	return buffer.String(), err
}

func rpcResponse2XMLStr(rpc interface{}) (string, error) {
	buffer := bytes.NewBuffer(make([]byte, 0))
	err := rpcResponse2XML(rpc, buffer)
	return buffer.String(), err
}

func rpcResponse2XML(rpc interface{}, writer io.Writer) error {
	fmt.Fprintf(writer, "<methodResponse>")
	err := rpcParams2XML(rpc, writer)
	fmt.Fprintf(writer, "</methodResponse>")
	return err
}

func rpcParams2XML(rpc interface{}, writer io.Writer) error {
	var err error
	fmt.Fprintf(writer, "<params>")
	for i := 0; i < reflect.ValueOf(rpc).Elem().NumField(); i++ {
		fmt.Fprintf(writer, "<param>")
		err = rpc2XML(reflect.ValueOf(rpc).Elem().Field(i).Interface(), writer)
		fmt.Fprintf(writer, "</param>")
	}
	fmt.Fprintf(writer, "</params>")
	return err
}

func rpc2XML(value interface{}, writer io.Writer) error {
	fmt.Fprintf(writer, "<value>")
	switch reflect.ValueOf(value).Kind() {
	case reflect.Int:
		fmt.Fprintf(writer, "<int>%d</int>", value.(int))
	case reflect.Float64:
		fmt.Fprintf(writer, "<double>%f</double>", value.(float64))
	case reflect.String:
		string2XML(value.(string), writer)
	case reflect.Bool:
		bool2XML(value.(bool), writer)
	case reflect.Struct:
		if reflect.TypeOf(value).String() != "time.Time" {
			struct2XML(value, writer)
		} else {
			time2XML(value.(time.Time), writer)
		}
	case reflect.Slice, reflect.Array:
		// FIXME: is it the best way to recognize '[]byte'?
		if reflect.TypeOf(value).String() != "[]uint8" {
			array2XML(value, writer)
		} else {
			base642XML(value.([]byte), writer)
		}
	case reflect.Ptr:
		if reflect.ValueOf(value).IsNil() {
			fmt.Fprintf(writer, "<nil/>")
		}
	}
	fmt.Fprintf(writer, "</value>")
	return nil
}

func bool2XML(value bool, writer io.Writer) {
	var b string
	if value {
		b = "1"
	} else {
		b = "0"
	}
	fmt.Fprintf(writer, "<boolean>%s</boolean>", b)
}

func string2XML(value string, writer io.Writer) {
	value = strings.Replace(value, "&", "&amp;", -1)
	value = strings.Replace(value, "\"", "&quot;", -1)
	value = strings.Replace(value, "<", "&lt;", -1)
	value = strings.Replace(value, ">", "&gt;", -1)
	fmt.Fprintf(writer, "<string>%s</string>", value)
}

func struct2XML(value interface{}, writer io.Writer) {
	fmt.Fprintf(writer, "<struct>")
	for i := 0; i < reflect.TypeOf(value).NumField(); i++ {
		field := reflect.ValueOf(value).Field(i)
		field_type := reflect.TypeOf(value).Field(i)
		var name string
		if field_type.Tag.Get("xml") != "" {
			name = field_type.Tag.Get("xml")
		} else {
			name = field_type.Name
		}
		fmt.Fprintf(writer, "<member>")
		fmt.Fprintf(writer, "<name>%s</name>", name)
		rpc2XML(field.Interface(), writer)
		fmt.Fprintf(writer, "</member>")
	}
	fmt.Fprintf(writer, "</struct>")
	return
}

func array2XML(value interface{}, writer io.Writer) {
	fmt.Fprintf(writer, "<array><data>")
	for i := 0; i < reflect.ValueOf(value).Len(); i++ {
		rpc2XML(reflect.ValueOf(value).Index(i).Interface(), writer)
	}
	fmt.Fprintf(writer, "</data></array>")
}

func time2XML(t time.Time, writer io.Writer) {
	/*
		// TODO: find out whether we need to deal
		// here with TZ
		var tz string;
		zone, offset := t.Zone()
		if zone == "UTC" {
			tz = "Z"
		} else {
			tz = fmt.Sprintf("%03d00", offset / 3600 )
		}
	*/
	fmt.Fprintf(writer, "<dateTime.iso8601>%04d%02d%02dT%02d:%02d:%02d</dateTime.iso8601>",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

func base642XML(data []byte, writer io.Writer) {
	str := base64.StdEncoding.EncodeToString(data)
	fmt.Fprintf(writer, "<base64>%s</base64>", str)
}
