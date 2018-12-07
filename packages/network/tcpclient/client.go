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

package tcpclient

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	log "github.com/sirupsen/logrus"
)

var wrongAddressError = errors.New("Wrong address")

// NormalizeHostAddress get address. if port not defined returns combined string with ip and defaultPort
func NormalizeHostAddress(address string, defaultPort int64) (string, error) {

	_, _, err := net.SplitHostPort(address)
	if err != nil {
		if strings.HasSuffix(err.Error(), "missing port in address") {
			return fmt.Sprintf("%s:%d", address, defaultPort), nil
		}

		return "", err
	}

	return address, nil
}

func newConnection(addr string) (net.Conn, error) {
	if len(addr) == 0 {
		return nil, wrongAddressError
	}

	host, err := NormalizeHostAddress(addr, consts.DEFAULT_TCP_PORT)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "host": addr, "error": err}).Error("on normalize host address")
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", host, consts.TCPConnTimeout)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "address": host}).Debug("dialing tcp")
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(consts.READ_TIMEOUT * time.Second))
	conn.SetWriteDeadline(time.Now().Add(consts.WRITE_TIMEOUT * time.Second))
	return conn, nil
}
