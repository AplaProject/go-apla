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

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/parser"
	"github.com/AplaProject/go-apla/packages/utils"
)

const fileMode = 0644

// ErrFirstBlockHostIsEmpty host for first block is not specified
var ErrFirstBlockHostIsEmpty = errors.New("FirstBlockHost is empty")

func createKeyPair(privFilename, pubFilename string) (priv, pub string, err error) {
	priv, pub, err = crypto.GenHexKeys()
	if err != nil {
		return
	}

	err = createFile(privFilename, []byte(priv))
	if err != nil {
		return
	}

	err = createFile(pubFilename, []byte(pub))
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

func getPublicKeyAndCreateKeyPair(value *string, privFilename, pubFilename string) ([]byte, error) {
	if len(*value) == 0 {
		_, pub, err := createKeyPair(filepath.Join(conf.Config.PrivateDir, privFilename), filepath.Join(conf.Config.PrivateDir, pubFilename))
		if err != nil {
			return nil, err
		}
		*value = pub
	}

	key, err := hex.DecodeString(*value)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding key from hex to string")
		return nil, err
	}

	return key, nil
}

func firstBlockPublicKey() ([]byte, error) {
	return getPublicKeyAndCreateKeyPair(conf.FirstBlockPublicKey, consts.PrivateKeyFilename, consts.PublicKeyFilename)
}

func firstBlockNodePublicKey() ([]byte, error) {
	return getPublicKeyAndCreateKeyPair(conf.FirstBlockNodePublicKey, consts.NodePrivateKeyFilename, consts.NodePublicKeyFilename)
}

func createKeyIDFile(publicKey []byte) error {
	address := crypto.Address(publicKey)
	conf.Config.KeyID = address

	keyIDFile := filepath.Join(conf.Config.PrivateDir, consts.KeyIDFilename)
	data := []byte(strconv.FormatInt(address, 10))
	return createFile(keyIDFile, data)
}

func generateFirstBlock(publicKey, nodePublicKey []byte) error {
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
		return err
	}

	return createFile(*conf.FirstBlockPath, block)
}

// GenerateFirstBlock generates the first block
func GenerateFirstBlock() error {
	publicKey, err := firstBlockPublicKey()
	if err != nil {
		return err
	}

	nodePublicKey, err := firstBlockNodePublicKey()
	if err != nil {
		return err
	}

	err = createKeyIDFile(publicKey)
	if err != nil {
		return err
	}

	return generateFirstBlock(publicKey, nodePublicKey)
}

// IsNotExistFirstBlock returns a boolean indicating whether first block file not exist
func IsNotExistFirstBlock() bool {
	_, err := os.Stat(*conf.FirstBlockPath)
	return os.IsNotExist(err)
}
