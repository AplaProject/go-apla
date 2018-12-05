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
	"net/http"
	"strconv"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func getAvatarHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	memberID := converter.StrToInt64(params["member"])
	ecosystemID := converter.StrToInt64(params["ecosystem"])

	member := &model.Member{}
	member.SetTablePrefix(converter.Int64ToStr(ecosystemID))

	found, err := member.Get(memberID)
	if err != nil {
		logger.WithFields(log.Fields{
			"type":      consts.DBError,
			"error":     err,
			"ecosystem": ecosystemID,
			"member_id": memberID,
		}).Error("getting member")
		errorResponse(w, err)
		return
	}

	if !found {
		errorResponse(w, errNotFound)
		return
	}

	if member.ImageID == nil {
		errorResponse(w, errNotFound)
		return
	}

	bin := &model.Binary{}
	bin.SetTablePrefix(converter.Int64ToStr(ecosystemID))
	found, err = bin.GetByID(*member.ImageID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "image_id": *member.ImageID}).Errorf("on getting binary by id")
		errorResponse(w, err)
		return
	}

	if !found {
		errorResponse(w, errNotFound)
		return
	}

	if len(bin.Data) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject, "error": err, "image_id": *member.ImageID}).Errorf("on check avatar size")
		errorResponse(w, errNotFound)
		return
	}

	w.Header().Set("Content-Type", bin.MimeType)
	w.Header().Set("Content-Length", strconv.Itoa(len(bin.Data)))
	if _, err := w.Write(bin.Data); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("unable to write image")
	}
}
