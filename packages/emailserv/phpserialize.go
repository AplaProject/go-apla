package main

import (
	"bytes"
	"fmt"
	"strconv"
)

/* Thanks to https://github.com/wulijun/go-php-serialize */

const ( TYPE_VALUE_SEPARATOR = ':'
		VALUES_SEPARATOR = ';'
)

func Encode(value interface{}) (result []byte, err error) {
	buf := new(bytes.Buffer)
	err = encodeValue(buf, value)
	if err == nil {
		result = buf.Bytes()
	}
	return
}

func encodeValue(buf *bytes.Buffer, value interface{}) (err error) {
	switch t := value.(type) {
	default:
		err = fmt.Errorf("Unexpected type %T", t)
	case nil:
		buf.WriteString("N")
		buf.WriteRune(VALUES_SEPARATOR)
	case int, int64, int32, int16, int8:
		buf.WriteString("i")
		buf.WriteRune(TYPE_VALUE_SEPARATOR)
		strValue := fmt.Sprintf("%v", t)
		buf.WriteString(strValue)
		buf.WriteRune(VALUES_SEPARATOR)		
	case []byte:
		buf.WriteString("s")
		buf.WriteRune(TYPE_VALUE_SEPARATOR)
		encodeByte(buf, t)
		buf.WriteRune(VALUES_SEPARATOR)
	case string:
		buf.WriteString("s")
		buf.WriteRune(TYPE_VALUE_SEPARATOR)
		encodeString(buf, t)
		buf.WriteRune(VALUES_SEPARATOR)
	case map[interface{}]interface{}:
		buf.WriteString("a")
		buf.WriteRune(TYPE_VALUE_SEPARATOR)
		err = encodeArrayCore(buf, t)
	}
	return
}

func encodeByte(buf *bytes.Buffer, byteValue []byte) {
	valLen := strconv.Itoa(len(byteValue))
	buf.WriteString(valLen)
	buf.WriteRune(TYPE_VALUE_SEPARATOR)
	buf.WriteRune('"')
	buf.Write(byteValue)
	buf.WriteRune('"')
}

func encodeString(buf *bytes.Buffer, strValue string) {
	valLen := strconv.Itoa(len(strValue))
	buf.WriteString(valLen)
	buf.WriteRune(TYPE_VALUE_SEPARATOR)
	buf.WriteRune('"')
	buf.WriteString(strValue)
	buf.WriteRune('"')
}

func encodeArrayCore(buf *bytes.Buffer, arrValue map[interface{}]interface{}) (err error) {
	valLen := strconv.Itoa(len(arrValue))
	buf.WriteString(valLen)
	buf.WriteRune(TYPE_VALUE_SEPARATOR)

	buf.WriteRune('{')
	for k, v := range arrValue {
		if intKey, _err := strconv.Atoi(fmt.Sprintf("%v", k)); _err == nil {
			if err = encodeValue(buf, intKey); err != nil {
				break
			}
		} else {
			if err = encodeValue(buf, k); err != nil {
				break
			}
		}
		if err = encodeValue(buf, v); err != nil {
			break
		}
	}
	buf.WriteRune('}')
	return err
}
