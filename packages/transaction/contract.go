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

package transaction

import (
	"bytes"
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils/tx"
)

const (
	errUnknownContract = `Cannot find %s contract`
)

func CreateContract(contractName string, keyID int64, params map[string]interface{},
	privateKey []byte) error {
	ecosysID, _ := converter.ParseName(contractName)
	if ecosysID == 0 {
		ecosysID = 1
	}
	contract := smart.GetContract(contractName, uint32(ecosysID))
	if contract == nil {
		return fmt.Errorf(errUnknownContract, contractName)
	}
	sc := tx.SmartContract{
		Header: tx.Header{
			ID:          int(contract.Block.Info.(*script.ContractInfo).ID),
			Time:        time.Now().Unix(),
			EcosystemID: ecosysID,
			KeyID:       keyID,
			NetworkID:   conf.Config.NetworkID,
		},
		Params: params,
	}
	txData, _, err := tx.NewTransaction(sc, privateKey)
	if err == nil {
		rtx := &RawTransaction{}
		if err = rtx.Unmarshall(bytes.NewBuffer(txData)); err == nil {
			err = model.SendTx(rtx, sc.KeyID)
		}
	}
	return err
}
