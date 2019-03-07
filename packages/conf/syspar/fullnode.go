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

package syspar

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	log "github.com/sirupsen/logrus"
)

const publicKeyLength = 64

var (
	errFullNodeInvalidValues = errors.New("Invalid values of the full_node parameter")
)

//because of PublicKey is byte
type fullNodeJSON struct {
	TCPAddress string      `json:"tcp_address"`
	APIAddress string      `json:"api_address"`
	KeyID      json.Number `json:"key_id"`
	PublicKey  string      `json:"public_key"`
	UnbanTime  json.Number `json:"unban_time,er"`
	Stopped    bool        `json:"stopped"`
}

// FullNode is storing full node data
type FullNode struct {
	TCPAddress string
	APIAddress string
	KeyID      int64
	PublicKey  []byte
	UnbanTime  time.Time
	Stopped    bool
}

// UnmarshalJSON is custom json unmarshaller
func (fn *FullNode) UnmarshalJSON(b []byte) (err error) {
	data := fullNodeJSON{}
	if err = json.Unmarshal(b, &data); err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err, "value": string(b)}).Error("Unmarshalling full nodes to json")
		return err
	}

	fn.TCPAddress = data.TCPAddress
	fn.APIAddress = data.APIAddress
	fn.KeyID = converter.StrToInt64(data.KeyID.String())
	fn.Stopped = data.Stopped
	if fn.PublicKey, err = crypto.HexToPub(data.PublicKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": data.PublicKey}).Error("converting full nodes public key from hex")
		return err
	}
	fn.UnbanTime = time.Unix(converter.StrToInt64(data.UnbanTime.String()), 0)

	if err = fn.Validate(); err != nil {
		return err
	}

	return nil
}

func (fn *FullNode) MarshalJSON() ([]byte, error) {
	jfn := fullNodeJSON{
		TCPAddress: fn.TCPAddress,
		APIAddress: fn.APIAddress,
		KeyID:      json.Number(strconv.FormatInt(fn.KeyID, 10)),
		PublicKey:  crypto.PubToHex(fn.PublicKey),
		UnbanTime:  json.Number(strconv.FormatInt(fn.UnbanTime.Unix(), 10)),
	}

	data, err := json.Marshal(jfn)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Marshalling full nodes to json")
		return nil, err
	}

	return data, nil
}

// ValidateURL returns error if the URL is invalid
func validateURL(rawurl string) error {
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return err
	}

	if len(u.Scheme) == 0 {
		return fmt.Errorf("Invalid scheme: %s", rawurl)
	}

	if len(u.Host) == 0 {
		return fmt.Errorf("Invalid host: %s", rawurl)
	}

	return nil
}

// Validate checks values
func (fn *FullNode) Validate() error {
	if fn.KeyID == 0 || len(fn.PublicKey) != publicKeyLength || len(fn.TCPAddress) == 0 {
		return errFullNodeInvalidValues
	}

	if err := validateURL(fn.APIAddress); err != nil {
		return err
	}

	return nil
}
