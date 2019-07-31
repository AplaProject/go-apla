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

package model

// TransactionsAttempts is model
type TransactionsAttempts struct {
	Hash    []byte `gorm:"primary_key;not null"`
	Attempt int8   `gorm:"not null"`
}

// GetByHash returns TransactionsAttempts existence by hash
func (ta *TransactionsAttempts) GetByHash() (bool, error) {
	return isFound(DBConn.Where("hash = ?", ta.Hash).First(ta))
}

// IncrementTxAttemptCount increases attempt column
func IncrementTxAttemptCount(transactionHash []byte) (int64, error) {
	ta := &TransactionsAttempts{
		Hash: transactionHash,
	}

	found, err := ta.GetByHash()
	if err != nil {
		return 0, err
	}
	if found {
		err = DBConn.Exec("update transactions_attempts set attempt=attempt+1 where hash = ?",
			transactionHash).Error
		if err != nil {
			return 0, err
		}
		ta.Attempt++
	} else {
		ta.Hash = transactionHash
		ta.Attempt = 1
		if err = DBConn.Create(ta).Error; err != nil {
			return 0, err
		}
	}
	return int64(ta.Attempt), nil
}

func DecrementTxAttemptCount(transactionHash []byte) error {
	return DBConn.Exec("update transactions_attempts set attempt=attempt-1 where hash = ?",
		transactionHash).Error
}
