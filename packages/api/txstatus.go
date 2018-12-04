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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type txstatusError struct {
	Type  string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
	Id    string `json:"id,omitempty"`
}

type txstatusResult struct {
	BlockID string         `json:"blockid"`
	Message *txstatusError `json:"errmsg,omitempty"`
	Result  string         `json:"result"`
}

func getTxStatus(r *http.Request, hash string) (*txstatusResult, error) {
	logger := getLogger(r)

	var status txstatusResult
	if _, err := hex.DecodeString(hash); err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding tx hash from hex")
		return nil, errHashWrong
	}
	ts := &model.TransactionStatus{}
	found, err := ts.Get([]byte(converter.HexToBin(hash)))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("getting transaction status by hash")
		return nil, err
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": []byte(converter.HexToBin(hash))}).Error("getting transaction status by hash")
		return nil, errHashNotFound
	}
	if ts.BlockID > 0 {
		status.BlockID = converter.Int64ToStr(ts.BlockID)
		status.Result = ts.Error
	} else if len(ts.Error) > 0 {
		if err := json.Unmarshal([]byte(ts.Error), &status.Message); err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "text": ts.Error, "error": err}).Warn("unmarshalling txstatus error")
			status.Message = &txstatusError{
				Type:  "txError",
				Error: ts.Error,
			}
		}
	}
	return &status, nil
}

type multiTxStatusResult struct {
	Results map[string]*txstatusResult `json:"results"`
}

type txstatusRequest struct {
	Hashes []string `json:"hashes"`
}

func getTxStatusHandler(w http.ResponseWriter, r *http.Request) {
	result := &multiTxStatusResult{}
	result.Results = map[string]*txstatusResult{}

	var request txstatusRequest
	if err := json.Unmarshal([]byte(r.FormValue("data")), &request); err != nil {
		errorResponse(w, errHashWrong)
		return
	}
	for _, hash := range request.Hashes {
		status, err := getTxStatus(r, hash)
		if err != nil {
			errorResponse(w, err)
			return
		}
		result.Results[hash] = status
	}

	jsonResponse(w, result)
}
