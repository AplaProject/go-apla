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

package daemons

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/queue"
	"github.com/AplaProject/go-apla/packages/smart"

	log "github.com/sirupsen/logrus"
)

const (
	callDelayedContract = "CallDelayedContract"
	firstEcosystemID    = 1
)

// DelayedTx represents struct which works with delayed contracts
type DelayedTx struct {
	logger     *log.Entry
	privateKey string
	publicKey  string
}

// RunForBlockID creates the transactions that need to be run for blockID
func (dtx *DelayedTx) RunForBlockID(blockID int64) {
	contracts, err := model.GetAllDelayedContractsForBlockID(blockID)
	if err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting delayed contracts for block")
		return
	}

	for _, c := range contracts {
		if err := dtx.createTx(c.Contract, c.ID, c.KeyID); err != nil {
			dtx.logger.WithFields(log.Fields{"error": err}).Debug("can't create transaction for delayed contract")
		}
	}
}

func (dtx *DelayedTx) createTx(name string, delayedContractID, keyID int64) error {
	params := map[string]string{"Id": converter.Int64ToStr(delayedContractID)}
	smartTx := &blockchain.Transaction{
		Header: blockchain.TxHeader{
			Name:        name,
			Time:        time.Now().Unix(),
			EcosystemID: firstEcosystemID,
			KeyID:       keyID,
			NetworkID:   consts.NETWORK_ID,
		},
		SignedBy: smart.PubToID(dtx.publicKey),
		Params:   params,
	}

	signature, err := crypto.Sign(
		[]byte(dtx.privateKey),
		[]byte(fmt.Sprintf("%s,%d", smartTx.ForSign(), delayedContractID)),
	)
	if err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing by node private key")
		return err
	}
	smartTx.Header.BinSignatures = converter.EncodeLengthPlusData(signature)

	if smartTx.Header.PublicKey, err = hex.DecodeString(dtx.publicKey); err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex")
		return err
	}

	if err := queue.ValidateTxQueue.Enqueue(smartTx); err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("calculating hash of smart contract")
		return err
	}
	return nil
}
