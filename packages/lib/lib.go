package lib

import (
	"bytes"
	"fmt"
	"encoding/hex"
	"crypto/ecdsa"
 	"crypto/elliptic"
	"crypto/sha256"
	crand "crypto/rand"
	"encoding/binary"
	b58 "github.com/jbenet/go-base58"
	"golang.org/x/crypto/ripemd160"
)

// Converts binary address to DayLight address.
func BytesToAddress(address []byte) string {
	return `D` + b58.Encode(address)
}

// DecodeLenInt64 gets int64 from []byte and shift the slice. The []byte should  be
// encoded with EncodeLengthPlusInt64.
func DecodeLenInt64(data *[]byte) (int64,error) {
	length := int((*data)[0]) + 1
	if len(*data) < length {
		return 0, fmt.Errorf(`length of data %d < %d`, len(*data), length)
	}
    buf := make([]byte, 8)
	copy(buf, (*data)[1:length])
	x := int64(binary.LittleEndian.Uint64(buf))
	*data = (*data)[length:]
	return x, nil
}

// Encode values into binary data. The format parameter can contains the following characters:
// 1 - 1 byte for encoding byte, int8, uint8
// 4 - 4 bytes for encoding int32, uint32
// i - 2-9 bytes for encoding int64, uint64 by EncodeLenInt64 function
// s - for encoding string or []byte by EncodeLenByte function
/*func EncodeBinary(out *[]byte, format string, args ...interface{}) error {
	if out == nil {
		*out = make([]byte, 0, 2048)
	}
	if len(format) != len(args) {
		return fmt.Errorf(`wrong count of parameters %d != %d`, len(format), len(args))
	}
	for i, ch := range format {
		
	}
}
*/

// Encodes int64 number to []byte. If it is less than 128 then it returns []byte{length}.
// Otherwise, it returns (0x80 | len of int64) + int64 as BigEndian []byte 
//
//   67 => 0x43
//   1024 => 0x820400
//   1000000 => 0x830f4240
//
func EncodeLength(length int64) []byte {
	if length > 0 && length <= 127 {
		return []byte{byte(length)}
	}
	buf := make([]byte, 9)
	binary.BigEndian.PutUint64(buf[1:], uint64(length))
	i := 1
	for ; buf[i] == 0; i++ {
	}
	buf[0] = 0x80 | byte(9-i)
	return append(buf[:1], buf[i:]...)
}

// Decodes []byte to int64 and shifts buf. Bytes must be encoded with EncodeLength function.
//
//   0x43 => 67
//   0x820400 => 1024
//   0x830f4240 => 1000000
//
func DecodeLength(buf *[]byte) (ret int64, err error) {
	if len(*buf) == 0 {
		return 
	}
	length := (*buf)[0]
	if (length & 0x80) != 0 {
		length &= 0x7F
		if len(*buf) < int(length + 1) {
			return 0, fmt.Errorf(`input slice has small size`)
		}
		ret = int64(binary.BigEndian.Uint64(append(make([]byte,8-length), (*buf)[1:length+1]...)))
	} else {
		ret = int64(length)
		length = 0
	}
	*buf = (*buf)[length + 1:]
	return 
}

// Appends the length of the slice + the buf slice. 
func EncodeLenByte(out *[]byte, buf []byte) *[]byte {
	*out = append(append(*out, EncodeLength(int64(len(buf)))...), buf...)
	return out
}

// EncodeLenInt64 appends int64 to []byte as uint8 + little-endian order of uint8.
//
//  65000 => [0x02, 0xe8, 0xfd]
//
func EncodeLenInt64(data *[]byte, x int64) *[]byte {
	var length int
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(x))
	for length = 8;length > 0 && buf[length-1] == 0; length-- {
	}
	*data = append( append( *data, byte(length)), buf[:length]...)
	return data
}

// Fill the slice by zero at left if the size of the slice is less than 32.
func FillLeft(slice []byte) []byte {
	if len(slice) >= 32 {
		return slice
	}
	return append( make([]byte, 32 - len(slice)), slice...)
}

// Function generate a random pair of ECDSA private and public keys.
func GenKeys() (privKey string, pubKey string) {
    private,_  := ecdsa.GenerateKey(elliptic.P256(), crand.Reader) 
	privKey = hex.EncodeToString( private.D.Bytes())
	pubKey = hex.EncodeToString(append( FillLeft(private.PublicKey.X.Bytes()), FillLeft(private.PublicKey.Y.Bytes())...))
	return
}

// Function IsValidAddress checks if the specified address is DayLight address.
func IsValidAddress(address string) bool {
	if address[0] != 'D' { 
		return false
	}
	key := b58.Decode(address[1:])
	checksum := key[len(key)-4:]
	finger := key[:len(key)-4]
	h256 := sha256.Sum256(finger)
	h256 = sha256.Sum256(h256[:])
	return bytes.Compare(checksum, h256[:4]) == 0
}

// Converts a public key to DayLight address.
func KeyToAddress(pubKey []byte) string {
    h256 := sha256.Sum256(pubKey)
    h := ripemd160.New()
    h.Write(h256[:])
    finger := h.Sum(nil)
	h256 = sha256.Sum256(finger)
	h256 = sha256.Sum256(h256[:])
	checksum := h256[:4]
	return BytesToAddress(append(finger, checksum...))
}

