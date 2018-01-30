package tokenMovementMonitor

import (
	"fmt"

	"net/smtp"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/model"
	log "github.com/sirupsen/logrus"
)

const (
	networkPerDayLimit            = 100000000
	fromToPerDayLimit             = 10000
	tokenMovementQtyPerBlockLimit = 100

	networkPerDayMsgTemplate      = "day APL movement volume =  %f"
	fromToDayLimitMsgTemplate     = "from %d to %d sended volume = %f"
	perBlockTokenMovementTemplate = "from wallet %d token movement count = %d in block: %d"
)

func commonTokenMovementPerDay(conf conf.TokenMovementConfig) error {
	query := `SELECT SUM(amount) sum_amount 
	FROM "1_history" 
	WHERE created_at > NOW() - interval '24 hours'
	AND amount > 0`

	var qty float64
	if err := model.GetDB(nil).Raw(query).Row().Scan(&qty); err != nil {
		return err
	}

	if qty > networkPerDayLimit {
		sendEmail(conf, fmt.Sprintf(networkPerDayMsgTemplate, qty))
	}

	return nil
}

func fromToTokenMovementPerDay(conf conf.TokenMovementConfig) error {
	query := `SELECT sender_id, recipient_id, SUM(amount) sum_amount 
	FROM "1_history" 
	WHERE created_at > NOW() - interval '24 hours'
	AND amount > 0
	GROUP BY sender_id, recipient_id
	HAVING SUM(amount) > ?`

	rows, err := model.GetDB(nil).Raw(query, fromToPerDayLimit).Rows()
	if err != nil {
		return err
	}

	var (
		sender    int64
		recipient int64
		amount    float64
	)

	for rows.Next() {
		if err := rows.Scan(&sender, &recipient, &amount); err != nil {
			return err
		}

		sendEmail(conf, fmt.Sprintf(fromToDayLimitMsgTemplate, sender, recipient, amount))
	}

	return nil
}

func tokenMovementQtyPerBlock(conf conf.TokenMovementConfig, blockID int64) error {
	query := `SELECT sender_id, count(*)
	FROM "1_history" 
	WHERE block_id = ? AND amount > 0
	GROUP BY sender_id
	HAVING count(*) > ?`

	rows, err := model.GetDB(nil).Raw(query, blockID, tokenMovementQtyPerBlockLimit).Rows()
	if err != nil {
		return err
	}

	var (
		sender int64
		qty    int64
	)

	for rows.Next() {
		if err := rows.Scan(&sender, &qty); err != nil {
			return err
		}

		sendEmail(conf, fmt.Sprintf(perBlockTokenMovementTemplate, sender, qty, blockID))
	}

	return nil
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
func CheckTokenMovementLimits(conf conf.TokenMovementConfig, blockID int64) {
	if err := commonTokenMovementPerDay(conf); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("check common token movement")
	}

	if err := fromToTokenMovementPerDay(conf); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("check from to token movement")
	}

	if err := tokenMovementQtyPerBlock(conf, blockID); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("check token movement per block")
	}
}
