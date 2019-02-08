// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package converter

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"bytes"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var FirstEcosystemTables = map[string]bool{
	`keys`:               false,
	`menu`:               true,
	`pages`:              true,
	`blocks`:             true,
	`languages`:          true,
	`contracts`:          true,
	`tables`:             true,
	`parameters`:         true,
	`history`:            true,
	`sections`:           true,
	`members`:            false,
	`roles`:              true,
	`roles_participants`: true,
	`notifications`:      true,
	`applications`:       true,
	`binaries`:           true,
	`buffer_data`:        true,
	`app_params`:         true,
}

// FillLeft is filling slice
func FillLeft(slice []byte) []byte {
	if len(slice) >= 32 {
		return slice
	}
	return append(make([]byte, 32-len(slice)), slice...)
}

func EncodeLenInt64(data *[]byte, x int64) *[]byte {
	var length int
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(x))
	for length = 8; length > 0 && buf[length-1] == 0; length-- {
	}
	*data = append(append(*data, byte(length)), buf[:length]...)
	return data
}

func EncodeLenInt64InPlace(x int64) []byte {
	buf := make([]byte, 9)
	value := buf[1:]
	binary.LittleEndian.PutUint64(value, uint64(x))
	var length byte
	for length = 8; length > 0 && value[length-1] == 0; length-- {
	}
	buf[0] = length
	return buf[:length+1]
}

func EncodeLenByte(out *[]byte, buf []byte) *[]byte {
	*out = append(append(*out, EncodeLength(int64(len(buf)))...), buf...)
	return out
}

// EncodeLength encodes int64 number to []byte. If it is less than 128 then it returns []byte{length}.
// Otherwise, it returns (0x80 | len of int64) + int64 as BigEndian []byte
//
//   67 => 0x43
//   1024 => 0x820400
//   1000000 => 0x830f4240
//
func EncodeLength(length int64) []byte {
	if length >= 0 && length <= 127 {
		return []byte{byte(length)}
	}
	buf := make([]byte, 9)
	binary.BigEndian.PutUint64(buf[1:], uint64(length))
	i := 1
	for ; buf[i] == 0 && i < 8; i++ {
	}
	buf[0] = 0x80 | byte(9-i)
	return append(buf[:1], buf[i:]...)
}

// DecodeLenInt64 gets int64 from []byte and shift the slice. The []byte should  be
// encoded with EncodeLengthPlusInt64.
func DecodeLenInt64(data *[]byte) (int64, error) {
	if len(*data) == 0 {
		return 0, nil
	}
	length := int((*data)[0]) + 1
	if len(*data) < length {
		log.WithFields(log.Fields{"data_length": len(*data), "length": length, "type": consts.UnmarshallingError}).Error("length of data is smaller then encoded length")
		return 0, fmt.Errorf(`length of data %d < %d`, len(*data), length)
	}
	buf := make([]byte, 8)
	copy(buf, (*data)[1:length])
	x := int64(binary.LittleEndian.Uint64(buf))
	*data = (*data)[length:]
	return x, nil
}

func DecodeLenInt64Buf(buf *bytes.Buffer) (int64, error) {
	if buf.Len() == 0 {
		return 0, nil
	}

	val, err := buf.ReadByte()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("cannot read bytes from buffer")
		return 0, err
	}

	length := int(val)
	if buf.Len() < length {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "data_length": buf.Len(), "length": length}).Error("length of data is smaller then encoded length")
		return 0, fmt.Errorf(`length of data %d < %d`, buf.Len(), length)
	}
	data := make([]byte, 8)
	copy(data, buf.Next(length))

	return int64(binary.LittleEndian.Uint64(data)), nil

}

// DecodeLength decodes []byte to int64 and shifts buf. Bytes must be encoded with EncodeLength function.
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
		if len(*buf) < int(length+1) {
			log.WithFields(log.Fields{"data_length": len(*buf), "length": int(length + 1)}).Error("length of data is smaller then encoded length")
			return 0, fmt.Errorf(`input slice has small size`)
		}
		ret = int64(binary.BigEndian.Uint64(append(make([]byte, 8-length), (*buf)[1:length+1]...)))
	} else {
		ret = int64(length)
		length = 0
	}
	*buf = (*buf)[length+1:]
	return
}

