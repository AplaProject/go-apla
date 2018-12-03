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
	"encoding/json"
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/notificator"

	log "github.com/sirupsen/logrus"
)

type idItem struct {
	ID          string `json:"id"`
	EcosystemID string `json:"ecosystem"`
}

type updateNotificatorResult struct {
	Result bool `json:"result"`
}

func updateNotificator(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var list []idItem

	err := json.Unmarshal([]byte(data.params["ids"].(string)), &list)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling ids")
		return errorAPI(w, err, http.StatusInternalServerError)
	}

	stateList := make(map[int64][]int64)

	for _, item := range list {
		ecosystem := converter.StrToInt64(item.EcosystemID)
		if _, ok := stateList[ecosystem]; !ok {
			stateList[ecosystem] = make([]int64, 0)
		}
		stateList[ecosystem] = append(stateList[ecosystem], converter.StrToInt64(item.ID))
	}

	go notificator.SendNotificationsByRequest(stateList)
	data.result = &updateNotificatorResult{Result: true}
	return nil
}
