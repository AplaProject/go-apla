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
	"errors"
	"fmt"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

var (
	jwtSecret = []byte(crypto.RandSeq(15))
	jwtPrefix = "Bearer "
	jwtExpire = 36000 // By default, seconds

	errJWTAuthValue      = errors.New("wrong authorization value")
	errEcosystemNotFound = errors.New("ecosystem not found")
)

// JWTClaims is storing jwt claims
type JWTClaims struct {
	UID         string `json:"uid,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	RoleID      string `json:"role_id,omitempty"`
	IsMobile    bool   `json:"is_mobile,omitempty"`
	jwt.StandardClaims
}

func generateJWTToken(claims JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func parseJWTToken(header string) (*jwt.Token, error) {
	if len(header) == 0 {
		return nil, nil
	}

	if strings.HasPrefix(header, jwtPrefix) {
		header = header[len(jwtPrefix):]
	} else {
		return nil, errJWTAuthValue
	}

	return jwt.ParseWithClaims(header, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
}

func getClientFromToken(token *jwt.Token) (*Client, error) {
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, nil
	}
	if len(claims.KeyID) == 0 {
		return nil, nil
	}

	client := &Client{
		EcosystemID: converter.StrToInt64(claims.EcosystemID),
		KeyID:       converter.StrToInt64(claims.KeyID),
		IsMobile:    claims.IsMobile,
		RoleID:      converter.StrToInt64(claims.RoleID),
	}

	ecosystem := &model.Ecosystem{}
	found, err := ecosystem.Get(client.EcosystemID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting ecosystem from db")
		return nil, err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "id": client.EcosystemID, "error": errEcosystemNotFound}).Error("ecosystem not found")
		return nil, err
	}

	client.EcosystemName = ecosystem.Name

	return client, nil
}