func DecodeLengthBuf(buf *bytes.Buffer) (int, error) {
	if buf.Len() == 0 {
		return 0, nil
	}

	length, err := buf.ReadByte()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("cannot read bytes from buffer")
		return 0, err
	}

	if (length & 0x80) == 0 {
		return int(length), nil
	}

	length &= 0x7F
	if buf.Len() < int(length) {
		log.WithFields(log.Fields{"data_length": buf.Len(), "length": int(length), "type": consts.UnmarshallingError}).Error("length of data is smaller then encoded length")
		return 0, fmt.Errorf(`input slice has small size`)
	}

	n := int(binary.BigEndian.Uint64(append(make([]byte, 8-length), buf.Next(int(length))...)))
	if n < 0 {
		return 0, fmt.Errorf(`input slice has negative size`)
	}

	return n, nil
}

// BinMarshal converts v parameter to []byte slice.
func BinMarshal(out *[]byte, v interface{}) (*[]byte, error) {
	var err error

	t := reflect.ValueOf(v)
	if *out == nil {
		*out = make([]byte, 0, 2048)
	}

	switch t.Kind() {
	case reflect.Uint8, reflect.Int8:
		*out = append(*out, uint8(t.Uint()))
	case reflect.Uint32:
		tmp := make([]byte, 4)
		binary.BigEndian.PutUint32(tmp, uint32(t.Uint()))
		*out = append(*out, tmp...)
	case reflect.Int32:
		if uint32(t.Int()) < 128 {
			*out = append(*out, uint8(t.Int()))
		} else {
			var i uint8
			tmp := make([]byte, 4)
			binary.BigEndian.PutUint32(tmp, uint32(t.Int()))
			for ; i < 4; i++ {
				if tmp[i] != uint8(0) {
					break
				}
			}
			*out = append(*out, 128+4-i)
			*out = append(*out, tmp[i:]...)
		}
	case reflect.Float64:
		bin := float2Bytes(t.Float())
		*out = append(*out, bin...)
	case reflect.Int64:
		EncodeLenInt64(out, t.Int())
	case reflect.Uint64:
		tmp := make([]byte, 8)
		binary.BigEndian.PutUint64(tmp, t.Uint())
		*out = append(*out, tmp...)
	case reflect.String:
		*out = append(append(*out, EncodeLength(int64(t.Len()))...), []byte(t.String())...)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if out, err = BinMarshal(out, t.Field(i).Interface()); err != nil {
				return out, err
			}
		}
	case reflect.Slice:
		*out = append(append(*out, EncodeLength(int64(t.Len()))...), t.Bytes()...)
	case reflect.Ptr:
		if out, err = BinMarshal(out, t.Elem().Interface()); err != nil {
			return out, err
		}
	default:
		return out, fmt.Errorf(`unsupported type of BinMarshal`)
	}
	return out, nil
}

func BinUnmarshalBuff(buf *bytes.Buffer, v interface{}) error {
	t := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if buf.Len() == 0 {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": "input slice is empty"}).Error("input slice is empty")
		return fmt.Errorf(`input slice is empty`)
	}
	switch t.Kind() {
	case reflect.Uint8, reflect.Int8:
		val, err := buf.ReadByte()
		if err != nil {
			return err
		}
		t.SetUint(uint64(val))

	case reflect.Uint32:
		t.SetUint(uint64(binary.BigEndian.Uint32(buf.Next(4))))

	case reflect.Int32:
		val, err := buf.ReadByte()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading bytes from buffer")
			return err
		}
		if val < 128 {
			t.SetInt(int64(val))
		} else {
			var i uint8
			size := val - 128
			tmp := make([]byte, 4)
			if buf.Len() <= int(size) || size > 4 {
				log.WithFields(log.Fields{"type": consts.UnmarshallingError, "data_length": buf.Len(), "length": int(size)}).Error("bin unmarshalling int32")
				return fmt.Errorf(`wrong input data`)
			}
			for ; i < size; i++ {
				byteVal, err := buf.ReadByte()
				if err != nil {
					log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading bytes from buffer")
					return err
				}
				tmp[4-size+i] = byteVal
			}
			t.SetInt(int64(binary.BigEndian.Uint32(tmp)))
		}
	case reflect.Float64:
		t.SetFloat(bytes2Float(buf.Next(8)))

	case reflect.Int64:
		val, err := DecodeLenInt64Buf(buf)
		if err != nil {
			return err
		}
		t.SetInt(val)

	case reflect.Uint64:
		t.SetUint(binary.BigEndian.Uint64(buf.Next(8)))

	case reflect.String:
		val, err := DecodeLengthBuf(buf)
		if err != nil {
			return err
		}
		if buf.Len() < int(val) {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "data_length": buf.Len(), "length": int(val)}).Error("bin unmarshalling string")
			return fmt.Errorf(`input slice is short`)
		}
		t.SetString(string(buf.Next(val)))

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if err := BinUnmarshalBuff(buf, t.Field(i).Addr().Interface()); err != nil {
				return err
			}
		}
	case reflect.Slice:
		val, err := DecodeLengthBuf(buf)
		if err != nil {
			return err
		}
		if buf.Len() < int(val) {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "data_length": buf.Len(), "length": int(val)}).Error("bin unmarshalling slice")
			return fmt.Errorf(`input slice is short`)
		}
		t.SetBytes(buf.Next(int(val)))

	default:
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "value_type": t.Kind()}).Error("BinUnmrashal unsupported type")
		return fmt.Errorf(`unsupported type of BinUnmarshal %v`, t.Kind())
	}
	return nil

}

