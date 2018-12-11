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

package block

import (
	"errors"
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/transaction"

	log "github.com/sirupsen/logrus"
)

const (
	letPreprocess = 0x0001 // checking before generating block
	letGenBlock   = 0x0002 // checking during generating block
	letParsing    = 0x0004 // common checking during parsing block
)

// Limits is used for saving current limit information
type Limits struct {
	Mode     int
	Block    *PlayableBlock // it equals nil if checking before generatin block
	Limiters []Limiter      // the list of limiters
}

// Limiter describes interface functions for limits
type Limiter interface {
	init(*PlayableBlock)
	check(*transaction.Transaction, int) error
}

type limiterModes struct {
	limiter Limiter
	modes   int // combination of letPreprocess letGenBlock letParsing
}

var (
	// ErrLimitSkip returns when tx should be skipped during generating block
	ErrLimitSkip = errors.New(`skip tx`)
	// ErrLimitStop returns when the generation of the block should be stopped
	ErrLimitStop = errors.New(`stop generating block`)
	// ErrLimitTime returns when the time limit exceeded
	ErrLimitTime = errors.New(`Time limit exceeded`)
)

// NewLimits initializes Limits structure.
func NewLimits(b *PlayableBlock) (limits *Limits) {
	limits = &Limits{Block: b, Limiters: make([]Limiter, 0, 8)}
	if b == nil {
		limits.Mode = letPreprocess
	} else if b.GenBlock {
		limits.Mode = letGenBlock
	} else {
		limits.Mode = letParsing
	}
	allLimiters := []limiterModes{
		{&txMaxSize{}, letPreprocess | letParsing},
		{&txUserLimit{}, letPreprocess | letParsing},
		{&txMaxLimit{}, letPreprocess | letParsing},
		{&txUserEcosysLimit{}, letPreprocess | letParsing},
		{&timeBlockLimit{}, letGenBlock},
		{&txMaxFuel{}, letGenBlock | letParsing},
	}
	for _, limiter := range allLimiters {
		if limiter.modes&limits.Mode == 0 {
			continue
		}
		limiter.limiter.init(b)
		limits.Limiters = append(limits.Limiters, limiter.limiter)
	}
	return
}

// CheckLimit calls each limiter
func (limits *Limits) CheckLimit(t *transaction.Transaction) error {
	for _, limiter := range limits.Limiters {
		if err := limiter.check(t, limits.Mode); err != nil {
			return err
		}
	}
	return nil
}

func limitError(limitName, msg string, args ...interface{}) error {
	err := fmt.Errorf(msg, args...)
	log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Error(limitName)
	return script.SetVMError(`panic`, err)
}

// Checking the max tx in the block
type txMaxLimit struct {
	Count int // the current count
	Limit int // max count of tx in the block
}

func (bl *txMaxLimit) init(b *PlayableBlock) {
	bl.Limit = syspar.GetMaxTxCount()
}

func (bl *txMaxLimit) check(t *transaction.Transaction, mode int) error {
	bl.Count++
	if bl.Count > bl.Limit {
		if mode == letPreprocess {
			return ErrLimitStop
		}
		return limitError(`txMaxLimit`, `Max tx in the block`)
	}
	return nil
}

// Checking the time of the start of generating block
type timeBlockLimit struct {
	Start time.Time     // the time of the start of generating block
	Limit time.Duration // the maximum time
}

func (bl *timeBlockLimit) init(b *PlayableBlock) {
	bl.Start = time.Now()
	bl.Limit = time.Millisecond * time.Duration(syspar.GetMaxBlockGenerationTime())
}

func (bl *timeBlockLimit) check(t *transaction.Transaction, mode int) error {
	if time.Since(bl.Start) < bl.Limit {
		return nil
	}

	if mode == letGenBlock {
		return ErrLimitStop
	}

	return limitError("txBlockTimeLimit", "Block generation time exceeded")
}

// Checking the max tx from one user in the block
type txUserLimit struct {
	TxUsers map[int64]int // the counter of tx from one user
	Limit   int           // the value of max tx from one user
}

func (bl *txUserLimit) init(b *PlayableBlock) {
	bl.TxUsers = make(map[int64]int)
	bl.Limit = syspar.GetMaxBlockUserTx()
}

