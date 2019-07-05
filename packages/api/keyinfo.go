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
	"encoding/json"
	"net/http"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type roleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type notifyInfo struct {
	RoleID string `json:"role_id"`
	Count  int64  `json:"count"`
}

type keyInfoResult struct {
	Account       string       `json:"account"`
	Ecosystem     string       `json:"ecosystem"`
	Name          string       `json:"name"`
	Roles         []roleInfo   `json:"roles,omitempty"`
	Notifications []notifyInfo `json:"notifications,omitempty"`
}

func (m Mode) getKeyInfoHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)

	keysList := make([]*keyInfoResult, 0)
	keyID := converter.StringToAddress(params["wallet"])
	if keyID == 0 {
		errorResponse(w, errInvalidWallet.Errorf(params["wallet"]))
		return
	}

	ids, names, err := m.EcosysLookupGetter.GetEcosystemLookup()
	if err != nil {
		errorResponse(w, err)
		return
	}

	var (
		found bool
	)

	for i, ecosystemID := range ids {
		key := &model.Key{}
		key.SetTablePrefix(ecosystemID)
		found, err = key.Get(keyID)
		if err != nil {
			errorResponse(w, err)
			return
		}
		if !found {
			continue
		}
		keyRes := &keyInfoResult{
			Account:   key.AccountID,
			Ecosystem: converter.Int64ToStr(ecosystemID),
			Name:      names[i],
		}
		ra := &model.RolesParticipants{}
		roles, err := ra.SetTablePrefix(ecosystemID).GetActiveMemberRoles(key.AccountID)
		if err != nil {
			errorResponse(w, err)
			return
		}
		for _, r := range roles {
			var role roleInfo
			if err := json.Unmarshal([]byte(r.Role), &role); err != nil {
				logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling role")
				errorResponse(w, err)
				return
			}
			keyRes.Roles = append(keyRes.Roles, role)
		}
		keyRes.Notifications, err = m.getNotifications(ecosystemID, key)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting notifications")
			errorResponse(w, err)
			return
		}

		keysList = append(keysList, keyRes)
	}

	// in test mode, registration is open in the first ecosystem
	if len(keysList) == 0 && syspar.IsTestMode() {
		keysList = append(keysList, &keyInfoResult{
			Account:   converter.AddressToString(keyID),
			Ecosystem: converter.Int64ToStr(ids[0]),
			Name:      names[0],
		})
	}

	jsonResponse(w, &keysList)
}

func (m Mode) getNotifications(ecosystemID int64, key *model.Key) ([]notifyInfo, error) {
	notif, err := model.GetNotificationsCount(ecosystemID, []string{key.AccountID})
	if err != nil {
		return nil, err
	}

	list := make([]notifyInfo, 0)
	for _, n := range notif {
		if n.RecipientID != key.ID {
			continue
		}

		list = append(list, notifyInfo{
			RoleID: converter.Int64ToStr(n.RoleID),
			Count:  n.Count,
		})
	}
	return list, nil
}
