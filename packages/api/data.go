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
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"

	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

const binaryColumn = "data"

var errWrongHash = errors.New("Wrong hash")

func compareHash(w http.ResponseWriter, bin *model.Binary, data []byte, ps hr.Params) bool {
	urlHash := strings.ToLower(ps.ByName(`hash`))
	if len(urlHash) == 32 && fmt.Sprintf(`%x`, md5.Sum(data)) == urlHash {
		return true
	}
	if len(urlHash) == 64 {
		var hashData string
		if bin == nil {
			hash, _ := crypto.Hash([]byte(data))
			hashData = fmt.Sprintf(`%x`, hash)
		} else {
			hashData = bin.Hash
		}
		if hashData == urlHash {
			return true
		}
	}
	errorAPI(w, errWrongHash, http.StatusNotFound)
	return false
}

func dataHandler() hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		tblname := ps.ByName("table")
		column := ps.ByName("column")

		if strings.Contains(tblname, model.BinaryTableSuffix) && column == binaryColumn {
			binary(w, r, ps)
			return
		}

		id := ps.ByName(`id`)
		data, err := model.GetColumnByID(tblname, column, id)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting data from table")
			errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
			return
		}

		if !compareHash(w, nil, []byte(data), ps) {
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(data))
		return
	})
}

func binary(w http.ResponseWriter, r *http.Request, ps hr.Params) {
	bin := model.Binary{}
	bin.SetTableName(ps.ByName("table"))

	found, err := bin.GetByID(converter.StrToInt64(ps.ByName("id")))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Errorf("getting binary by id")
		errorAPI(w, "E_SERVER", http.StatusInternalServerError)
		return
	}

	if !found {
		errorAPI(w, "E_SERVER", http.StatusNotFound)
		return
	}

	if !compareHash(w, &bin, bin.Data, ps) {
		return
	}

	w.Header().Set("Content-Type", bin.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, bin.Name))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(bin.Data)
	return
}