// BinUnmarshal converts []byte slice which has been made with BinMarshal to v
func BinUnmarshal(out *[]byte, v interface{}) error {
	t := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if len(*out) == 0 {
		return fmt.Errorf(`input slice is empty`)
	}
	switch t.Kind() {
	case reflect.Uint8, reflect.Int8:
		val := uint64((*out)[0])
		t.SetUint(val)
		*out = (*out)[1:]
	case reflect.Uint32:
		t.SetUint(uint64(binary.BigEndian.Uint32((*out)[:4])))
		*out = (*out)[4:]
	case reflect.Int32:
		val := (*out)[0]
		if val < 128 {
			t.SetInt(int64(val))
			*out = (*out)[1:]
		} else {
			var i uint8
			size := val - 128
			tmp := make([]byte, 4)
			if len(*out) <= int(size) || size > 4 {
				return fmt.Errorf(`wrong input data`)
			}
			for ; i < size; i++ {
				tmp[4-size+i] = (*out)[i+1]
			}
			t.SetInt(int64(binary.BigEndian.Uint32(tmp)))
			*out = (*out)[size+1:]
		}
	case reflect.Float64:
		t.SetFloat(bytes2Float((*out)[:8]))
		*out = (*out)[8:]
	case reflect.Int64:
		val, err := DecodeLenInt64(out)
		if err != nil {
			return err
		}
		t.SetInt(val)
	case reflect.Uint64:
		t.SetUint(binary.BigEndian.Uint64((*out)[:8]))
		*out = (*out)[8:]
	case reflect.String:
		val, err := DecodeLength(out)
		if err != nil {
			return err
		}
		if len(*out) < int(val) {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "data_length": len(*out), "length": int(val)}).Error("input slice is short")
			return fmt.Errorf(`input slice is short`)
		}
		t.SetString(string((*out)[:val]))
		*out = (*out)[val:]
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if err := BinUnmarshal(out, t.Field(i).Addr().Interface()); err != nil {
				return err
			}
		}
	case reflect.Slice:
		val, err := DecodeLength(out)
		if err != nil {
			return err
		}
		if len(*out) < int(val) {
			return fmt.Errorf(`input slice is short`)
		}
		t.SetBytes((*out)[:val])
		*out = (*out)[val:]
	default:
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "value_type": t.Kind()}).Error("BinUnmrashal unsupported type")
		return fmt.Errorf(`unsupported type of BinUnmarshal %v`, t.Kind())
	}
	return nil
}

// Sanitize deletes unaccessable characters from input string
func Sanitize(name string, available string) string {
	out := make([]rune, 0, len(name))
	for _, ch := range name {
		if ch > 127 || (ch >= '0' && ch <= '9') || ch == '_' || (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') || strings.IndexRune(available, ch) >= 0 {
			out = append(out, ch)
		}
	}
	return string(out)
}

// SanitizeScript deletes unaccessable characters from input string
func SanitizeScript(input string) string {
	return strings.Replace(strings.Replace(input, `<script`, `&lt;script`, -1), `script>`, `script&gt;`, -1)
}

