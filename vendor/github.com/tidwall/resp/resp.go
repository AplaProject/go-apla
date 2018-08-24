package resp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const bufsz = 4096

// Type represents a Value type
type Type byte

const (
	SimpleString Type = '+'
	Error        Type = '-'
	Integer      Type = ':'
	BulkString   Type = '$'
	Array        Type = '*'
)

// TypeName returns name of the underlying RESP type.
func (t Type) String() string {
	switch t {
	default:
		return "Unknown"
	case '+':
		return "SimpleString"
	case '-':
		return "Error"
	case ':':
		return "Integer"
	case '$':
		return "BulkString"
	case '*':
		return "Array"
	}
}

// Value represents the data of a valid RESP type.
type Value struct {
	typ     Type
	integer int
	str     []byte
	array   []Value
	null    bool
}

// Integer converts Value to an int. If Value cannot be converted, Zero is returned.
func (v Value) Integer() int {
	switch v.typ {
	default:
		n, _ := strconv.ParseInt(v.String(), 10, 64)
		return int(n)
	case ':':
		return v.integer
	}
}

// String converts Value to a string.
func (v Value) String() string {
	if v.typ == '$' {
		return string(v.str)
	}
	switch v.typ {
	case '+', '-':
		return string(v.str)
	case ':':
		return strconv.FormatInt(int64(v.integer), 10)
	case '*':
		return fmt.Sprintf("%v", v.array)
	}
	return ""
}

// Bytes converts the Value to a byte array. An empty string is converted to a non-nil empty byte array. If it's a RESP Null value, nil is returned.
func (v Value) Bytes() []byte {
	switch v.typ {
	default:
		return []byte(v.String())
	case '$', '+', '-':
		return v.str
	}
}

// Float converts Value to a float64. If Value cannot be converted, Zero is returned.
func (v Value) Float() float64 {
	switch v.typ {
	default:
		f, _ := strconv.ParseFloat(v.String(), 64)
		return f
	case ':':
		return float64(v.integer)
	}
}

// IsNull indicates whether or not the base value is null.
func (v Value) IsNull() bool {
	return v.null
}

// Bool converts Value to an bool. If Value cannot be converted, false is returned.
func (v Value) Bool() bool {
	return v.Integer() != 0
}

// Error converts the Value to an error. If Value is not an error, nil is returned.
func (v Value) Error() error {
	switch v.typ {
	case '-':
		return errors.New(string(v.str))
	}
	return nil
}

// Array converts the Value to a an array. If Value is not an array or when it's is a RESP Null value, nil is returned.
func (v Value) Array() []Value {
	if v.typ == '*' && !v.null {
		return v.array
	}
	return nil
}

// Type returns the underlying RESP type. The following types are represent valid RESP values.
//   '+'  SimpleString
//   '-'  Error
//   ':'  Integer
//   '$'  BulkString
//   '*'  Array
func (v Value) Type() Type {
	return v.typ
}

func marshalSimpleRESP(typ Type, b []byte) ([]byte, error) {
	bb := make([]byte, 3+len(b))
	bb[0] = byte(typ)
	copy(bb[1:], b)
	bb[1+len(b)+0] = '\r'
	bb[1+len(b)+1] = '\n'
	return bb, nil
}

func marshalBulkRESP(v Value) ([]byte, error) {
	if v.null {
		return []byte("$-1\r\n"), nil
	}
	szb := []byte(strconv.FormatInt(int64(len(v.str)), 10))
	bb := make([]byte, 5+len(szb)+len(v.str))
	bb[0] = '$'
	copy(bb[1:], szb)
	bb[1+len(szb)+0] = '\r'
	bb[1+len(szb)+1] = '\n'
	copy(bb[1+len(szb)+2:], v.str)
	bb[1+len(szb)+2+len(v.str)+0] = '\r'
	bb[1+len(szb)+2+len(v.str)+1] = '\n'
	return bb, nil
}

func marshalArrayRESP(v Value) ([]byte, error) {
	if v.null {
		return []byte("*-1\r\n"), nil
	}
	szb := []byte(strconv.FormatInt(int64(len(v.array)), 10))

	var buf bytes.Buffer
	buf.Grow(3 + len(szb) + 16*len(v.array)) // prime the buffer
	buf.WriteByte('*')
	buf.Write(szb)
	buf.WriteByte('\r')
	buf.WriteByte('\n')
	for i := 0; i < len(v.array); i++ {
		data, err := v.array[i].MarshalRESP()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}
	return buf.Bytes(), nil
}

