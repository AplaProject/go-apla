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
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type paramValue struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

type ecosystemParamsResult struct {
	List []paramValue `json:"list"`
}

func ecosystemParams(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var (
		result ecosystemParamsResult
		names  map[string]bool
	)
	_, prefix, err := checkEcosystem(w, data, logger)
	if err != nil {
		return err
	}
	sp := &model.StateParameter{}
	sp.SetTablePrefix(prefix)
	list, err := sp.GetAllStateParameters()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting all state parameters")
	}
	result.List = make([]paramValue, 0)
	if len(data.params[`names`].(string)) > 0 {
		names = make(map[string]bool)
		for _, item := range strings.Split(data.params[`names`].(string), `,`) {
			names[item] = true
		}
	}
	for _, item := range list {
		if names != nil && !names[item.Name] {
			continue
		}
		result.List = append(result.List, paramValue{ID: converter.Int64ToStr(item.ID),
			Name: item.Name, Value: item.Value, Conditions: item.Conditions})
	}
	data.result = &result
	return
}
