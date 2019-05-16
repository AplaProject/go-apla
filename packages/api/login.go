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
	"fmt"
	"net/http"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/publisher"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

// Special word used by frontend to sign UID generated by /getuid API command, sign is performed for contcatenated word and UID
func nonceSalt() string {
	return fmt.Sprintf("LOGIN%d", conf.Config.NetworkID)
}

type loginForm struct {
	EcosystemID int64          `schema:"ecosystem"`
	Expire      int64          `schema:"expire"`
	PublicKey   publicKeyValue `schema:"pubkey"`
	KeyID       string         `schema:"key_id"`
	Signature   hexValue       `schema:"signature"`
	RoleID      int64          `schema:"role_id"`
	IsMobile    bool           `schema:"mobile"`
}

type publicKeyValue struct {
	hexValue
}

func (pk *publicKeyValue) UnmarshalText(v []byte) (err error) {
	pk.value, err = hex.DecodeString(string(v))
	pk.value = crypto.CutPub(pk.value)
	return
}

func (f *loginForm) Validate(r *http.Request) error {
	if f.Expire == 0 {
		f.Expire = int64(jwtExpire)
	}

	return nil
}

type loginResult struct {
	Token       string        `json:"token,omitempty"`
	EcosystemID string        `json:"ecosystem_id,omitempty"`
	KeyID       string        `json:"key_id,omitempty"`
	Address     string        `json:"address,omitempty"`
	NotifyKey   string        `json:"notify_key,omitempty"`
	IsNode      bool          `json:"isnode,omitempty"`
	IsOwner     bool          `json:"isowner,omitempty"`
	IsOBS       bool          `json:"obs,omitempty"`
	Timestamp   string        `json:"timestamp,omitempty"`
	Roles       []rolesResult `json:"roles,omitempty"`
}

type rolesResult struct {
	RoleId   int64  `json:"role_id"`
	RoleName string `json:"role_name"`
}

