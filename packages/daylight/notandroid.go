// +build !android,!ios
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

package daylight

import (
	"net"
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"

	log "github.com/sirupsen/logrus"
)

func httpListener(ListenHTTPHost string, route http.Handler) {
	l, err := net.Listen("tcp", ListenHTTPHost)
	log.WithFields(log.Fields{"host": ListenHTTPHost, "type": consts.NetworkError}).Debug("trying to listen at")
	if err == nil {
		log.WithFields(log.Fields{"host": ListenHTTPHost}).Info("listening at")
	} else {
		log.WithFields(log.Fields{"host": ListenHTTPHost, "error": err, "type": consts.NetworkError}).Debug("cannot listen at host")
	}

	go func() {
		srv := &http.Server{Handler: route}
		err = srv.Serve(l)
		if err != nil {
			log.WithFields(log.Fields{"host": ListenHTTPHost, "error": err, "type": consts.NetworkError}).Fatal("serving http at host")
			panic(err)
		}
	}()
}
