package notificator

import (
	"fmt"

	"net/smtp"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
)

const (
	networkPerDayLimit            = 100000000
	networkPerDayMsgTemplate      = "day APL movement volume =  %f"
	fromToDayLimitMsgTemplate     = "from %d to %d sended volume = %f"
	perBlockTokenMovementTemplate = "from wallet %d token movement count = %f in block: %d"
)

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
func CheckTokenMovementLimits(conf conf.TokenMovementConfig, blockID int64) {

	amount, err := model.GetExcessCommonTokenMovementPerDay()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("check common token movement")
	} else if amount > networkPerDayLimit {
		msg := fmt.Sprintf(networkPerDayMsgTemplate, amount)
		sendEmail(conf, msg)
	}

	transfers, err := model.GetExcessFromToTokenMovementPerDay()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("check from to token movement")
	} else {
		for _, transfer := range transfers {
			msg := fmt.Sprintf(fromToDayLimitMsgTemplate, transfer.SenderID, transfer.RecipientID, transfer.Amount)
			sendEmail(conf, msg)
		}
	}

	transfers, err = model.GetExcessTokenMovementQtyPerBlock(blockID)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("check token movement per block")
	} else {
		for _, transfer := range transfers {
			msg := fmt.Sprintf(perBlockTokenMovementTemplate, transfer.SenderID, transfer.Amount, blockID)
			sendEmail(conf, msg)
		}
	}
}
