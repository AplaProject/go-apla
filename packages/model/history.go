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

import (
	"errors"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/shopspring/decimal"
)

var errLowBalance = errors.New("not enough APL on the balance")

// History represent record of history table
type History struct {
	ecosystem   int64
	ID          int64
	SenderID    int64
	RecipientID int64
	Amount      decimal.Decimal
	Comment     string
	BlockID     int64
	TxHash      []byte `gorm:"column:txhash"`
	CreatedAt   time.Time
}

// SetTablePrefix is setting table prefix
func (h *History) SetTablePrefix(prefix int64) *History {
	h.ecosystem = prefix
	return h
}

// TableName returns table name
func (h *History) TableName() string {
	if h.ecosystem == 0 {
		h.ecosystem = 1
	}
	return `1_history`
}

// APLTransfer from to amount
type APLTransfer struct {
	SenderID    int64
	RecipientID int64
	Amount      decimal.Decimal
}

//APLSenderTxCount struct to scan query result
type APLSenderTxCount struct {
	SenderID int64
	TxCount  int64
}

// GetExcessCommonTokenMovementPerDay returns sum of amounts 24 hours
func GetExcessCommonTokenMovementPerDay(tx *DbTransaction) (amount decimal.Decimal, err error) {
	db := GetDB(tx)
	type result struct {
		Amount decimal.Decimal
	}

	var res result
	err = db.Table("1_history").Select("SUM(amount) as amount").
		Where("to_timestamp(created_at) > NOW() - interval '24 hours' AND amount > 0").Scan(&res).Error

	return res.Amount, err
}

// GetExcessFromToTokenMovementPerDay returns from to pairs where sum of amount greather than fromToPerDayLimit per 24 hours
func GetExcessFromToTokenMovementPerDay(tx *DbTransaction) (excess []APLTransfer, err error) {
	db := GetDB(tx)
	err = db.Table("1_history").
		Select("sender_id, recipient_id, SUM(amount) amount").
		Where("to_timestamp(created_at) > NOW() - interval '24 hours' AND amount > 0").
		Group("sender_id, recipient_id").
		Having("SUM(amount) > ?", consts.FromToPerDayLimit).
		Scan(&excess).Error

	return excess, err
}

// GetExcessTokenMovementQtyPerBlock returns from to pairs where APL transactions count greather than tokenMovementQtyPerBlockLimit per 24 hours
func GetExcessTokenMovementQtyPerBlock(tx *DbTransaction, blockID int64) (excess []APLSenderTxCount, err error) {
	db := GetDB(tx)
	err = db.Table("1_history").
		Select("sender_id, count(*) tx_count").
		Where("block_id = ? AND amount > ?", blockID, 0).
		Group("sender_id").
		Having("count(*) > ?", consts.TokenMovementQtyPerBlockLimit).
		Scan(&excess).Error

	return excess, err
}
