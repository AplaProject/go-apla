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

type FirstBlock struct {
	Type          byte
	Time          uint32
	WalletId      int64
	CitizenId     int64
	PublicKey     []byte
	NodePublicKey []byte
	Host          string
}

var blockStructs = make(map[string]reflect.Type)

func init() {
	blockStructs[reflect.TypeOf(FirstBlock{}).Name()] = reflect.TypeOf(FirstBlock{})
}

func MakeStruct(name string) interface{} {
    v := reflect.New(blockStructs[name])//.Elem()
    return v.Interface()
}