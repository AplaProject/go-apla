//MIT License
//
//Copyright (c) 2016-2018 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package tcpserver

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

// TransactionType is type of transaction
type TransactionType struct {
	Type uint16
}

// MaxBlockRequest is max block request
type MaxBlockRequest struct{}

// MaxBlockResponse is max block response
type MaxBlockResponse struct {
	BlockID uint32
}

// GetBodyRequest contains BlockID
type GetBodyRequest struct {
	BlockID uint32
}

// GetBodyResponse is Data []bytes
type GetBodyResponse struct {
	Data []byte
}

// ConfirmRequest contains request data
type ConfirmRequest struct {
	BlockID uint32
}

// ConfirmResponse contains response data
type ConfirmResponse struct {
	ConfType uint8
	Hash     []byte `size:"32"`
}

// DisRequest contains request data
type DisRequest struct {
	Data []byte
}

// DisTrResponse contains response data
type DisTrResponse struct{}

// DisHashResponse contains response data
type DisHashResponse struct {
	Data []byte
}

// ReadRequest is reading request
func ReadRequest(request interface{}, r io.Reader) error {
	if reflect.ValueOf(request).Elem().Kind() != reflect.Struct {
		log.WithFields(log.Fields{"type": consts.ProtocolError}).Error("bad request type")
		panic("bad request type")
	}
	for i := 0; i < reflect.ValueOf(request).Elem().NumField(); i++ {
		t := reflect.ValueOf(request).Elem().Field(i)
		switch t.Kind() {
		case reflect.Slice:
			var size uint64
			var err error
			sizeVal := reflect.TypeOf(request).Elem().Field(i).Tag.Get("size")
			if sizeVal != "" {
				size, err = strconv.ParseUint(sizeVal, 10, 0)
				if err != nil {
					log.WithFields(log.Fields{"value": sizeVal, "type": consts.ConversionError, "error": err}).Error("parsing uint")
				}
			} else {
				size, err = readUint(r, 4) // read size
			}
			if err != nil {
				return err
			}
			value, err := readBytes(r, size)
			if err != nil {
				return err
			}
			t.Set(reflect.ValueOf(value))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val, err := readUint(r, int(t.Type().Size()))
			if err != nil {
				return err
			}
			t.SetUint(val)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, err := readUint(r, int(t.Type().Size()))
			if err != nil {
				return err
			}
			t.SetInt(int64(val))

		default:
			log.WithFields(log.Fields{"type": consts.ProtocolError}).Error("unsupported field")
			panic("unsupported field")
		}
	}
	return nil
}

// SendRequest in sending request
func SendRequest(request interface{}, w io.Writer) error {
	if reflect.ValueOf(request).Elem().Kind() != reflect.Struct {
		log.WithFields(log.Fields{"type": consts.ProtocolError}).Error("bad request type")
		panic("bad request type")
	}
	for i := 0; i < reflect.ValueOf(request).Elem().NumField(); i++ {
		t := reflect.ValueOf(request).Elem().Field(i)
		switch t.Kind() {
		case reflect.Slice:
			value := t.Bytes()

			sizeVal := reflect.TypeOf(request).Elem().Field(i).Tag.Get("size")
			if sizeVal != "" {
				size, err := strconv.Atoi(sizeVal)
				if err != nil {
					log.WithFields(log.Fields{"value": sizeVal, "type": consts.ConversionError, "error": err}).Error("Converting str to int")
					panic("bad size tag")
				}
				if size != len(value) {
					log.WithFields(log.Fields{"size": size, "len": len(value), "type": consts.ProtocolError}).Error("bad slice len")
					return fmt.Errorf("bug, bad slice len, want: %d, got %d", size, len(value))
				}
			} else {
				_, err := w.Write(converter.DecToBin(len(value), 4))
				if err != nil {
					log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing bytes")
					return err
				}
			}
			_, err := w.Write(value)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing bytes")
				return err
			}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			_, err := w.Write(converter.DecToBin(t.Uint(), int64(t.Type().Size())))
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing bytes")
				return err
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			_, err := w.Write(converter.DecToBin(t.Int(), int64(t.Type().Size())))
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing bytes")
				return err
			}
		}
	}
	return nil
}

func readUint(r io.Reader, byteCount int) (uint64, error) {
	buf, err := readBytes(r, uint64(byteCount))
	if err != nil {
		return 0, utils.ErrInfo(err)
	}
	return uint64(converter.BinToDec(buf)), nil
}

func readBytes(r io.Reader, size uint64) ([]byte, error) {
	var maxSize uint64 = 10485760
	if size > maxSize { // TODO
		log.WithFields(log.Fields{"size": size, "max_size": maxSize, "type": consts.ParameterExceeded}).Error("bytes size to read exceeds max allowed size")
		return nil, errors.New("bad size")
	}
	value := make([]byte, int(size))
	_, err := io.ReadFull(r, value)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.IOError}).Error("cannot read bytes")
	}
	return value, err
}
