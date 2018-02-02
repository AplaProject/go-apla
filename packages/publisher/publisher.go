package publisher

import (
	"encoding/hex"
	"strconv"
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
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

// GetHMACSign returns HMACS sign for userID
func GetHMACSign(userID int64) (string, error) {
	secret, err := crypto.GetHMAC(conf.Config.Centrifugo.Secret, strconv.FormatInt(userID, 10))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("HMAC getting error")
		return "", err
	}
	result := hex.EncodeToString(secret)
	clientsChannels.Set(userID, result)
	return result, nil
}

// Write is publishing data to server
func Write(userID int64, data string) (bool, error) {
	return publisher.Publish("client#"+strconv.FormatInt(userID, 10), []byte(data))
}