// SanitizeName deletes unaccessable characters from name string
func SanitizeName(input string) string {
	return Sanitize(input, `- `)
}

// SanitizeNumber deletes unaccessable characters from number or name string
func SanitizeNumber(input string) string {
	return Sanitize(input, `+.- `)
}

func EscapeSQL(name string) string {
	return strings.Replace(strings.Replace(strings.Replace(name, `"`, `""`, -1),
		`;`, ``, -1), `'`, `''`, -1)
}

// EscapeName deletes unaccessable characters for input name(s)
func EscapeName(name string) string {
	out := make([]byte, 1, len(name)+2)
	out[0] = '"'
	available := `() ,`
	for _, ch := range []byte(name) {
		if (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' || (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') || strings.IndexByte(available, ch) >= 0 {
			out = append(out, ch)
		}
	}
	if strings.IndexAny(string(out), available) >= 0 {
		return string(out[1:])
	}
	return string(append(out, '"'))
}

// Float2Bytes converts float64 to []byte
func float2Bytes(float float64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, math.Float64bits(float))
	return bytes
}

// Bytes2Float converts []byte to float64
func bytes2Float(bytes []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(bytes))
}

// UInt32ToStr converts uint32 to string
func UInt32ToStr(num uint32) string {
	return strconv.FormatInt(int64(num), 10)
}

// Int64ToStr converts int64 to string
func Int64ToStr(num int64) string {
	return strconv.FormatInt(num, 10)
}

// Int64ToByte converts int64 to []byte
func Int64ToByte(num int64) []byte {
	return []byte(strconv.FormatInt(num, 10))
}

// IntToStr converts integer to string
func IntToStr(num int) string {
	return strconv.Itoa(num)
}

// DecToBin converts interface to []byte
func DecToBin(v interface{}, sizeBytes int64) []byte {
	var dec int64
	switch v.(type) {
	case int:
		dec = int64(v.(int))
	case int64:
		dec = v.(int64)
	case uint64:
		dec = int64(v.(uint64))
	case string:
		dec = StrToInt64(v.(string))
	}
	Hex := fmt.Sprintf("%0"+Int64ToStr(sizeBytes*2)+"x", dec)
	return HexToBin([]byte(Hex))
}

// BinToHex converts interface to hex []byte
func BinToHex(v interface{}) []byte {
	var bin []byte
	switch v.(type) {
	case []byte:
		bin = v.([]byte)
	case int64:
		bin = Int64ToByte(v.(int64))
	case string:
		bin = []byte(v.(string))
	}
	return []byte(fmt.Sprintf("%x", bin))
}

// HexToBin converts hex interface to binary []byte
func HexToBin(ihexdata interface{}) []byte {
	var hexdata string
	switch ihexdata.(type) {
	case []byte:
		hexdata = string(ihexdata.([]byte))
	case int64:
		hexdata = Int64ToStr(ihexdata.(int64))
	case string:
		hexdata = ihexdata.(string)
	}
	var str []byte
	str, err := hex.DecodeString(hexdata)
	if err != nil {
		log.WithFields(log.Fields{"data": hexdata, "error": err, "type": consts.ConversionError}).Error("decoding string to hex")
		log.Printf("HexToBin error: %s", err)
	}
	return str
}

// BinToDec converts input binary []byte to int64
func BinToDec(bin []byte) int64 {
	var a uint64
	l := len(bin)
	for i, b := range bin {
		shift := uint64((l - i - 1) * 8)
		a |= uint64(b) << shift
	}
	return int64(a)
}

// BinToDecBytesShift converts the input binary []byte to int64 and shifts the input bin
func BinToDecBytesShift(bin *[]byte, num int64) int64 {
	return BinToDec(BytesShift(bin, num))
}

// BytesShift returns the index bytes of the input []byte and shift str pointer
func BytesShift(str *[]byte, index int64) (ret []byte) {
	if int64(len(*str)) < index || index == 0 {
		*str = (*str)[:0]
		return []byte{}
	}
	ret, *str = (*str)[:index], (*str)[index:]
	return
}

