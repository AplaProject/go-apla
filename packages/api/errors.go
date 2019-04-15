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
	"net/http"
)

var (
	defaultStatus        = http.StatusBadRequest
	ErrEcosystemNotFound = errors.New("Ecosystem not found")
	errContract          = errType{"E_CONTRACT", "There is not %s contract", http.StatusNotFound}
	errDBNil             = errType{"E_DBNIL", "DB is nil", defaultStatus}
	errDeletedKey        = errType{"E_DELETEDKEY", "The key is deleted", http.StatusForbidden}
	errEcosystem         = errType{"E_ECOSYSTEM", "Ecosystem %d doesn't exist", defaultStatus}
	errEmptyPublic       = errType{"E_EMPTYPUBLIC", "Public key is undefined", http.StatusBadRequest}
	errKeyNotFound       = errType{"E_KEYNOTFOUND", "Key has not been found", http.StatusNotFound}
	errEmptySign         = errType{"E_EMPTYSIGN", "Signature is undefined", defaultStatus}
	errHashWrong         = errType{"E_HASHWRONG", "Hash is incorrect", http.StatusBadRequest}
	errHashNotFound      = errType{"E_HASHNOTFOUND", "Hash has not been found", defaultStatus}
	errHeavyPage         = errType{"E_HEAVYPAGE", "This page is heavy", defaultStatus}
	errInstalled         = errType{"E_INSTALLED", "Apla is already installed", defaultStatus}
	errInvalidWallet     = errType{"E_INVALIDWALLET", "Wallet %s is not valid", http.StatusBadRequest}
	errLimitForsign      = errType{"E_LIMITFORSIGN", "Length of forsign is too big (%d)", defaultStatus}
	errLimitTxSize       = errType{"E_LIMITTXSIZE", "The size of tx is too big (%d)", defaultStatus}
	errNotFound          = errType{"E_NOTFOUND", "Page not found", http.StatusNotFound}
	errParamNotFound     = errType{"E_PARAMNOTFOUND", "Parameter %s has not been found", http.StatusNotFound}
	errPermission        = errType{"E_PERMISSION", "Permission denied", http.StatusUnauthorized}
	errQuery             = errType{"E_QUERY", "DB query is wrong", http.StatusInternalServerError}
	errRecovered         = errType{"E_RECOVERED", "API recovered", http.StatusInternalServerError}
	errServer            = errType{"E_SERVER", "Server error", defaultStatus}
	errSignature         = errType{"E_SIGNATURE", "Signature is incorrect", http.StatusBadRequest}
	errUnknownSign       = errType{"E_UNKNOWNSIGN", "Unknown signature", defaultStatus}
	errStateLogin        = errType{"E_STATELOGIN", "%s is not a membership of ecosystem %s", http.StatusForbidden}
	errTableNotFound     = errType{"E_TABLENOTFOUND", "Table %s has not been found", http.StatusNotFound}
	errToken             = errType{"E_TOKEN", "Token is not valid", defaultStatus}
	errTokenExpired      = errType{"E_TOKENEXPIRED", "Token is expired by %s", http.StatusUnauthorized}
	errUnauthorized      = errType{"E_UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized}
	errUndefineval       = errType{"E_UNDEFINEVAL", "Value %s is undefined", defaultStatus}
	errUnknownUID        = errType{"E_UNKNOWNUID", "Unknown uid", defaultStatus}
	errOBS               = errType{"E_OBS", "Virtual Dedicated Ecosystem %d doesn't exist", defaultStatus}
	errOBSCreated        = errType{"E_OBSCREATED", "Virtual Dedicated Ecosystem is already created", http.StatusBadRequest}
	errRequestNotFound   = errType{"E_REQUESTNOTFOUND", "Request %s doesn't exist", defaultStatus}
	errUpdating          = errType{"E_UPDATING", "Node is updating blockchain", http.StatusServiceUnavailable}
	errStopping          = errType{"E_STOPPING", "Network is stopping", http.StatusServiceUnavailable}
	errNotImplemented    = errType{"E_NOTIMPLEMENTED", "Not implemented", http.StatusNotImplemented}
	errDiffKey           = errType{"E_DIFKEY", "Sender's key is different from tx key", defaultStatus}
	errBannded           = errType{"E_BANNED", "The key is banned till %s", http.StatusForbidden}
	errCheckRole         = errType{"E_CHECKROLE", "Access denied", http.StatusForbidden}
	errNewUser           = errType{"E_NEWUSER", "Can't create a new user", http.StatusUnauthorized}
)

type errType struct {
	Err     string `json:"error"`
	Message string `json:"msg"`
	Status  int    `json:"-"`
}

func (et errType) Error() string {
	return et.Err
}

func (et errType) Errorf(v ...interface{}) errType {
	et.Message = fmt.Sprintf(et.Message, v...)
	return et
}
