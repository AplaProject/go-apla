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
	"math/rand"
	"net/http"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

const jwtUIDExpire = time.Second * 5

type getUIDResult struct {
	UID         string `json:"uid,omitempty"`
	Token       string `json:"token,omitempty"`
	Expire      string `json:"expire,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	Address     string `json:"address,omitempty"`
	NetworkID   string `json:"network,omitempty"`
}

func getUIDHandler(w http.ResponseWriter, r *http.Request) {
	result := new(getUIDResult)
	result.NetworkID = converter.Int64ToStr(conf.Config.NetworkID)
	token := getToken(r)
	if token != nil {
		if claims, ok := token.Claims.(*JWTClaims); ok && len(claims.KeyID) > 0 {
			result.EcosystemID = claims.EcosystemID
			result.Expire = converter.Int64ToStr(claims.ExpiresAt - time.Now().Unix())
			result.KeyID = claims.KeyID
			jsonResponse(w, result)
			return
		}
	}

	result.UID = converter.Int64ToStr(rand.New(rand.NewSource(time.Now().Unix())).Int63())
	claims := JWTClaims{
		UID:         result.UID,
		EcosystemID: "1",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(jwtUIDExpire).Unix(),
		},
	}

	var err error
	if result.Token, err = generateJWTToken(claims); err != nil {
		logger := getLogger(r)
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		errorResponse(w, err)
		return
	}

	jsonResponse(w, result)
}
