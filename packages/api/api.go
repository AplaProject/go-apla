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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/service"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/statsd"
)

const (
	jwtPrefix    = "Bearer "
	jwtExpire    = 36000  // By default, seconds
	multipartBuf = 100000 // the buffer size for ParseMultipartForm
)

type apiData struct {
	status        int
	result        interface{}
	params        map[string]interface{}
	ecosystemId   int64
	ecosystemName string
	keyId         int64
	roleId        int64
	isMobile      string
	vde           bool
	vm            *script.VM
	token         *jwt.Token
}

// ParamString reaturs string value of the api params
func (a *apiData) ParamString(key string) string {
	v, ok := a.params[key]
	if !ok {
		return ""
	}
	return v.(string)
}

// ParamInt64 reaturs int64 value of the api params
func (a *apiData) ParamInt64(key string) int64 {
	v, ok := a.params[key]
	if !ok {
		return 0
	}
	return v.(int64)
}

type forSign struct {
	Time    string `json:"time"`
	ForSign string `json:"forsign"`
}

type hashTx struct {
	Hash string `json:"hash"`
}

const (
	pInt64 = iota
	pHex
	pString

	pOptional = 0x100
)

type apiHandle func(http.ResponseWriter, *http.Request, *apiData, *log.Entry) error

func errorAPI(w http.ResponseWriter, err interface{}, code int, params ...interface{}) error {
	var (
		msg, errCode, errParams string
	)

	switch v := err.(type) {
	case string:
		errCode = v
		if val, ok := apiErrors[v]; ok {
			if len(params) > 0 {
				list := make([]string, 0)
				msg = fmt.Sprintf(val, params...)
				for _, item := range params {
					list = append(list, fmt.Sprintf(`"%v"`, item))
				}
				errParams = fmt.Sprintf(`, "params": [%s]`, strings.Join(list, `,`))
			} else {
				msg = val
			}
		} else {
			msg = v
		}
	case interface{}:
		errCode = `E_SERVER`
		if reflect.TypeOf(v).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			msg = v.(error).Error()
		}
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(`{"error": %q, "msg": %q %s}`, errCode, msg, errParams))
	return fmt.Errorf(msg)
}

func getPrefix(data *apiData) string {
	return converter.Int64ToStr(data.ecosystemId)
}

// DefaultHandler is a common handle function for api requests
func DefaultHandler(method, pattern string, params map[string]int, handlers ...apiHandle) hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		r.ParseMultipartForm(multipartBuf)
		counterName := statsd.APIRouteCounterName(method, pattern)
		statsd.Client.Inc(counterName+statsd.Count, 1, 1.0)
		startTime := time.Now()
		var (
			err  error
			data = &apiData{ecosystemId: 1}
		)
		requestLogger := log.WithFields(log.Fields{"headers": r.Header, "path": r.URL.Path, "protocol": r.Proto, "remote": r.RemoteAddr})
		requestLogger.Info("received http request")

		defer func() {
			endTime := time.Now()
			statsd.Client.TimingDuration(counterName+statsd.Time, endTime.Sub(startTime), 1.0)
			if r := recover(); r != nil {
				requestLogger.WithFields(log.Fields{"type": consts.PanicRecoveredError, "error": r, "stack": string(debug.Stack())}).Error("panic recovered error")
				fmt.Println("API Recovered", fmt.Sprintf("%s: %s", r, debug.Stack()))
				errorAPI(w, `E_RECOVERED`, http.StatusInternalServerError)
			}
		}()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		data.params = make(map[string]interface{})
		for _, par := range ps {
			data.params[par.Key] = par.Value
		}

		ihandlers := append([]apiHandle{
			fillToken,
			fillParams(params),
		}, handlers...)

		for _, handler := range ihandlers {
			if handler(w, r, data, requestLogger) != nil {
				return
			}
		}

		jsonResult, err := json.Marshal(data.result)
		if err != nil {
			requestLogger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marhsalling http response to json")
			errorAPI(w, err, http.StatusInternalServerError)
			return
		}

		w.Write(jsonResult)
	})
}