func (m Mode) loginHandler(w http.ResponseWriter, r *http.Request) {
	var (
		publicKey []byte
		wallet    int64
		uid       string
		err       error
		form      = new(loginForm)
	)

	if uid, err = getUID(r); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	if err = parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	client := getClient(r)
	logger := getLogger(r)

	if form.EcosystemID > 0 {
		client.EcosystemID = form.EcosystemID
	} else if client.EcosystemID == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Warning("state is empty, using 1 as a state")
		client.EcosystemID = 1
	}

	if len(form.KeyID) > 0 {
		wallet = converter.StringToAddress(form.KeyID)
	} else if len(form.PublicKey.Bytes()) > 0 {
		wallet = crypto.Address(form.PublicKey.Bytes())
	}

	account := &model.Key{}
	account.SetTablePrefix(client.EcosystemID)
	isAccount, err := account.Get(wallet)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		errorResponse(w, err)
		return
	}

	if isAccount {
		if account.Deleted == 1 {
			errorResponse(w, errDeletedKey)
			return
		}
		publicKey = account.PublicKey
	} else {
		if !allowCreateUser(client) {
			errorResponse(w, errKeyNotFound)
			return
		}

		publicKey = form.PublicKey.Bytes()
		if len(publicKey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			errorResponse(w, errEmptyPublic)
			return
		}

		nodePrivateKey, err := utils.GetNodePrivateKey()
		if err != nil || len(nodePrivateKey) < 1 {
			if err == nil {
				log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
			}

			errorResponse(w, err)
			return
		}

		contract := smart.GetContract("NewUser", 1)
		sc := tx.SmartContract{
			Header: tx.Header{
				ID:          int(contract.Block.Info.(*script.ContractInfo).ID),
				Time:        time.Now().Unix(),
				EcosystemID: 1,
				KeyID:       conf.Config.KeyID,
				NetworkID:   conf.Config.NetworkID,
			},
			Params: map[string]interface{}{
				"NewPubkey": hex.EncodeToString(publicKey),
			},
		}

		txData, txHash, err := tx.NewInternalTransaction(sc, nodePrivateKey)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ContractError, "err": err}).Error("Building transaction")
		} else {
			err = m.ContractRunner.RunContract(txData, txHash, sc.KeyID, logger)
		}
		if err != nil {
			errorResponse(w, err)
			return
		}
		ts := &model.LogTransaction{}
		for i := 0; i < 30; i++ {
			found, err := ts.GetByHash(txHash)
			if err != nil {
				errorResponse(w, err)
				return
			}
			if found && ts.Block > 0 {
				break
			}
			time.Sleep(time.Second)
		}
		if ts.Block == 0 {
			errorResponse(w, errNewUser)
			return
		}
	}

	if len(publicKey) == 0 {
		if client.EcosystemID > 1 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty, and state is not default")
			errorResponse(w, errStateLogin.Errorf(wallet, client.EcosystemID))
			return
		}

		if len(form.PublicKey.Bytes()) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			errorResponse(w, errEmptyPublic)
			return
		}
	}

	if form.RoleID != 0 && client.RoleID == 0 {
		checkedRole, err := checkRoleFromParam(form.RoleID, client.EcosystemID, wallet)
		if err != nil {
			errorResponse(w, err)
			return
		}

		if checkedRole != form.RoleID {
			errorResponse(w, errCheckRole)
			return
		}

		client.RoleID = checkedRole
	}

	verify, err := crypto.CheckSign(publicKey, []byte(nonceSalt()+uid), form.Signature.Bytes())
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "pubkey": publicKey, "uid": uid, "signature": form.Signature.Bytes()}).Error("checking signature")
		errorResponse(w, err)
		return
	}

	if !verify {
		logger.WithFields(log.Fields{"type": consts.InvalidObject, "pubkey": publicKey, "uid": uid, "signature": form.Signature.Bytes()}).Error("incorrect signature")
		errorResponse(w, errSignature)
		return
	}

	var (
		address = crypto.KeyToAddress(publicKey)
		sp      model.StateParameter
		founder int64
	)

	sp.SetTablePrefix(converter.Int64ToStr(client.EcosystemID))
	if ok, err := sp.Get(nil, "founder_account"); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting founder_account parameter")
		errorResponse(w, err)
		return
	} else if ok {
		founder = converter.StrToInt64(sp.Value)
	}

	result := &loginResult{
		EcosystemID: converter.Int64ToStr(client.EcosystemID),
		KeyID:       converter.Int64ToStr(wallet),
		Address:     address,
		IsOwner:     founder == wallet,
		IsNode:      conf.Config.KeyID == wallet,
		IsOBS:       conf.Config.IsSupportingOBS(),
	}

	claims := JWTClaims{
		KeyID:       result.KeyID,
		EcosystemID: result.EcosystemID,
		IsMobile:    form.IsMobile,
		RoleID:      converter.Int64ToStr(form.RoleID),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(form.Expire)).Unix(),
		},
	}

	result.Token, err = generateJWTToken(claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		errorResponse(w, err)
		return
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Hour * 30 * 24).Unix()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		errorResponse(w, err)
		return
	}
	result.NotifyKey, result.Timestamp, err = publisher.GetHMACSign(wallet)
	if err != nil {
		errorResponse(w, err)
		return
	}

	ra := &model.RolesParticipants{}
	roles, err := ra.SetTablePrefix(client.EcosystemID).GetActiveMemberRoles(wallet)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting roles")
		errorResponse(w, err)
		return
	}

	for _, r := range roles {
		var res map[string]string
		if err := json.Unmarshal([]byte(r.Role), &res); err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling role")
			errorResponse(w, err)
			return
		} else {
			result.Roles = append(result.Roles, rolesResult{RoleId: converter.StrToInt64(res["id"]), RoleName: res["name"]})
		}
	}

	jsonResponse(w, result)
}

func getUID(r *http.Request) (string, error) {
	var uid string

	token := getToken(r)
	if token != nil {
		if claims, ok := token.Claims.(*JWTClaims); ok {
			uid = claims.UID
		}
	} else if len(uid) == 0 {
		logger := getLogger(r)
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("UID is empty")
		return "", errUnknownUID
	}

	return uid, nil
}

func checkRoleFromParam(role, ecosystemID, wallet int64) (int64, error) {
	if role > 0 {
		ok, err := model.MemberHasRole(nil, ecosystemID, wallet, role)
		if err != nil {
			log.WithFields(log.Fields{
				"type":      consts.DBError,
				"member":    wallet,
				"role":      role,
				"ecosystem": ecosystemID}).Error("check role")

			return 0, err
		}

		if !ok {
			log.WithFields(log.Fields{
				"type":      consts.NotFound,
				"member":    wallet,
				"role":      role,
				"ecosystem": ecosystemID,
			}).Error("member hasn't role")

			return 0, nil
		}
	}
	return role, nil
}

func allowCreateUser(c *Client) bool {
	if conf.Config.IsSupportingOBS() {
		return true
	}

	return syspar.IsTestMode() && c.EcosystemID == 1
}
