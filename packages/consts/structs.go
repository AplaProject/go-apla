package consts

import (
	"reflect"
)

type BlockHeader struct {
	Type     byte
	BlockId  uint32
	Time     uint32
	WalletId int64
	CBId     byte
	Sign     []byte
}

type TxHeader struct {
	Type          byte
	Time          uint32
	WalletId      int64
	CitizenId     int64
}

type FirstBlock struct {
	TxHeader
	PublicKey     []byte
	NodePublicKey []byte
	Host          string
}

type CitizenRequest struct {
	TxHeader
	StateId       int64
	Sign          []byte
}

var blockStructs = make(map[string]reflect.Type)

func init() {
	list := []interface{}{ FirstBlock{}, CitizenRequest{},
	   // New structures must be inserted here
	}
	for _, item := range list {
		blockStructs[reflect.TypeOf(item).Name()] = reflect.TypeOf(item)
	}
}

func MakeStruct(name string) interface{} {
    v := reflect.New(blockStructs[name])//.Elem()
    return v.Interface()
}

func IsStruct(tx int) bool {
	return tx > 0 && tx <= 2 /*CitizenRequest*/
}

func Header(v interface{}) TxHeader {
	return reflect.ValueOf(v).Elem().Field(0).Interface().(TxHeader)
}

func Sign(v interface{}) (sign []byte) {
	field := reflect.ValueOf(v).Elem().FieldByName(`Sign`)
	if field.IsValid() && field.Kind() == reflect.Slice {
		sign = field.Bytes()
	}
	return
}