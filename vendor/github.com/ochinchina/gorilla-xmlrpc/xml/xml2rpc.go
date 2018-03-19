// Copyright 2013 Ivan Danyliuk
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xml

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/rogpeppe/go-charset/charset"
	_ "github.com/rogpeppe/go-charset/data"
)

// Types used for unmarshalling
type response struct {
	Name   xml.Name   `xml:"methodResponse"`
	Params []param    `xml:"params>param"`
	Fault  faultValue `xml:"fault,omitempty"`
}

type param struct {
	Value value `xml:"value"`
}

type value struct {
	Array    []value  `xml:"array>data>value"`
	Struct   []member `xml:"struct>member"`
	String   string   `xml:"string"`
	Int      string   `xml:"int"`
	Int4     string   `xml:"i4"`
	Double   string   `xml:"double"`
	Boolean  string   `xml:"boolean"`
	DateTime string   `xml:"dateTime.iso8601"`
	Base64   string   `xml:"base64"`
	Raw      string   `xml:",innerxml"` // the value can be defualt string
}

type member struct {
	Name  string `xml:"name"`
	Value value  `xml:"value"`
}

func xml2RPC(xmlraw string, rpc interface{}) error {
	// Unmarshal raw XML into the temporal structure
	var ret response
	decoder := xml.NewDecoder(bytes.NewReader([]byte(xmlraw)))
	decoder.CharsetReader = charset.NewReader
	err := decoder.Decode(&ret)
	if err != nil {
		return FaultDecode
	}

	if !ret.Fault.IsEmpty() {
		return getFaultResponse(ret.Fault)
	}

	// Now, convert temporal structure into the
	// passed rpc variable, according to it's structure
	fieldNum := reflect.TypeOf(rpc).Elem().NumField()
	//for i, param := range ret.Params {
	for i := 0; i < fieldNum; i += 1 {
		field := reflect.ValueOf(rpc).Elem().Field(i)
		if len(ret.Params) > i {
			err = value2Field(ret.Params[i].Value, &field)
		} else if reflect.TypeOf(rpc).Elem().Field(i).Tag.Get("default") != "" {
			err = value2Field(createValue(reflect.TypeOf(rpc).Elem().Field(i).Type.Kind(), reflect.TypeOf(rpc).Elem().Field(i).Tag.Get("default")), &field)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func createValue(kind reflect.Kind, val string) value {
	v := value{}
	if kind == reflect.Bool {
		v.Boolean = val
	} else if kind == reflect.Int {
		v.Int = val
	}
	return v
}

// getFaultResponse converts faultValue to Fault.
func getFaultResponse(fault faultValue) Fault {
	var (
		code int
		str  string
	)

	for _, field := range fault.Value.Struct {
		if field.Name == "faultCode" {
			code, _ = strconv.Atoi(field.Value.Int)
		} else if field.Name == "faultString" {
			str = field.Value.String
			if str == "" {
				str = field.Value.Raw
			}
		}
	}

	return Fault{Code: code, String: str}
}

func value2Field(value value, field *reflect.Value) error {
	if !field.CanSet() {
		return FaultApplicationError
	}

	var (
		err error
		val interface{}
	)

	switch {
	case value.Int != "":
		val, _ = strconv.Atoi(value.Int)
	case value.Int4 != "":
		val, _ = strconv.Atoi(value.Int4)
	case value.Double != "":
		val, _ = strconv.ParseFloat(value.Double, 64)
	case value.String != "":
		val = value.String
	case value.Boolean != "":
		val = xml2Bool(value.Boolean)
	case value.DateTime != "":
		val, err = xml2DateTime(value.DateTime)
	case value.Base64 != "":
		val, err = xml2Base64(value.Base64)
	case len(value.Struct) != 0:
		if field.Kind() != reflect.Struct {
			fault := FaultInvalidParams
			fault.String += fmt.Sprintf("structure fields mismatch: %s != %s", field.Kind(), reflect.Struct.String())
			return fault
		}
		s := value.Struct
		for i := 0; i < len(s); i++ {
			// Uppercase first letter for field name to deal with
			// methods in lowercase, which cannot be used
			field_name := uppercaseFirst(s[i].Name)
			f := field.FieldByName(field_name)
			err = value2Field(s[i].Value, &f)
		}
	case len(value.Array) != 0:
		a := value.Array
		f := *field
		slice := reflect.MakeSlice(reflect.TypeOf(f.Interface()),
			len(a), len(a))
		for i := 0; i < len(a); i++ {
			item := slice.Index(i)
			err = value2Field(a[i], &item)
		}
		f = reflect.AppendSlice(f, slice)
		val = f.Interface()

	default:
		// value field is default to string, see http://en.wikipedia.org/wiki/XML-RPC#Data_types
		// also can be <nil/>
		if value.Raw != "<nil/>" {
			val = value.Raw
		}
	}

	if val != nil {
		if reflect.TypeOf(val) != reflect.TypeOf(field.Interface()) {
			fault := FaultInvalidParams
			fault.String += fmt.Sprintf(": fields type mismatch: %s != %s",
				reflect.TypeOf(val),
				reflect.TypeOf(field.Interface()))
			return fault
		}

		field.Set(reflect.ValueOf(val))
	}

	return err
}

func xml2Bool(value string) bool {
	var b bool
	switch value {
	case "1", "true", "TRUE", "True":
		b = true
	case "0", "false", "FALSE", "False":
		b = false
	}
	return b
}

func xml2DateTime(value string) (time.Time, error) {
	var (
		year, month, day     int
		hour, minute, second int
	)
	_, err := fmt.Sscanf(value, "%04d%02d%02dT%02d:%02d:%02d",
		&year, &month, &day,
		&hour, &minute, &second)
	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
	return t, err
}

func xml2Base64(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(value)
}

func uppercaseFirst(in string) (out string) {
	r, n := utf8.DecodeRuneInString(in)
	return string(unicode.ToUpper(r)) + in[n:]
}