func marshalAnyRESP(v Value) ([]byte, error) {
	switch v.typ {
	default:
		if v.typ == 0 && v.null {
			return []byte("$-1\r\n"), nil
		}
		return nil, errors.New("unknown resp type encountered")
	case '-', '+':
		return marshalSimpleRESP(v.typ, v.str)
	case ':':
		return marshalSimpleRESP(v.typ, []byte(strconv.FormatInt(int64(v.integer), 10)))
	case '$':
		return marshalBulkRESP(v)
	case '*':
		return marshalArrayRESP(v)
	}
}

// Equals compares one value to another value.
func (v Value) Equals(value Value) bool {
	data1, err := v.MarshalRESP()
	if err != nil {
		return false
	}
	data2, err := value.MarshalRESP()
	if err != nil {
		return false
	}
	return string(data1) == string(data2)
}

// MarshalRESP returns the original serialized byte representation of Value.
// For more information on this format please see http://redis.io/topics/protocol.
func (v Value) MarshalRESP() ([]byte, error) {
	return marshalAnyRESP(v)
}

var nullValue = Value{null: true}

type errProtocol struct{ msg string }

func (err errProtocol) Error() string {
	return "Protocol error: " + err.msg
}

// Reader is a specialized RESP Value type reader.
type Reader struct {
	rd *bufio.Reader
}

// NewReader returns a Reader for reading Value types.
func NewReader(rd io.Reader) *Reader {
	r := &Reader{rd: bufio.NewReader(rd)}
	return r
}

// ReadValue reads the next Value from Reader.
func (rd *Reader) ReadValue() (value Value, n int, err error) {
	value, _, n, err = rd.readValue(false, false)
	return
}

// ReadMultiBulk reads the next multi bulk Value from Reader.
// A multi bulk value is a RESP array that contains one or more bulk strings.
// For more information on RESP arrays and strings please see http://redis.io/topics/protocol.
func (rd *Reader) ReadMultiBulk() (value Value, telnet bool, n int, err error) {
	return rd.readValue(true, false)
}

func (rd *Reader) readValue(multibulk, child bool) (val Value, telnet bool, n int, err error) {
	var rn int
	var c byte
	c, err = rd.rd.ReadByte()
	if err != nil {
		return nullValue, false, n, err
	}
	n++
	if c == '*' {
		val, rn, err = rd.readArrayValue(multibulk)
	} else if multibulk && !child {
		telnet = true
	} else {
		switch c {
		default:
			if multibulk && child {
				return nullValue, telnet, n, &errProtocol{"expected '$', got '" + string(c) + "'"}
			}
			if child {
				return nullValue, telnet, n, &errProtocol{"unknown first byte"}
			}
			telnet = true
		case '-', '+':
			val, rn, err = rd.readSimpleValue(c)
		case ':':
			val, rn, err = rd.readIntegerValue()
		case '$':
			val, rn, err = rd.readBulkValue()
		}
	}
	if telnet {
		n--
		rd.rd.UnreadByte()
		val, rn, err = rd.readTelnetMultiBulk()
		if err == nil {
			telnet = true
		}
	}
	n += rn
	if err == io.EOF {
		return nullValue, telnet, n, io.ErrUnexpectedEOF
	}
	return val, telnet, n, err
}

func (rd *Reader) readTelnetMultiBulk() (v Value, n int, err error) {
	values := make([]Value, 0, 8)
	var c byte
	var bline []byte
	var quote, mustspace bool
	for {
		c, err = rd.rd.ReadByte()
		if err != nil {
			return nullValue, n, err
		}
		n += 1
		if c == '\n' {
			if len(bline) > 0 && bline[len(bline)-1] == '\r' {
				bline = bline[:len(bline)-1]
			}
			break
		}
		if mustspace && c != ' ' {
			return nullValue, n, &errProtocol{"unbalanced quotes in request"}
		}
		if c == ' ' {
			if quote {
				bline = append(bline, c)
			} else {
				values = append(values, Value{typ: '$', str: bline})
				bline = nil
			}
		} else if c == '"' {
			if quote {
				mustspace = true
			} else {
				if len(bline) > 0 {
					return nullValue, n, &errProtocol{"unbalanced quotes in request"}
				}
				quote = true
			}
		} else {
			bline = append(bline, c)
		}
	}
	if quote {
		return nullValue, n, &errProtocol{"unbalanced quotes in request"}
	}
	if len(bline) > 0 {
		values = append(values, Value{typ: '$', str: bline})
	}
	return Value{typ: '*', array: values}, n, nil
}