// InterfaceToStr converts the interfaces to the string
func InterfaceToStr(v interface{}) (string, error) {
	var str string
	if v == nil {
		return ``, nil
	}
	switch v.(type) {
	case int:
		str = IntToStr(v.(int))
	case float64:
		str = Float64ToStr(v.(float64))
	case int64:
		str = Int64ToStr(v.(int64))
	case string:
		str = v.(string)
	case []byte:
		str = string(v.([]byte))
	default:
		if reflect.TypeOf(v).String() == `map[string]interface {}` ||
			reflect.TypeOf(v).String() == `*types.Map` {
			if out, err := json.Marshal(v); err != nil {
				log.WithFields(log.Fields{"error": err, "type": consts.JSONMarshallError}).Error("marshalling map for jsonb")
				return ``, err
			} else {
				str = string(out)
			}
		} else if reflect.TypeOf(v).String() == `decimal.Decimal` {
			str = v.(decimal.Decimal).String()
		}
	}
	return str, nil
}

// InterfaceSliceToStr converts the slice of interfaces to the slice of strings
func InterfaceSliceToStr(i []interface{}) (strs []string, err error) {
	var val string
	for _, v := range i {
		val, err = InterfaceToStr(v)
		if err != nil {
			return
		}
		strs = append(strs, val)
	}
	return
}

// InterfaceToFloat64 converts the interfaces to the float64
func InterfaceToFloat64(i interface{}) float64 {
	var result float64
	switch i.(type) {
	case int:
		result = float64(i.(int))
	case float64:
		result = i.(float64)
	case int64:
		result = float64(i.(int64))
	case string:
		result = StrToFloat64(i.(string))
	case []byte:
		result = BytesToFloat64(i.([]byte))
	}
	return result
}

// BytesShiftReverse gets []byte from the end of the input and cut the input pointer to []byte
func BytesShiftReverse(str *[]byte, v interface{}) []byte {
	var index int64
	switch v.(type) {
	case int:
		index = int64(v.(int))
	case int64:
		index = v.(int64)
	}

	var substr []byte
	slen := int64(len(*str))
	if slen < index {
		index = slen
	}
	substr = (*str)[slen-index:]
	*str = (*str)[:slen-index]
	return substr
}

// StrToInt64 converts string to int64
func StrToInt64(s string) int64 {
	int64, _ := strconv.ParseInt(s, 10, 64)
	return int64
}

// BytesToInt64 converts []bytes to int64
func BytesToInt64(s []byte) int64 {
	int64, _ := strconv.ParseInt(string(s), 10, 64)
	return int64
}

// StrToUint64 converts string to the unsinged int64
func StrToUint64(s string) uint64 {
	ret, _ := strconv.ParseUint(s, 10, 64)
	return ret
}

// StrToInt converts string to integer
func StrToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// Float64ToStr converts float64 to string
func Float64ToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 13, 64)
}

// StrToFloat64 converts string to float64
func StrToFloat64(s string) float64 {
	Float64, _ := strconv.ParseFloat(s, 64)
	return Float64
}

// BytesToFloat64 converts []byte to float64
func BytesToFloat64(s []byte) float64 {
	Float64, _ := strconv.ParseFloat(string(s), 64)
	return Float64
}

// BytesToInt converts []byte to integer
func BytesToInt(s []byte) int {
	i, _ := strconv.Atoi(string(s))
	return i
}

// StrToMoney rounds money string to float64
func StrToMoney(str string) float64 {
	ind := strings.Index(str, ".")
	var new string
	if ind != -1 {
		end := 2
		if len(str[ind+1:]) > 1 {
			end = 3
		}
		new = str[:ind] + "." + str[ind+1:ind+end]
	} else {
		new = str
	}
	return StrToFloat64(new)
}

// AddressToString converts int64 address to apla address as XXXX-...-XXXX.
func AddressToString(address int64) (ret string) {
	num := strconv.FormatUint(uint64(address), 10)
	val := []byte(strings.Repeat("0", 20-len(num)) + num)

	for i := 0; i < 4; i++ {
		ret += string(val[i*4:(i+1)*4]) + `-`
	}
	ret += string(val[16:])
	return
}

// EncodeLengthPlusData encoding interface into []byte
func EncodeLengthPlusData(idata interface{}) []byte {
	var data []byte
	switch idata.(type) {
	case int64:
		data = Int64ToByte(idata.(int64))
	case string:
		data = []byte(idata.(string))
	case []byte:
		data = idata.([]byte)
	}
	//log.Debug("data: %x", data)
	//log.Debug("len data: %d", len(data))
	return append(EncodeLength(int64(len(data))), data...)
}

