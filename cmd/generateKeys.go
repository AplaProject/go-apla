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

package cmd

import (
	"encoding/hex"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"os"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const fileMode = 0600

// generateKeysCmd represents the generateKeys command
var generateKeysCmd = &cobra.Command{
	Use:    "generateKeys",
	Short:  "Keys generation",
	PreRun: loadConfig,
	Run: func(cmd *cobra.Command, args []string) {
		_, publicKey, err := createKeyPair(
			filepath.Join(conf.Config.KeysDir, consts.PrivateKeyFilename),
			filepath.Join(conf.Config.KeysDir, consts.PublicKeyFilename),
		)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("generating user keys")
			return
		}
		_, _, err = createKeyPair(
			filepath.Join(conf.Config.KeysDir, consts.NodePrivateKeyFilename),
			filepath.Join(conf.Config.KeysDir, consts.NodePublicKeyFilename),
		)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("generating node keys")
			return
		}
		address := crypto.Address(publicKey)
		keyIDPath := filepath.Join(conf.Config.KeysDir, consts.KeyIDFilename)
		err = createFile(keyIDPath, []byte(strconv.FormatInt(address, 10)))
		if err != nil {
			log.WithFields(log.Fields{"error": err, "path": keyIDPath}).Fatal("generating node keys")
			return
		}
		log.Info("keys generated")
	},
}

func createFile(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0775)
		if err != nil {
			return errors.Wrapf(err, "creating dir %s", dir)
		}
	}

	return ioutil.WriteFile(filename, data, fileMode)
}

func createKeyPair(privFilename, pubFilename string) (priv, pub []byte, err error) {
	priv, pub, err = crypto.GenBytesKeys()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("generate keys")
		return
	}

	err = createFile(privFilename, []byte(hex.EncodeToString(priv)))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "path": privFilename}).Error("creating private key")
		return
	}

	err = createFile(pubFilename, []byte(crypto.PubToHex(pub)))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "path": pubFilename}).Error("creating public key")
		return
	}

	return
}
