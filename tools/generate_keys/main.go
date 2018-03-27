package main

import (
	"encoding/hex"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/crypto"
	log "github.com/sirupsen/logrus"
)

const fileMode = 0644

const PrivateKeyFilename = "PrivateKey"
const PublicKeyFilename = "PublicKey"
const NodePrivateKeyFilename = "NodePrivateKey"
const NodePublicKeyFilename = "NodePublicKey"
const KeyIDFilename = "KeyID"

var path *string = flag.String("path", ".", "path to keys folder")

func createFile(filename string, data []byte) error {
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

	err = createFile(pubFilename, []byte(hex.EncodeToString(pub)))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "path": pubFilename}).Error("creating public key")
		return
	}

	return
}

func main() {
	flag.Parse()
	_, publicKey, err := createKeyPair(
		filepath.Join(*path, PrivateKeyFilename),
		filepath.Join(*path, PublicKeyFilename),
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("generating user keys")
		return
	}
	_, _, err = createKeyPair(
		filepath.Join(*path, NodePrivateKeyFilename),
		filepath.Join(*path, NodePublicKeyFilename),
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("generating node keys")
		return
	}
	address := crypto.Address(publicKey)
	keyIDPath := filepath.Join(*path, KeyIDFilename)
	err = createFile(keyIDPath, []byte(strconv.FormatInt(address, 10)))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "path": keyIDPath}).Fatal("generating node keys")
		return
	}
	log.Info("keys generated")
}
