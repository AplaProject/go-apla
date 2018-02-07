// MIT License
//
// Copyright (c) 2016 GenesisKernel
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
package install

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/GenesisCommunity/go-genesis/packages/conf"
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"
	"github.com/GenesisCommunity/go-genesis/packages/parser"
	"github.com/GenesisCommunity/go-genesis/packages/utils"
)

const fileMode = 0644

// ErrFirstBlockHostIsEmpty host for first block is not specified
var ErrFirstBlockHostIsEmpty = errors.New("FirstBlockHost is empty")

func createKeyPair(privFilename, pubFilename string) (priv, pub []byte, err error) {
	priv, pub, err = crypto.GenBytesKeys()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("generate keys")
		return
	}

	err = createFile(privFilename, []byte(hex.EncodeToString(priv)))
	if err != nil {
		return
	}

	err = createFile(pubFilename, []byte(hex.EncodeToString(pub)))
	if err != nil {
		return
	}

	return
}

func createFile(filename string, data []byte) error {
	err := ioutil.WriteFile(filename, data, fileMode)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing file")
		return err
	}
	return nil
}

func generateFirstBlock(publicKey, nodePublicKey []byte) error {
	if len(*conf.FirstBlockHost) == 0 {
		return ErrFirstBlockHostIsEmpty
	}

	now := time.Now().Unix()

	header := &utils.BlockData{
		BlockID:      1,
		Time:         now,
		EcosystemID:  0,
		KeyID:        conf.Config.KeyID,
		NodePosition: 0,
		Version:      consts.BLOCK_VERSION,
	}

	var tx []byte
	_, err := converter.BinMarshal(&tx,
		&consts.FirstBlock{
			TxHeader: consts.TxHeader{
				// TODO: move types to enum
				Type: 1, // FirstBlock

				Time:  uint32(now),
				KeyID: conf.Config.KeyID,
			},
			PublicKey:     publicKey,
			NodePublicKey: nodePublicKey,
			Host:          *conf.FirstBlockHost,
		},
	)

	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("first block body bin marshalling")
		return err
	}

	block, err := parser.MarshallBlock(header, [][]byte{tx}, []byte("0"), "")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("first block marshalling")
		return err
	}

	return createFile(*conf.FirstBlockPath, block)
}

// GenerateFirstBlock generates the first block
func GenerateFirstBlock() error {
	var publicKey, nodePublicKey []byte
	var err error

	// publicKey
	if len(*conf.FirstBlockPublicKey) > 0 {
		publicKey, err = hex.DecodeString(*conf.FirstBlockPublicKey)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding key from hex to string")
			return err
		}
	} else {
		_, publicKey, err = createKeyPair(
			filepath.Join(conf.Config.PrivateDir, consts.PrivateKeyFilename),
			filepath.Join(conf.Config.PrivateDir, consts.PublicKeyFilename),
		)
		if err != nil {
			return err
		}
	}

	// nodePublicKey
	if len(*conf.FirstBlockNodePublicKey) > 0 {
		nodePublicKey, err = hex.DecodeString(*conf.FirstBlockNodePublicKey)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding key from hex to string")
			return err
		}
	} else {
		_, nodePublicKey, err = createKeyPair(
			filepath.Join(conf.Config.PrivateDir, consts.NodePrivateKeyFilename),
			filepath.Join(conf.Config.PrivateDir, consts.NodePublicKeyFilename),
		)
		if err != nil {
			return err
		}
	}

	address := crypto.Address(publicKey)
	conf.Config.KeyID = address

	err = createFile(
		filepath.Join(conf.Config.PrivateDir, consts.KeyIDFilename),
		[]byte(strconv.FormatInt(address, 10)),
	)
	if err != nil {
		return err
	}

	return generateFirstBlock(publicKey, nodePublicKey)
}

// IsExistFirstBlock returns a boolean indicating whether first block file exists
func IsExistFirstBlock() bool {
	if _, err := os.Stat(*conf.FirstBlockPath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