// StringToAddress converts string apla address to int64 address. The input address can be a positive or negative
// number, or apla address in XXXX-...-XXXX format. Returns 0 when error occurs.
func StringToAddress(address string) (result int64) {
	var (
		err error
		ret uint64
	)
	if len(address) == 0 {
		return 0
	}
	if address[0] == '-' {
		var id int64
		id, err = strconv.ParseInt(address, 10, 64)
		if err != nil {
			return 0
		}
		address = strconv.FormatUint(uint64(id), 10)
	}
	if len(address) < 20 {
		address = strings.Repeat(`0`, 20-len(address)) + address
	}

	val := []byte(strings.Replace(address, `-`, ``, -1))
	if len(val) != 20 {
		return
	}
	if ret, err = strconv.ParseUint(string(val), 10, 64); err != nil {
		return 0
	}
	if checkSum(val[:len(val)-1]) != int(val[len(val)-1]-'0') {
		return 0
	}
	result = int64(ret)
	return
}

// CheckSum calculates the 0-9 check sum of []byte
func checkSum(val []byte) int {
	var one, two int
	for i, ch := range val {
		digit := int(ch - '0')
		if i&1 == 1 {
			one += digit
		} else {
			two += digit
		}
	}
	checksum := (two + 3*one) % 10
	if checksum > 0 {
		checksum = 10 - checksum
	}
	return checksum
}

// EGSMoney converts qEGS to EGS. For example, 123455000000000000000 => 123.455
func EGSMoney(money string) string {
	digit := consts.MoneyDigits
	if len(money) < digit+1 {
		money = strings.Repeat(`0`, digit+1-len(money)) + money
	}
	money = money[:len(money)-digit] + `.` + money[len(money)-digit:]
	return strings.TrimRight(strings.TrimRight(money, `0`), `.`)
}

// EscapeForJSON replaces quote to slash and quote
func EscapeForJSON(data string) string {
	return strings.Replace(data, `"`, `\"`, -1)
}

