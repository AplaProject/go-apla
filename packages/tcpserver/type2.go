//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package tcpserver

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"
	"github.com/GenesisCommunity/go-genesis/packages/model"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

	"github.com/GenesisCommunity/go-genesis/packages/config/syspar"
	log "github.com/sirupsen/logrus"
)

// Type2 serves requests from disseminator
func Type2(r *DisRequest) (*DisTrResponse, error) {
	binaryData := r.Data
	// take the transactions from usual users but not nodes.
	_, _, decryptedBinData, err := DecryptData(&binaryData)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	if int64(len(binaryData)) > syspar.GetMaxTxSize() {
		log.WithFields(log.Fields{"type": consts.ParameterExceeded, "max_size": syspar.GetMaxTxSize(), "size": len(binaryData)}).Error("transaction size exceeds max size")
		return nil, utils.ErrInfo("len(txBinData) > max_tx_size")
	}

	if len(binaryData) < 5 {
		log.WithFields(log.Fields{"type": consts.ProtocolError, "len": len(binaryData), "should_be_equal": 5}).Error("binary data slice has incorrect length")
		return nil, utils.ErrInfo("len(binaryData) < 5")
	}

	decryptedBinDataFull := decryptedBinData
	hash, err := crypto.Hash(decryptedBinDataFull)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err, "value": decryptedBinDataFull}).Fatal("cannot hash tx bindata")
	}

	_, err = model.DeleteQueueTxByHash(nil, hash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "hash": hash}).Error("Deleting queue_tx with hash")
		return nil, utils.ErrInfo(err)
	}

	//hexBinData := converter.BinToHex(decryptedBinDataFull)
	queueTx := &model.QueueTx{Hash: hash, Data: decryptedBinData, FromGate: 0}
	err = queueTx.Create()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Creating queue_tx")
		return nil, utils.ErrInfo(err)
	}

	return &DisTrResponse{}, nil
}

// DecryptData is decrypting data
func DecryptData(binaryTx *[]byte) ([]byte, []byte, []byte, error) {
	if len(*binaryTx) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("binary tx is empty")
		return nil, nil, nil, utils.ErrInfo("len(binaryTx) == 0")
	}

	myUserID := converter.BinToDecBytesShift(&*binaryTx, 5)
	log.WithFields(log.Fields{"user_id": myUserID}).Debug("decrypted userID is")

	// remove the encrypted key, and all that stay in $binary_tx will be encrypted keys of the transactions/blocks
	length, err := converter.DecodeLength(&*binaryTx)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ProtocolError, "error": err}).Fatal("Decoding binary tx length")
	}
	encryptedKey := converter.BytesShift(&*binaryTx, length)
	iv := converter.BytesShift(&*binaryTx, 16)
	log.WithFields(log.Fields{"encryptedKey": encryptedKey, "iv": iv}).Debug("binary tx encryptedKey and iv is")

	if len(encryptedKey) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("binary tx encrypted key is empty")
		return nil, nil, nil, utils.ErrInfo("len(encryptedKey) == 0")
	}

	if len(*binaryTx) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("binary tx is empty")
		return nil, nil, nil, utils.ErrInfo("len(*binaryTx) == 0")
	}

	nodeKeyPrivate, _, err := utils.GetNodeKeys()
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}
	if len(nodeKeyPrivate) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		return nil, nil, nil, utils.ErrInfo("len(nodePrivateKey) == 0")
	}

	block, _ := pem.Decode([]byte(nodeKeyPrivate))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.WithFields(log.Fields{"type": consts.CryptoError}).Error("No valid PEM data found")
		return nil, nil, nil, utils.ErrInfo("No valid PEM data found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("Parse PKCS1PrivateKey")
		return nil, nil, nil, utils.ErrInfo(err)
	}

	decKey, err := rsa.DecryptPKCS1v15(crand.Reader, privateKey, encryptedKey)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("rsa Decrypt")
		return nil, nil, nil, utils.ErrInfo(err)
	}

	log.WithFields(log.Fields{"key": decKey}).Debug("decrypted key")
	if len(decKey) == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("decrypted key is empty")
		return nil, nil, nil, utils.ErrInfo("len(decKey)")
	}

	log.WithFields(log.Fields{"binaryTx": *binaryTx, "iv": iv}).Debug("binaryTx and iv is")
	decrypted, err := crypto.Decrypt(iv, *binaryTx, decKey)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("Decryption binary tx")
		return nil, nil, nil, utils.ErrInfo(err)
	}

	return decKey, iv, decrypted, nil
}
