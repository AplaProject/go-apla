// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package tcpserver

import (
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/network"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/utils"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	log "github.com/sirupsen/logrus"
)

// Type2 serves requests from disseminator
func Type2(rw io.ReadWriter) (*network.DisTrResponse, error) {
	r := &network.DisRequest{}
	if err := r.Read(rw); err != nil {
		return nil, err
	}

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

	tx := transaction.RawTransaction{}
	if err = tx.Unmarshall(bytes.NewBuffer(decryptedBinData)); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
		return nil, err
	}

	_, err = model.DeleteQueueTxByHash(nil, tx.Hash())
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "hash": tx.Hash()}).Error("Deleting queue_tx with hash")
		return nil, utils.ErrInfo(err)
	}

	queueTx := &model.QueueTx{Hash: tx.Hash(), Data: decryptedBinData, FromGate: 0}
	err = queueTx.Create()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Creating queue_tx")
		return nil, utils.ErrInfo(err)
	}

	return &network.DisTrResponse{}, nil
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

	nodeKeyPrivate, _ := utils.GetNodeKeys()
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
