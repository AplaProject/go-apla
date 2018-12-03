// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package consts

import (
	"reflect"
)

// BlockHeader is a structure of the block header
type BlockHeader struct {
	Type     byte
	BlockID  uint32
	Time     uint32
	WalletID int64
	StateID  byte
	Sign     []byte
}

// TxHeader is the old version of the transaction header
type TxHeader struct {
	Type  byte
	Time  uint32
	KeyID int64
}

// FirstBlock is the header of FirstBlock transaction
type FirstBlock struct {
	TxHeader
	PublicKey             []byte
	NodePublicKey         []byte
	StopNetworkCertBundle []byte
	Test                  int64
	PrivateBlockchain     uint64
}

type StopNetwork struct {
	TxHeader
	StopNetworkCert []byte
}

// Don't forget to insert the structure in init() - list

var blockStructs = make(map[string]reflect.Type)

func init() {
	// New structures must be inserted here
	list := []interface{}{
		FirstBlock{},
		StopNetwork{},
	}

	for _, item := range list {
		blockStructs[reflect.TypeOf(item).Name()] = reflect.TypeOf(item)
	}
}

// MakeStruct is only used for FirstBlock now
func MakeStruct(name string) interface{} {
	v := reflect.New(blockStructs[name]) //.Elem()
	return v.Interface()
}

// IsStruct is only used for FirstBlock now
func IsStruct(tx int64) bool {
	return tx == TxTypeFirstBlock || tx == TxTypeStopNetwork
}

// Header returns TxHeader
func Header(v interface{}) TxHeader {
	return reflect.ValueOf(v).Elem().Field(0).Interface().(TxHeader)
}

// Sign returns the signature attached to the header
func Sign(v interface{}) (sign []byte) {
	field := reflect.ValueOf(v).Elem().FieldByName(`Sign`)
	if field.IsValid() && field.Kind() == reflect.Slice {
		sign = field.Bytes()
	}
	return
}
