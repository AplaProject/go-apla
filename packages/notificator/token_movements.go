// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package notificator

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"net/smtp"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	log "github.com/sirupsen/logrus"
)

const (
	networkPerDayLimit            = 100000000
	networkPerDayMsgTemplate      = "day APL movement volume =  %s"
	fromToDayLimitMsgTemplate     = "from %d to %d sended volume = %s"
	perBlockTokenMovementTemplate = "from wallet %d token movement count = %d in block: %d"

	networkPerDayEvent         = 1
	fromToDayLimitEvent        = 2
	perBlockTokenMovementEvent = 3
)

var lastLimitEvents map[uint8]time.Time

func init() {
	lastLimitEvents = make(map[uint8]time.Time, 0)
}

func sendEmail(conf conf.TokenMovementConfig, message string) error {
	auth := smtp.PlainAuth("", conf.Username, conf.Password, conf.Host)
	to := []string{conf.To}
	msg := []byte(fmt.Sprintf("From: %s\r\n", conf.From) +
		fmt.Sprintf("To: %s\r\n", conf.To) +
		fmt.Sprintf("Subject: %s\r\n", conf.Subject) +
		"\r\n" +
		fmt.Sprintf("%s\r\n", message))
	err := smtp.SendMail(fmt.Sprintf("%s:%d", conf.Host, conf.Port), auth, conf.From, to, msg)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("sending email")
	}
	return err
}

// CheckTokenMovementLimits check all limits
func CheckTokenMovementLimits(tx *model.DbTransaction, conf conf.TokenMovementConfig, blockID int64) {
	var messages []string
	if needCheck(networkPerDayEvent) {
		amount, err := model.GetExcessCommonTokenMovementPerDay(tx)

		if err != nil {

			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("check common token movement")
		} else if amount.GreaterThanOrEqual(decimal.NewFromFloat(networkPerDayLimit)) {

			messages = append(messages, fmt.Sprintf(networkPerDayMsgTemplate, amount.String()))
			lastLimitEvents[networkPerDayEvent] = time.Now()
		}
	}

	if needCheck(fromToDayLimitEvent) {
		transfers, err := model.GetExcessFromToTokenMovementPerDay(tx)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("check from to token movement")
		} else {
			for _, transfer := range transfers {
				messages = append(messages, fmt.Sprintf(fromToDayLimitMsgTemplate, transfer.SenderID, transfer.RecipientID, transfer.Amount))
			}

			if len(transfers) > 0 {
				lastLimitEvents[fromToDayLimitEvent] = time.Now()
			}
		}
	}

	excesses, err := model.GetExcessTokenMovementQtyPerBlock(tx, blockID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("check token movement per block")
	} else {
		for _, excess := range excesses {
			messages = append(messages, fmt.Sprintf(perBlockTokenMovementTemplate, excess.SenderID, excess.TxCount, blockID))
		}
	}

	if len(messages) > 0 {
		sendEmail(conf, strings.Join(messages, "\n"))
	}
}

// checks needed only if we have'nt prevent events or if event older then 1 day
func needCheck(event uint8) bool {
	t, ok := lastLimitEvents[event]
	if !ok {
		return true
	}

	return time.Now().Sub(t) >= 24*time.Hour
}
