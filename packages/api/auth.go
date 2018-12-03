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
	"fmt"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

var (
	jwtSecret = crypto.RandSeq(15)
)

// JWTClaims is storing jwt claims
type JWTClaims struct {
	UID         string `json:"uid,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	RoleID      string `json:"role_id,omitempty"`
	IsMobile    string `json:"is_mobile,omitempty"`
	jwt.StandardClaims
}

func jwtToken(r *http.Request) (*jwt.Token, error) {
	auth := r.Header.Get(`Authorization`)
	if len(auth) == 0 {
		return nil, nil
	}
	if strings.HasPrefix(auth, jwtPrefix) {
		auth = auth[len(jwtPrefix):]
	} else {
		return nil, fmt.Errorf(`wrong authorization value`)
	}
	return jwt.ParseWithClaims(auth, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
}

func jwtGenerateToken(w http.ResponseWriter, claims JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func authWallet(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	if data.keyId == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("wallet is empty")
		return errorAPI(w, `E_UNAUTHORIZED`, http.StatusUnauthorized)
	}
	return nil
}

func authState(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	if data.keyId == 0 || data.ecosystemId <= 1 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("state is empty")
		return errorAPI(w, `E_UNAUTHORIZED`, http.StatusUnauthorized)
	}
	return nil
}