// ValidateEmail validates email
func ValidateEmail(email string) bool {
	Re := regexp.MustCompile(`^(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

// ParseName gets a state identifier and the name of the contract or table
// from the full name like @[id]name
func ParseName(in string) (id int64, name string) {
	re := regexp.MustCompile(`(?is)^@(\d+)(\w[_\w\d]*)$`)
	ret := re.FindStringSubmatch(in)
	if len(ret) == 3 {
		id = StrToInt64(ret[1])
		name = ret[2]
	}
	return
}

func ParseTable(tblname string, defaultEcosystem int64) string {
	ecosystem, name := ParseName(tblname)
	if ecosystem == 0 {
		if FirstEcosystemTables[tblname] {
			ecosystem = 1
		} else {
			ecosystem = defaultEcosystem
		}
		name = tblname
	}
	return strings.ToLower(fmt.Sprintf(`%d_%s`, ecosystem, Sanitize(name, ``)))
}

func IsByteColumn(table, column string) bool {
	predefined := map[string]string{"txhash": "history", "pub": "keys", "data": "binaries"}
	if suffix, ok := predefined[column]; ok {
		re := regexp.MustCompile(`(?i)^\d+_` + suffix + `$`)
		return re.MatchString(table)
	}
	return false
}

// SliceReverse reverses the slice of int64
func SliceReverse(s []int64) []int64 {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// SortMap sorts map to the slice of maps
func SortMap(m map[int64]string) []map[int64]string {
	var keys []int
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	var result []map[int64]string
	for _, k := range keys {
		result = append(result, map[int64]string{int64(k): m[int64(k)]})
	}
	return result
}

// RSortMap sorts map to the reversed slice of maps
func RSortMap(m map[int64]string) []map[int64]string {

	var keys []int
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	var result []map[int64]string
	for _, k := range keys {
		result = append(result, map[int64]string{int64(k): m[int64(k)]})
	}
	return result
}

// InSliceString searches the string in the slice of strings
func InSliceString(search string, slice []string) bool {
	for _, v := range slice {
		if v == search {
			return true
		}
	}
	return false
}

// StripTags replaces < and > to &lt; and &gt;
func StripTags(value string) string {
	return strings.Replace(strings.Replace(value, `<`, `&lt;`, -1), `>`, `&gt;`, -1)
}

// IsLatin checks if the specified string contains only latin character, digits and '-', '_'.
func IsLatin(name string) bool {
	for _, ch := range []byte(name) {
		if !((ch >= '0' && ch <= '9') || ch == '_' || ch == '-' || (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z')) {
			return false
		}
	}
	return true
}

// IsValidAddress checks if the specified address is apla address.
func IsValidAddress(address string) bool {
	val := []byte(strings.Replace(address, `-`, ``, -1))
	if len(val) != 20 {
		return false
	}
	if _, err := strconv.ParseUint(string(val), 10, 64); err != nil {
		return false
	}
	return checkSum(val[:len(val)-1]) == int(val[len(val)-1]-'0')
}

// Escape deletes unaccessable characters
func Escape(data string) string {
	out := make([]rune, 0, len(data))
	available := `_ ,=!-'()"?*$#{}<>: `
	for _, ch := range []rune(data) {
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') || strings.IndexByte(available, byte(ch)) >= 0 ||
			unicode.IsLetter(ch) {
			out = append(out, ch)
		}
	}
	return string(out)
}

// FieldToBytes returns the value of n-th field of v as []byte
func FieldToBytes(v interface{}, num int) []byte {
	t := reflect.ValueOf(v)
	ret := make([]byte, 0, 2048)
	if t.Kind() == reflect.Struct && num < t.NumField() {
		field := t.Field(num)
		switch field.Kind() {
		case reflect.Uint8, reflect.Uint32, reflect.Uint64:
			ret = append(ret, []byte(fmt.Sprintf("%d", field.Uint()))...)
		case reflect.Int8, reflect.Int32, reflect.Int64:
			ret = append(ret, []byte(fmt.Sprintf("%d", field.Int()))...)
		case reflect.Float64:
			ret = append(ret, []byte(fmt.Sprintf("%f", field.Float()))...)
		case reflect.String:
			ret = append(ret, []byte(field.String())...)
		case reflect.Slice:
			ret = append(ret, field.Bytes()...)
			//		case reflect.Ptr:
			//		case reflect.Struct:
			//		default:
		}
	}
	return ret
}

// NumString insert spaces between each three digits. 7123456 => 7 123 456
func NumString(in string) string {
	if strings.IndexByte(in, '.') >= 0 {
		lr := strings.Split(in, `.`)
		return NumString(lr[0]) + `.` + lr[1]
	}
	buf := []byte(in)
	out := make([]byte, len(in)+4)
	for len(buf) > 3 {
		out = append(append([]byte(` `), buf[len(buf)-3:]...), out...)
		buf = buf[:len(buf)-3]
	}
	return string(append(buf, out...))
}

func Round(num float64) int64 {
	//log.Debug("num", num)
	//num += ROUND_FIX
	//	return int(StrToFloat64(Float64ToStr(num)) + math.Copysign(0.5, num))
	//log.Debug("num", num)
	return int64(num + math.Copysign(0.5, num))
}

// RoundWithPrecision rounds float64 value
func RoundWithPrecision(num float64, precision int) float64 {
	num += consts.ROUND_FIX
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}

// RoundWithoutPrecision is round float64 without precision
func RoundWithoutPrecision(num float64) int64 {
	//log.Debug("num", num)
	//num += ROUND_FIX
	//	return int(StrToFloat64(Float64ToStr(num)) + math.Copysign(0.5, num))
	//log.Debug("num", num)
	return int64(num + math.Copysign(0.5, num))
}

// ValueToInt converts interface (string or int64) to int64
func ValueToInt(v interface{}) (ret int64, err error) {
	switch val := v.(type) {
	case float64:
		ret = int64(val)
	case int64:
		ret = val
	case string:
		if len(val) == 0 {
			return 0, nil
		}
		ret, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			errText := err.Error()
			if strings.Contains(errText, `:`) {
				errText = errText[strings.LastIndexByte(errText, ':'):]
			} else {
				errText = ``
			}
			err = fmt.Errorf(`%s is not a valid integer %s`, val, errText)
		}
	default:
		if v == nil {
			return 0, nil
		}
		err = fmt.Errorf(`%v is not a valid integer`, val)
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err,
			"value": fmt.Sprint(v)}).Error("converting value to int")
	}
	return
}

func Int64ToDateStr(date int64, format string) string {
	t := time.Unix(date, 0)
	return t.Format(format)
}
