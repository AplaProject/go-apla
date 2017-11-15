package publisher

import (
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/centrifugal/gocent"
	log "github.com/sirupsen/logrus"
)

type ClientsChannels struct {
	storage map[int64]string
	sync.RWMutex
}

func (cn *ClientsChannels) Set(id int64, s string) {
	cn.Lock()
	defer cn.Unlock()
	cn.storage[id] = s
}

func (cn *ClientsChannels) Get(id int64) string {
	cn.RLock()
	defer cn.RUnlock()
	return cn.storage[id]
}

var (
	clientsChannels   = ClientsChannels{storage: make(map[int64]string)}
	centrifugoTimeout = time.Second * 5
	publisher         *gocent.Client
)

// InitCentrifugo client
func InitCentrifugo(cfg conf.CentrifugoConfig) {
	publisher = gocent.NewClient(cfg.URL, cfg.Secret, centrifugoTimeout)
}

func GetHMACSign(userID int64) (string, string, error) {
	timestamp := time.Now().Unix()
	secret, err := crypto.GetHMAC(centrifugoSecret, strconv.FormatInt(userID, 10), timestamp)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("HMAC getting error")
		return "", "", err
	}
	result := hex.EncodeToString(secret)
	clientsChannels[userID] = result
	return result, strconv.FormatInt(timestamp, 10), nil
}

// Write is publishing data to server
func Write(userID int64, data string) (bool, error) {
	return publisher.Publish("client#"+strconv.FormatInt(userID, 10), []byte(data))
}
