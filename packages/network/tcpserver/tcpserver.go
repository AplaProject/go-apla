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

package tcpserver

import (
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network"
	"github.com/AplaProject/go-apla/packages/service"

	log "github.com/sirupsen/logrus"
)

var (
	counter int64
)

// HandleTCPRequest proceed TCP requests
func HandleTCPRequest(rw net.Conn) {
	defer func() {
		atomic.AddInt64(&counter, -1)
	}()

	count := atomic.AddInt64(&counter, +1)
	if count > 20 {
		return
	}

	dType := &network.RequestType{}
	err := dType.Read(rw)
	if err != nil {
		log.Errorf("read request type failed: %s", err)
		return
	}

	log.WithFields(log.Fields{"request_type": dType.Type}).Debug("tcpserver got request type")
	var response interface{}

	switch dType.Type {
	case network.RequestTypeFullNode:
		if service.IsNodePaused() {
			return
		}
		err = Type1(rw)

	case network.RequestTypeNotFullNode:
		if service.IsNodePaused() {
			return
		}
		response, err = Type2(rw)

	case network.RequestTypeStopNetwork:
		req := &network.StopNetworkRequest{}
		if err = req.Read(rw); err == nil {
			err = Type3(req, rw)
		}

	case network.RequestTypeConfirmation:
		if service.IsNodePaused() {
			return
		}

		req := &network.ConfirmRequest{}
		if err = req.Read(rw); err == nil {
			response, err = Type4(req)
		}

	case network.RequestTypeBlockCollection:
		req := &network.GetBodiesRequest{}
		if err = req.Read(rw); err == nil {
			err = Type7(req, rw)
		}

	case network.RequestTypeMaxBlock:
		response, err = Type10()
	}

	if err != nil || response == nil {
		return
	}

	log.WithFields(log.Fields{"response": response, "request_type": dType.Type}).Debug("tcpserver responded")
	if err = response.(network.SelfReaderWriter).Write(rw); err != nil {
		// err = SendRequest(response, rw)
		log.Errorf("tcpserver handle error: %s", err)
	}
}

// TcpListener is listening tcp address
func TcpListener(laddr string) error {

	if strings.HasPrefix(laddr, "127.") {
		log.Warn("Listening at local address: ", laddr)
	}

	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": laddr}).Error("Error listening")
		return err
	}

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": laddr}).Error("Error accepting")
				time.Sleep(time.Second)
			} else {
				go func(conn net.Conn) {
					HandleTCPRequest(conn)
					conn.Close()
				}(conn)
			}
		}
	}()

	return nil
}
