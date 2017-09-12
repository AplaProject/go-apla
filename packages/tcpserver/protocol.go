package tcpserver

import (
	"errors"
	"io"
	"reflect"

	"strconv"

	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type TransactionType struct {
	Type uint16
}

// type 10
type MaxBlockRequest struct{}
type MaxBlockResponse struct {
	BlockID uint32
}

// type 7
type GetBodyRequest struct {
	BlockID uint32
}
type GetBodyResponse struct {
	Data []byte
}

// type 4
type ConfirmRequest struct {
	BlockID uint32
}
type ConfirmResponse struct {
	ConfType uint8
	Hash     []byte `size:"32"`
}

// type 2
type DisRequest struct {
	Data []byte
}
type DisTrResponse struct{}

// type 1
type DisHashResponse struct {
	Data []byte
}

func ReadRequest(request interface{}, r io.Reader) error {
	logger.LogDebug(consts.FuncStarted, "")
	if reflect.ValueOf(request).Elem().Kind() != reflect.Struct {
		logger.LogError(consts.TCPCserverError, "bad request type")
		panic("bad request type")
	}
	for i := 0; i < reflect.ValueOf(request).Elem().NumField(); i++ {
		t := reflect.ValueOf(request).Elem().Field(i)
		switch t.Kind() {
		case reflect.Slice:
			size := uint64(0)
			var err error
			sizeVal := reflect.TypeOf(request).Elem().Field(i).Tag.Get("size")
			if sizeVal != "" {
				size, err = strconv.ParseUint(sizeVal, 10, 0)
			} else {
				size, err = readUint(r, 4) // read size
			}
			if err != nil {
				logger.LogError(consts.IncompatibleTypesError, err)
				return err
			}
			value, err := readBytes(r, size)
			if err != nil {
				logger.LogError(consts.IOError, err)
				return err
			}
			t.Set(reflect.ValueOf(value))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val, err := readUint(r, int(t.Type().Size()))
			if err != nil {
				logger.LogError(consts.IOError, err)
				return err
			}
			t.SetUint(val)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, err := readUint(r, int(t.Type().Size()))
			if err != nil {
				logger.LogError(consts.IOError, err)
				return err
			}
			t.SetInt(int64(val))

		default:
			logger.LogError(consts.TCPCserverError, "unsupported field")
			panic("unsupported field")
		}
	}
	return nil
}

func SendRequest(request interface{}, w io.Writer) error {
	logger.LogDebug(consts.FuncStarted, "")
	if reflect.ValueOf(request).Elem().Kind() != reflect.Struct {
		logger.LogError(consts.TCPCserverError, "bad requset type")
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
					logger.LogError(consts.TCPCserverError, "bad size tag")
					panic("bad size tag")
				}
				if size != len(value) {
					logger.LogError(consts.TCPCserverError, fmt.Sprintf("bug, bad slice len, want: %d, got %d", size, len(value)))
					return fmt.Errorf("bug, bad slice len, want: %d, got %d", size, len(value))
				}
			} else {
				_, err := w.Write(converter.DecToBin(len(value), 4))
				if err != nil {
					logger.LogError(consts.IOError, err)
					return err
				}
			}
			_, err := w.Write(value)
			if err != nil {
				logger.LogError(consts.IOError, err)
				return err
			}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			_, err := w.Write(converter.DecToBin(t.Uint(), int64(t.Type().Size())))
			if err != nil {
				logger.LogError(consts.IOError, err)
				return err
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			_, err := w.Write(converter.DecToBin(t.Int(), int64(t.Type().Size())))
			if err != nil {
				logger.LogError(consts.IOError, err)
				return err
			}
		}
	}
	return nil
}

func readUint(r io.Reader, byteCount int) (uint64, error) {
	logger.LogDebug(consts.DebugMessage, "")
	buf, err := readBytes(r, uint64(byteCount))
	if err != nil {
		logger.LogError(consts.IOError, err)
		return 0, utils.ErrInfo(err)
	}
	return uint64(converter.BinToDec(buf)), nil
}

func readBytes(r io.Reader, size uint64) ([]byte, error) {
	logger.LogDebug(consts.DebugMessage, "")
	if size > 10485760 { // TODO
		logger.LogError(consts.TCPCserverError, "bad server")
		return nil, errors.New("bad size")
	}
	value := make([]byte, int(size))
	_, err := io.ReadFull(r, value)
	if err != nil {
		logger.LogError(consts.IOError, err)
	}
	return value, err
}
