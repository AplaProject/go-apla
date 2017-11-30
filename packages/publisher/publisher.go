package publisher

import (
	"encoding/hex"
	"strconv"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/centrifugal/gocent"
	log "github.com/sirupsen/logrus"
)

var (
	clientsChannels = map[int64]string{}
	// centrifugoSecret  = ""
	// centrifugoURL     = ""
	centrifugoTimeout = time.Second * 5
	publisher         *gocent.Client
)

// InitCentrifugo client
func InitCentrifugo(cfg conf.CentrifugoConfig) {
	publisher = gocent.NewClient(cfg.URL, cfg.Secret, centrifugoTimeout)
}

// GetHMACSign returns HMACS sign for userID
func GetHMACSign(userID int64) (string, error) {
	secret, err := crypto.GetHMAC(conf.Config.Centrifugo.Secret, strconv.FormatInt(userID, 10))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("HMAC getting error")
		return "", err
	}
	result := hex.EncodeToString(secret)
	clientsChannels[userID] = result
	return result, nil
}

// Write is publishing data to server
func Write(userID int64, data string) (bool, error) {
	return publisher.Publish("client#"+strconv.FormatInt(userID, 10), []byte(data))
}
