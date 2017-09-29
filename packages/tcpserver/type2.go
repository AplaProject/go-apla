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
	"fmt"

	"encoding/hex"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// Type2 serves requests from disseminator
func Type2(r *DisRequest) (*DisTrResponse, error) {
	logger.LogDebug(consts.FuncStarted, "")
	binaryData := r.Data
	// take the transactions from usual users but not nodes.
	_, _, decryptedBinData, err := DecryptData(&binaryData)
	if err != nil {
		logger.LogError(consts.CryptoError, err)
		return nil, utils.ErrInfo(err)
	}

	if int64(len(binaryData)) > consts.MAX_TX_SIZE {
		logger.LogError(consts.TransactionError, "len(txBinData) > max_tx_size")
		return nil, utils.ErrInfo("len(txBinData) > max_tx_size")
	}

	if len(binaryData) < 5 {
		logger.LogError(consts.TransactionError, "len(binaryData) < 5")
		return nil, utils.ErrInfo("len(binaryData) < 5")
	}

	decryptedBinDataFull := decryptedBinData
	hash, err := crypto.Hash(decryptedBinDataFull)
	if err != nil {
		logger.LogFatal(consts.TransactionError, err)
	}

	err = model.DeleteQueuedTransaction(hash)
	if err != nil {
		logger.LogError(consts.DBError, err)
		return nil, utils.ErrInfo(err)
	}

	queueTx := &model.QueueTx{Hash: hash, Data: decryptedBinData, FromGate: 0}
	err = queueTx.Create()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return nil, utils.ErrInfo(err)
	}

	return &DisTrResponse{}, nil
}

func DecryptData(binaryTx *[]byte) ([]byte, []byte, []byte, error) {
	logger.LogDebug(consts.FuncStarted, "")
	if len(*binaryTx) == 0 {
		logger.LogError(consts.TransactionError, "len(binaryTx) == 0")
		return nil, nil, nil, utils.ErrInfo("len(binaryTx) == 0")
	}

	myUserID := converter.BinToDecBytesShift(&*binaryTx, 5)
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("myUserId: %d", myUserID))

	// remove the encrypted key, and all that stay in $binary_tx will be encrypted keys of the transactions/blocks
	length, err := converter.DecodeLength(&*binaryTx)
	if err != nil {
		logger.LogFatal(consts.TransactionError, err)
	}
	encryptedKey := converter.BytesShift(&*binaryTx, length)
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("encryptedKey: %s", encryptedKey))
	iv := converter.BytesShift(&*binaryTx, 16)
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("iv: %x", iv))

	if len(encryptedKey) == 0 {
		logger.LogError(consts.CryptoError, "len(encryptedKey) == 0")
		return nil, nil, nil, utils.ErrInfo("len(encryptedKey) == 0")
	}

	if len(*binaryTx) == 0 {
		logger.LogError(consts.TransactionError, "len(*binaryTx) == 0")
		return nil, nil, nil, utils.ErrInfo("len(*binaryTx) == 0")
	}

	nodeKey := &model.MyNodeKey{}
	err = nodeKey.GetNodeWithMaxBlockID()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return nil, nil, nil, utils.ErrInfo(err)
	}
	if len(nodeKey.PrivateKey) == 0 {
		logger.LogError(consts.RecordNotFoundError, err)
		return nil, nil, nil, utils.ErrInfo("len(nodePrivateKey) == 0")
	}

	block, _ := pem.Decode([]byte(hex.EncodeToString(nodeKey.PrivateKey)))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		logger.LogError(consts.CryptoError, "No valid PEM data found")
		return nil, nil, nil, utils.ErrInfo("No valid PEM data found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, nil, utils.ErrInfo(err)
	}

	decKey, err := rsa.DecryptPKCS1v15(crand.Reader, privateKey, encryptedKey)
	if err != nil {
		logger.LogError(consts.CryptoError, err)
		return nil, nil, nil, utils.ErrInfo(err)
	}
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("decrypted Key: %s", decKey))
	if len(decKey) == 0 {
		logger.LogError(consts.CryptoError, "len(decKey) == 0")
		return nil, nil, nil, utils.ErrInfo("len(decKey)")
	}

	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("binaryTx %x", *binaryTx))
	logger.LogDebug(consts.DebugMessage, fmt.Sprintf("iv %s", iv))
	decrypted, err := crypto.Decrypt(iv, *binaryTx, decKey)
	if err != nil {
		logger.LogError(consts.CryptoError, err)
		return nil, nil, nil, utils.ErrInfo(err)
	}

	return decKey, iv, decrypted, nil
}
