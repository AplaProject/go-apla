// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
