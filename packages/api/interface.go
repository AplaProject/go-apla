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
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	log "github.com/sirupsen/logrus"
)

type componentModel interface {
	SetTablePrefix(prefix string)
	Get(name string) (bool, error)
}

func getPageRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	return getInterfaceRow(w, r, data, logger, &model.Page{})
}

func getMenuRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	return getInterfaceRow(w, r, data, logger, &model.Menu{})
}

func getBlockInterfaceRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	return getInterfaceRow(w, r, data, logger, &model.BlockInterface{})
}

func getInterfaceRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry, c componentModel) error {
	c.SetTablePrefix(getPrefix(data))
	ok, err := c.Get(data.ParamString("name"))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting one row")
		return errorAPI(w, `E_QUERY`, http.StatusInternalServerError)
	}
	if !ok {
		return errorAPI(w, `E_NOTFOUND`, http.StatusInternalServerError)
	}

	data.result = c

	return nil
}
