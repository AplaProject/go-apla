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

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"

	log "github.com/sirupsen/logrus"
)

type txinfoResult struct {
	BlockID string        `json:"blockid"`
	Confirm int           `json:"confirm"`
	Data    *smart.TxInfo `json:"data,omitempty"`
}

type multiTxInfoResult struct {
	Results map[string]*txinfoResult `json:"results"`
}

func getTxInfo(txHash string, w http.ResponseWriter, cntInfo bool) (*txinfoResult, error) {
	var status txinfoResult
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, errorAPI(w, `E_HASHWRONG`, http.StatusBadRequest)
	}
	ltx := &model.LogTransaction{Hash: hash}
	found, err := ltx.GetByHash(hash)
	if err != nil {
		return nil, errorAPI(w, err, http.StatusInternalServerError)
	}
	if !found {
		return &status, nil
	}
	status.BlockID = converter.Int64ToStr(ltx.Block)
	var confirm model.Confirmation
	found, err = confirm.GetConfirmation(ltx.Block)
	if err != nil {
		return nil, errorAPI(w, err, http.StatusInternalServerError)
	}
	if found {
		status.Confirm = int(confirm.Good)
	}
	if cntInfo {
		status.Data, err = smart.TransactionData(ltx.Block, hash)
		if err != nil {
			return nil, errorAPI(w, err, http.StatusInternalServerError)
		}
	}
	return &status, nil
}

func txinfo(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	status, err := getTxInfo(data.params[`hash`].(string), w, data.params[`contractinfo`].(int64) > 0)
	if err != nil {
		return err
	}
	data.result = &status
	return nil
}

func txinfoMulti(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	result := &multiTxInfoResult{}
	result.Results = map[string]*txinfoResult{}
	var request struct {
		Hashes []string `json:"hashes"`
	}
	if err := json.Unmarshal([]byte(data.params["data"].(string)), &request); err != nil {
		return errorAPI(w, `E_HASHWRONG`, http.StatusBadRequest)
	}
	for _, hash := range request.Hashes {
		status, err := getTxInfo(hash, w, data.params[`contractinfo`].(int64) > 0)
		if err != nil {
			return err
		}
		result.Results[hash] = status
	}
	data.result = result
	return nil
}
