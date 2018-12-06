package publisher

import (
	"encoding/hex"
	"fmt"
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
	config            conf.CentrifugoConfig
)

// InitCentrifugo client
func InitCentrifugo(cfg conf.CentrifugoConfig) {
	config = cfg
	publisher = gocent.NewClient(cfg.URL, cfg.Secret, centrifugoTimeout)
}

func GetHMACSign(userID int64) (string, string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	secret, err := crypto.GetHMACWithTimestamp(config.Secret, strconv.FormatInt(userID, 10), timestamp)

	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("HMAC getting error")
		return "", "", err
	}

	result := hex.EncodeToString(secret)
	clientsChannels.Set(userID, result)
	return result, timestamp, nil
}

// Write is publishing data to server
func Write(userID int64, data string) (bool, error) {
	return publisher.Publish("client"+strconv.FormatInt(userID, 10), []byte(data))
}

// GetStats returns Stats
func GetStats() (gocent.Stats, error) {
	if publisher == nil {
		return gocent.Stats{}, fmt.Errorf("publisher not initialized")
	}

	return publisher.Stats()
}
