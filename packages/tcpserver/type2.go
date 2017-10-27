// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package tcpserver

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"
)

// Type2 serves requests from disseminator
func Type2(r *DisRequest) (*DisTrResponse, error) {
	binaryData := r.Data
	// take the transactions from usual users but not nodes.
	_, _, decryptedBinData, err := DecryptData(&binaryData)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	if int64(len(binaryData)) > consts.MAX_TX_SIZE {
		return nil, utils.ErrInfo("len(txBinData) > max_tx_size")
	}

	if len(binaryData) < 5 {
		return nil, utils.ErrInfo("len(binaryData) < 5")
	}

	decryptedBinDataFull := decryptedBinData
	hash, err := crypto.Hash(decryptedBinDataFull)
	if err != nil {
		log.Fatal(err)
	}

	_, err = model.DeleteQueueTxByHash(nil, hash)
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	//hexBinData := converter.BinToHex(decryptedBinDataFull)
	log.Debug("INSERT INTO queue_tx (hash, data) (%s, %s)", hash, decryptedBinData)
	queueTx := &model.QueueTx{Hash: hash, Data: decryptedBinData, FromGate: 0}
	err = queueTx.Create()
	if err != nil {
		return nil, utils.ErrInfo(err)
	}

	return &DisTrResponse{}, nil
}

func DecryptData(binaryTx *[]byte) ([]byte, []byte, []byte, error) {
	if len(*binaryTx) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(binaryTx) == 0")
	}

	myUserID := converter.BinToDecBytesShift(&*binaryTx, 5)
	log.Debug("myUserId: %d", myUserID)

	// remove the encrypted key, and all that stay in $binary_tx will be encrypted keys of the transactions/blocks
	length, err := converter.DecodeLength(&*binaryTx)
	if err != nil {
		log.Fatal(err)
	}
	encryptedKey := converter.BytesShift(&*binaryTx, length)

	iv := converter.BytesShift(&*binaryTx, 16)

	if len(encryptedKey) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(encryptedKey) == 0")
	}

	if len(*binaryTx) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(*binaryTx) == 0")
	}

	nodeKey := &model.MyNodeKey{}
	err = nodeKey.GetNodeWithMaxBlockID()
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}
	if len(nodeKey.PrivateKey) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(nodePrivateKey) == 0")
	}

	block, _ := pem.Decode([]byte(nodeKey.PrivateKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, nil, nil, utils.ErrInfo("No valid PEM data found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}

	decKey, err := rsa.DecryptPKCS1v15(crand.Reader, privateKey, encryptedKey)
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}

	if len(decKey) == 0 {
		return nil, nil, nil, utils.ErrInfo("len(decKey)")
	}

	decrypted, err := crypto.Decrypt(iv, *binaryTx, decKey)
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}

	return decKey, iv, decrypted, nil
}
