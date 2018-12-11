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
	"errors"
	"net"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network"
	"github.com/AplaProject/go-apla/packages/queue"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

var errStopCertAlreadyUsed = errors.New("Stop certificate is already used")

// Type3
func Type3(req *network.StopNetworkRequest, w net.Conn) error {
	hash, err := processStopNetwork(req.Data)
	if err != nil {
		return err
	}

	res := &network.StopNetworkResponse{hash}
	if err = res.Write(w); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.NetworkError}).Error("sending response")
		return err
	}

	return nil
}

func processStopNetwork(b []byte) ([]byte, error) {
	cert, err := utils.ParseCert(b)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ParseError}).Error("parsing cert")
		return nil, err
	}

	if cert.EqualBytes(consts.UsedStopNetworkCerts...) {
		log.WithFields(log.Fields{"error": errStopCertAlreadyUsed, "type": consts.InvalidObject}).Error("checking cert")
		return nil, errStopCertAlreadyUsed
	}

	fbdata, err := syspar.GetFirstBlockData()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ConfigError}).Error("getting data of first block")
		return nil, err
	}

	if err = cert.Validate(fbdata.StopNetworkCertBundle); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.InvalidObject}).Error("validating cert")
		return nil, err
	}
	smartTx, err := smart.CallContract("StopNetwork", 1, map[string]string{"Cert": string(b)}, []string{string(b)})
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ContractError}).Error("Executing contract")
		return nil, err
	}
	hash, err := smartTx.Hash()
	if err != nil {
		return nil, err
	}
	if err := queue.ValidateTxQueue.Enqueue(smartTx); err != nil {
		return nil, err
	}

	return hash, nil
}