func (rd *Reader) readSimpleValue(typ byte) (val Value, n int, err error) {
	var line []byte
	line, n, err = rd.readLine()
	if err != nil {
		return nullValue, n, err
	}
	return Value{typ: Type(typ), str: line}, n, nil
}
func (rd *Reader) readLine() (line []byte, n int, err error) {
	for {
		b, err := rd.rd.ReadBytes('\n')
		if err != nil {
			return nil, 0, err
		}
		n += len(b)
		line = append(line, b...)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}
func (rd *Reader) readBulkValue() (val Value, n int, err error) {
	var rn int
	var l int
	l, rn, err = rd.readInt()
	n += rn
	if err != nil {
		if _, ok := err.(*errProtocol); ok {
			return nullValue, n, &errProtocol{"invalid bulk length"}
		}
		return nullValue, n, err
	}
	if l < 0 {
		return Value{typ: '$', null: true}, n, nil
	}
	if l > 512*1024*1024 {
		return nullValue, n, &errProtocol{"invalid bulk length"}
	}
	b := make([]byte, l+2)
	rn, err = io.ReadFull(rd.rd, b)
	n += rn
	if err != nil {
		return nullValue, n, err
	}
	if b[l] != '\r' || b[l+1] != '\n' {
		return nullValue, n, &errProtocol{"invalid bulk line ending"}
	}
	return Value{typ: '$', str: b[:l]}, n, nil
}

func (rd *Reader) readArrayValue(multibulk bool) (val Value, n int, err error) {
	var rn int
	var l int
	l, rn, err = rd.readInt()
	n += rn
	if err != nil || l > 1024*1024 {
		if _, ok := err.(*errProtocol); ok {
			if multibulk {
				return nullValue, n, &errProtocol{"invalid multibulk length"}
			}
			return nullValue, n, &errProtocol{"invalid array length"}
		}
		return nullValue, n, err
	}
	if l < 0 {
		return Value{typ: '*', null: true}, n, nil
	}
	var aval Value
	vals := make([]Value, l)
	for i := 0; i < l; i++ {
		aval, _, rn, err = rd.readValue(multibulk, true)
		n += rn
		if err != nil {
			return nullValue, n, err
		}
		vals[i] = aval
	}
	return Value{typ: '*', array: vals}, n, nil
}

func (rd *Reader) readIntegerValue() (val Value, n int, err error) {
	var l int
	l, n, err = rd.readInt()
	if err != nil {
		if _, ok := err.(*errProtocol); ok {
			return nullValue, n, &errProtocol{"invalid integer"}
		}
		return nullValue, n, err
	}
	return Value{typ: ':', integer: l}, n, nil
}

func (rd *Reader) readInt() (x int, n int, err error) {
	line, n, err := rd.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

// AnyValue returns a RESP value from an interface. This function infers the types. Arrays are not allowed.
func AnyValue(v interface{}) Value {
	switch v := v.(type) {
	default:
		return StringValue(fmt.Sprintf("%v", v))
	case nil:
		return NullValue()
	case int:
		return IntegerValue(int(v))
	case uint:
		return IntegerValue(int(v))
	case int8:
		return IntegerValue(int(v))
	case uint8:
		return IntegerValue(int(v))
	case int16:
		return IntegerValue(int(v))
	case uint16:
		return IntegerValue(int(v))
	case int32:
		return IntegerValue(int(v))
	case uint32:
		return IntegerValue(int(v))
	case int64:
		return IntegerValue(int(v))
	case uint64:
		return IntegerValue(int(v))
	case bool:
		return BoolValue(v)
	case float32:
		return FloatValue(float64(v))
	case float64:
		return FloatValue(float64(v))
	case []byte:
		return BytesValue(v)
	case string:
		return StringValue(v)
	}
}

// SimpleStringValue returns a RESP simple string. A simple string has no new lines. The carriage return and new line characters are replaced with spaces.
func SimpleStringValue(s string) Value { return Value{typ: '+', str: []byte(formSingleLine(s))} }

// BytesValue returns a RESP bulk string. A bulk string can represent any data.
func BytesValue(b []byte) Value { return Value{typ: '$', str: b} }

// StringValue returns a RESP bulk string. A bulk string can represent any data.
func StringValue(s string) Value { return Value{typ: '$', str: []byte(s)} }

// NullValue returns a RESP null bulk string.
func NullValue() Value { return Value{typ: '$', null: true} }

// ErrorValue returns a RESP error.
func ErrorValue(err error) Value {
	if err == nil {
		return Value{typ: '-'}
	}
	return Value{typ: '-', str: []byte(err.Error())}
}

// IntegerValue returns a RESP integer.
func IntegerValue(i int) Value { return Value{typ: ':', integer: i} }

// BoolValue returns a RESP integer representation of a bool.
func BoolValue(t bool) Value {
	if t {
		return Value{typ: ':', integer: 1}
	}
	return Value{typ: ':', integer: 0}
}

// FloatValue returns a RESP bulk string representation of a float.
func FloatValue(f float64) Value { return StringValue(strconv.FormatFloat(f, 'f', -1, 64)) }

// ArrayValue returns a RESP array.
func ArrayValue(vals []Value) Value { return Value{typ: '*', array: vals} }

func formSingleLine(s string) string {
	bs1 := []byte(s)
	for i := 0; i < len(bs1); i++ {
		switch bs1[i] {
		case '\r', '\n':
			bs2 := make([]byte, len(bs1))
			copy(bs2, bs1)
			bs2[i] = ' '
			i++
			for ; i < len(bs2); i++ {
				switch bs1[i] {
				case '\r', '\n':
					bs2[i] = ' '
				}
			}
			return string(bs2)
		}
	}
	return s
}

// MultiBulkValue returns a RESP array which contains one or more bulk strings.
// For more information on RESP arrays and strings please see http://redis.io/topics/protocol.
func MultiBulkValue(commandName string, args ...interface{}) Value {
	vals := make([]Value, len(args)+1)
	vals[0] = StringValue(commandName)
	for i, arg := range args {
		if rval, ok := arg.(Value); ok && rval.Type() == BulkString {
			vals[i+1] = rval
			continue
		}
		switch arg := arg.(type) {
		default:
			vals[i+1] = StringValue(fmt.Sprintf("%v", arg))
		case []byte:
			vals[i+1] = StringValue(string(arg))
		case string:
			vals[i+1] = StringValue(arg)
		case nil:
			vals[i+1] = NullValue()
		}
	}
	return ArrayValue(vals)
}

// Writer is a specialized RESP Value type writer.
type Writer struct {
	wr io.Writer
}

// NewWriter returns a new Writer.
func NewWriter(wr io.Writer) *Writer {
	return &Writer{wr}
}

// WriteValue writes a RESP Value.
func (wr *Writer) WriteValue(v Value) error {
	b, err := v.MarshalRESP()
	if err != nil {
		return err
	}
	_, err = wr.wr.Write(b)
	return nil
}

// WriteSimpleString writes a RESP simple string. A simple string has no new lines. The carriage return and new line characters are replaced with spaces.
func (wr *Writer) WriteSimpleString(s string) error { return wr.WriteValue(SimpleStringValue(s)) }

// WriteBytes writes a RESP bulk string. A bulk string can represent any data.
func (wr *Writer) WriteBytes(b []byte) error { return wr.WriteValue(BytesValue(b)) }

// WriteString writes a RESP bulk string. A bulk string can represent any data.
func (wr *Writer) WriteString(s string) error { return wr.WriteValue(StringValue(s)) }

// WriteNull writes a RESP null bulk string.
func (wr *Writer) WriteNull() error { return wr.WriteValue(NullValue()) }

// WriteError writes a RESP error.
func (wr *Writer) WriteError(err error) error { return wr.WriteValue(ErrorValue(err)) }

// WriteInteger writes a RESP integer.
func (wr *Writer) WriteInteger(i int) error { return wr.WriteValue(IntegerValue(i)) }

// WriteArray writes a RESP array.
func (wr *Writer) WriteArray(vals []Value) error { return wr.WriteValue(ArrayValue(vals)) }

// WriteMultiBulk writes a RESP array which contains one or more bulk strings.
// For more information on RESP arrays and strings please see http://redis.io/topics/protocol.
func (wr *Writer) WriteMultiBulk(commandName string, args ...interface{}) error {
	return wr.WriteValue(MultiBulkValue(commandName, args...))
}
