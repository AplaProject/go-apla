package publisher

import (
	"encoding/hex"
	"strconv"
	"time"

	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/centrifugal/gocent"
	log "github.com/sirupsen/logrus"
)

var (
	clientsChannels   = map[int64]string{}
	centrifugoSecret  = ""
	centrifugoURL     = ""
	centrifugoTimeout = time.Second * 5
	publisher         *gocent.Client
)

func init() {
	err := config.Read()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConfigError, "errror": err}).Error("reading config")
	}
	centrifugoSecret = config.ConfigIni["centrifugo_secret"]
	centrifugoURL = config.ConfigIni["centrifugo_url"]
	publisher = gocent.NewClient(centrifugoURL, centrifugoSecret, centrifugoTimeout)
}

func GetHMACSign(userID int64) (string, error) {
	secret, err := crypto.GetHMAC(centrifugoSecret, strconv.FormatInt(userID, 10))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("HMAC getting error")
		return "", err
	}
	result := hex.EncodeToString(secret)
	clientsChannels[userID] = result
	return result, nil
}

func Write(userID int64, data string) (bool, error) {
	return publisher.Publish("client#"+strconv.FormatInt(userID, 10), []byte(data))
}
