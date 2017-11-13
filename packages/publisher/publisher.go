package publisher

import (
	"encoding/hex"
	"strconv"
	"time"

	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/centrifugal/gocent"
)

var (
	clientsChannels   = map[int64]string{}
	centrifugoSecret  = ""
	centrifugoURL     = "http://localhost:8000"
	centrifugoTimeout = time.Second * 5
	publisher         *gocent.Client
)

func init() {
	err := config.Read()
	if err != nil {
		// TODO add logs
	}
	centrifugoSecret = config.ConfigIni["centrifugo_secret"]
	publisher = gocent.NewClient(centrifugoURL, centrifugoSecret, centrifugoTimeout)
}

func GetHMACSign(userID int64) (string, error) {
	secret, err := crypto.GetHMAC(centrifugoSecret, strconv.FormatInt(userID, 10))
	if err != nil {
		//TODO add logs
		return "", err
	}
	result := hex.EncodeToString(secret)
	clientsChannels[userID] = result
	return result, nil
}

func Write(userID int64, data string) (bool, error) {
	ok, err := publisher.Publish("client#"+strconv.FormatInt(userID, 10), []byte(data))
	if err != nil {
		// TODO add logs fmt.Println("err", err)
	}
	return ok, err
}
