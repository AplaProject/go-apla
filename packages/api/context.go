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
	"context"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type contextKey int

const (
	contextKeyLogger contextKey = iota
	contextKeyToken
	contextKeyClient
)

func setContext(r *http.Request, key, value interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, value))
}

func getContext(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

func setLogger(r *http.Request, log *log.Entry) *http.Request {
	return setContext(r, contextKeyLogger, log)
}

func getLogger(r *http.Request) *log.Entry {
	if v := getContext(r, contextKeyLogger); v != nil {
		if v == nil {
			v = loggerFromRequest(r)
		}
		return v.(*log.Entry)
	}
	return nil
}

func setToken(r *http.Request, token *jwt.Token) *http.Request {
	return setContext(r, contextKeyToken, token)
}

func getToken(r *http.Request) *jwt.Token {
	if v := getContext(r, contextKeyToken); v != nil {
		return v.(*jwt.Token)
	}
	return nil
}

func setClient(r *http.Request, client *Client) *http.Request {
	return setContext(r, contextKeyClient, client)
}

func getClient(r *http.Request) *Client {
	if v := getContext(r, contextKeyClient); v != nil {
		return v.(*Client)
	}
	return nil
}