func fillToken(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	token, err := jwtToken(r)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("starting session")
		errmsg := err.Error()
		expired := `token is expired by`
		if strings.HasPrefix(errmsg, expired) {
			return errorAPI(w, `E_TOKENEXPIRED`, http.StatusUnauthorized, errmsg[len(expired):])
		}
		return errorAPI(w, err, http.StatusBadRequest)
	}

	data.token = token
	if token != nil && token.Valid {
		if claims, ok := token.Claims.(*JWTClaims); ok && len(claims.KeyID) > 0 {
			if err := fillTokenData(data, claims, logger); err != nil {
				return errorAPI(w, "E_SERVER", http.StatusNotFound, err)
			}
		}
	}

	return nil
}

func fillParams(params map[string]int) apiHandle {
	return func(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
		if conf.Config.IsSupportingVDE() {
			data.vde = true
		}

		data.vm = smart.GetVM()

		for key, par := range params {
			val := r.FormValue(key)
			if par&pOptional == 0 && len(val) == 0 {
				logger.WithFields(log.Fields{"type": consts.RouteError, "error": fmt.Sprintf("undefined val %s", key)}).Error("undefined val")
				return errorAPI(w, `E_UNDEFINEVAL`, http.StatusBadRequest, key)
			}
			switch par & 0xff {
			case pInt64:
				data.params[key] = converter.StrToInt64(val)
			case pHex:
				bin, err := hex.DecodeString(val)
				if err != nil {
					logger.WithFields(log.Fields{"type": consts.ConversionError, "value": val, "error": err}).Error("decoding http parameter from hex")
					return errorAPI(w, err, http.StatusBadRequest)
				}
				data.params[key] = bin
			case pString:
				data.params[key] = val
			}
		}

		return nil
	}
}

func checkEcosystem(w http.ResponseWriter, data *apiData, logger *log.Entry) (int64, string, error) {
	if conf.Config.IsSupportingVDE() {
		return consts.DefaultVDE, "1", nil
	}

	ecosystemID := data.ecosystemId
	if data.params[`ecosystem`].(int64) > 0 {
		ecosystemID = data.params[`ecosystem`].(int64)
		count, err := model.GetNextID(nil, "1_ecosystems")
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id ecosystems")
			return 0, ``, errorAPI(w, err, http.StatusBadRequest)
		}
		if ecosystemID >= count {
			logger.WithFields(log.Fields{"state_id": ecosystemID, "count": count, "type": consts.ParameterExceeded}).Error("state_id is larger then max count")
			return 0, ``, errorAPI(w, `E_ECOSYSTEM`, http.StatusBadRequest, ecosystemID)
		}
	}
	prefix := converter.Int64ToStr(ecosystemID)

	return ecosystemID, prefix, nil
}

func fillTokenData(data *apiData, claims *JWTClaims, logger *log.Entry) error {
	data.ecosystemId = converter.StrToInt64(claims.EcosystemID)
	data.keyId = converter.StrToInt64(claims.KeyID)
	data.isMobile = claims.IsMobile
	data.roleId = converter.StrToInt64(claims.RoleID)
	if !conf.Config.IsSupportingVDE() {
		ecosystem := &model.Ecosystem{}
		found, err := ecosystem.Get(data.ecosystemId)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting ecosystem from db")
			return err
		}

		if !found {
			err := fmt.Errorf("ecosystem not found")
			logger.WithFields(log.Fields{"type": consts.NotFound, "id": data.ecosystemId, "error": err}).Error("ecosystem not found")
		}

		data.ecosystemName = ecosystem.Name
	}
	return nil
}

func blockchainUpdatingState(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var reason string

	switch service.NodePauseType() {
	case service.NoPause:
		return nil
	case service.PauseTypeUpdatingBlockchain:
		reason = "E_UPDATING"
	case service.PauseTypeStopingNetwork:
		reason = "E_STOPPING"
	}

	return errorAPI(w, reason, http.StatusServiceUnavailable)
}
