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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

const binaryColumn = "data"

func compareHash(data []byte, urlHash string) bool {
	urlHash = strings.ToLower(urlHash)

	var hash []byte
	switch len(urlHash) {
	case 32:
		h := md5.Sum(data)
		hash = h[:]
	case 64:
		hash, _ = crypto.Hash(data)
	}

	return hex.EncodeToString(hash) == urlHash
}

func getDataHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	table, column := params["table"], params["column"]

	data, err := model.GetColumnByID(table, column, params["id"])
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting data from table")
		errorResponse(w, errNotFound)
		return
	}

	if !compareHash([]byte(data), params["hash"]) {
		errorResponse(w, errHashWrong)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(data))
	return
}

func getBinaryHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	bin := model.Binary{}
	found, err := bin.GetByID(converter.StrToInt64(params["id"]))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Errorf("getting binary by id")
		errorResponse(w, err)
		return
	}

	if !found {
		errorResponse(w, errNotFound)
		return
	}

	if !compareHash(bin.Data, params["hash"]) {
		errorResponse(w, errHashWrong)
		return
	}

	w.Header().Set("Content-Type", bin.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, bin.Name))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(bin.Data)
	return
}
