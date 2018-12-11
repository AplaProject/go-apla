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

package custom

import (
	"errors"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/service"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

var (
	messageNetworkStopping = "Attention! The network is stopped!"

	ErrNetworkStopping = errors.New("Network is stopping")
)

type StopNetworkTransaction struct {
	Logger *log.Entry
	Data   interface{}

	Cert *utils.Cert
}

func (t *StopNetworkTransaction) Init() error {
	return nil
}

func (t *StopNetworkTransaction) Validate() error {
	if err := t.validate(); err != nil {
		t.Logger.WithError(err).Error("validating tx")
		return err
	}

	return nil
}

func (t *StopNetworkTransaction) validate() error {
	data := t.Data.(*consts.StopNetwork)
	cert, err := utils.ParseCert(data.StopNetworkCert)
	if err != nil {
		return err
	}

	fbdata, err := syspar.GetFirstBlockData()
	if err != nil {
		return err
	}

	if err = cert.Validate(fbdata.StopNetworkCertBundle); err != nil {
		return err
	}

	t.Cert = cert
	return nil
}

func (t *StopNetworkTransaction) Action() error {
	// Allow execute transaction, if the certificate was used
	if t.Cert.EqualBytes(consts.UsedStopNetworkCerts...) {
		return nil
	}

	// Set the node in a pause state
	service.PauseNodeActivity(service.PauseTypeStopingNetwork)

	t.Logger.Warn(messageNetworkStopping)
	return ErrNetworkStopping
}

func (t *StopNetworkTransaction) Rollback() error {
	return nil
}

func (t StopNetworkTransaction) Header() *blockchain.TxHeader {
	return nil
}