func (bl *txUserLimit) check(t *transaction.Transaction, mode int) error {
	var (
		count int
		ok    bool
	)
	keyID := t.TxSmart.Header.KeyID
	if count, ok = bl.TxUsers[keyID]; ok {
		if count+1 > bl.Limit {
			if mode == letPreprocess {
				return ErrLimitSkip
			}
			return limitError(`txUserLimit`, `Max tx from one user %d`, keyID)
		}
	}
	bl.TxUsers[keyID] = count + 1
	return nil
}

// Checking the max tx from one user in the ecosystem contracts
type ecosysLimit struct {
	TxUsers map[int64]int // the counter of tx from one user in the ecosystem
	Limit   int           // the value of max tx from one user in the ecosystem
}

type txUserEcosysLimit struct {
	TxEcosys map[int64]ecosysLimit // the counter of tx from one user in ecosystems
}

func (bl *txUserEcosysLimit) init(b *PlayableBlock) {
	bl.TxEcosys = make(map[int64]ecosysLimit)
}

func (bl *txUserEcosysLimit) check(t *transaction.Transaction, mode int) error {
	keyID := t.TxSmart.Header.KeyID
	ecosystemID := t.TxSmart.Header.EcosystemID
	if val, ok := bl.TxEcosys[ecosystemID]; ok {
		if user, ok := val.TxUsers[keyID]; ok {
			if user+1 > val.Limit {
				if mode == letPreprocess {
					return ErrLimitSkip
				}
				return limitError(`txUserEcosysLimit`, `Max tx from one user %d in ecosystem %d`,
					keyID, ecosystemID)
			}
			val.TxUsers[keyID] = user + 1
		} else {
			val.TxUsers[keyID] = 1
		}
	} else {
		limit := syspar.GetMaxBlockUserTx()
		sp := &model.StateParameter{}
		sp.SetTablePrefix(converter.Int64ToStr(ecosystemID))
		found, err := sp.Get(t.DbTransaction, `max_tx_block_per_user`)
		if err != nil {
			return limitError(`txUserEcosysLimit`, err.Error())
		}
		if found {
			limit = converter.StrToInt(sp.Value)
		}
		bl.TxEcosys[ecosystemID] = ecosysLimit{TxUsers: make(map[int64]int), Limit: limit}
		bl.TxEcosys[ecosystemID].TxUsers[keyID] = 1
	}
	return nil
}

// Checking the max tx & block size
type txMaxSize struct {
	Size       int64 // the current size of the block
	LimitBlock int64 // max size of the block
	LimitTx    int64 // max size of tx
}

func (bl *txMaxSize) init(b *PlayableBlock) {
	bl.LimitBlock = syspar.GetMaxBlockSize()
	bl.LimitTx = syspar.GetMaxTxSize()
}

func (bl *txMaxSize) check(t *transaction.Transaction, mode int) error {
	size := int64(len(t.TxFullData))
	if size > bl.LimitTx {
		return limitError(`txMaxSize`, `Max size of tx`)
	}
	bl.Size += size
	if bl.Size > bl.LimitBlock {
		if mode == letPreprocess {
			return ErrLimitStop
		}
		return limitError(`txMaxSize`, `Max size of the block`)
	}
	return nil
}

// Checking the max tx & block size
type txMaxFuel struct {
	Fuel       int64 // the current fuel of the block
	LimitBlock int64 // max fuel of the block
	LimitTx    int64 // max fuel of tx
}

func (bl *txMaxFuel) init(b *PlayableBlock) {
	bl.LimitBlock = syspar.GetMaxBlockFuel()
	bl.LimitTx = syspar.GetMaxTxFuel()
}

func (bl *txMaxFuel) check(t *transaction.Transaction, mode int) error {
	fuel := t.TxFuel
	if fuel > bl.LimitTx {
		return limitError(`txMaxFuel`, `Max fuel of tx %d > %d`, fuel, bl.LimitTx)
	}
	bl.Fuel += fuel
	if bl.Fuel > bl.LimitBlock {
		if mode == letGenBlock {
			return ErrLimitStop
		}
		return limitError(`txMaxFuel`, `Max fuel of the block`)
	}
	return nil
}
