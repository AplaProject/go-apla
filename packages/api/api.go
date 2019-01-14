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

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/schema"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/types"

	log "github.com/sirupsen/logrus"
)

const (
	multipartBuf      = 100000 // the buffer size for ParseMultipartForm
	multipartFormData = "multipart/form-data"
	contentType       = "Content-Type"
)

type Mode struct {
	EcosysIDValidator  types.EcosystemIDValidator
	EcosysNameGetter   types.EcosystemNameGetter
	EcosysLookupGetter types.EcosystemLookupGetter
	ContractRunner     types.SmartContractRunner
	ClientTxProcessor  types.ClientTxPreprocessor
}

// Client represents data of client
type Client struct {
	KeyID         int64
	EcosystemID   int64
	EcosystemName string
	RoleID        int64
	IsMobile      bool
}

func (c *Client) Prefix() string {
	return converter.Int64ToStr(c.EcosystemID)
}

func jsonResponse(w http.ResponseWriter, v interface{}) {
	jsonResult, err := json.Marshal(v)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marhsalling http response to json")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonResult)
}

func errorResponse(w http.ResponseWriter, err error, code ...int) {
	et, ok := err.(errType)
	if !ok {
		et = errServer
		et.Message = err.Error()
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	if len(code) == 0 {
		w.WriteHeader(et.Status)
	} else {
		w.WriteHeader(code[0])
	}

	jsonResponse(w, et)
}

type formValidator interface {
	Validate(r *http.Request) error
}

type nopeValidator struct{}

func (np nopeValidator) Validate(r *http.Request) error {
	return nil
}

func parseForm(r *http.Request, f formValidator) (err error) {
	if isMultipartForm(r) {
		err = r.ParseMultipartForm(multipartBuf)
	} else {
		err = r.ParseForm()
	}
	if err != nil {
		return
	}

	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(f, r.Form); err != nil {
		return err
	}
	return f.Validate(r)
}

func isMultipartForm(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get(contentType), multipartFormData)
}

type hexValue struct {
	value []byte
}

func (hv hexValue) Bytes() []byte {
	return hv.value
}

func (hv *hexValue) UnmarshalText(v []byte) (err error) {
	hv.value, err = hex.DecodeString(string(v))
	return
}
