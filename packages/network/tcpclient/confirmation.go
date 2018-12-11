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
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/network"
	log "github.com/sirupsen/logrus"
)

func CheckConfirmation(host string, blockHash []byte, logger *log.Entry) (hash string) {
	conn, err := newConnection(host)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": host, "block_hash": blockHash}).Debug("dialing to host")
		return "0"
	}
	defer conn.Close()

	rt := &network.RequestType{Type: network.RequestTypeConfirmation}
	if err = rt.Write(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("sending request type")
		return "0"
	}

	req := &network.ConfirmRequest{
		BlockHash: blockHash,
	}
	if err = req.Write(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("sending confirmation request")
		return "0"
	}

	resp := &network.ConfirmResponse{}

	if err := resp.Read(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("receiving confirmation response")
		return "0"
	}
	return string(converter.BinToHex(resp.Hash))
}
