package main

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"encoding/hex"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/tools/update_server/config"
	"github.com/GenesisKernel/go-genesis/tools/update_server/crypto"
	"github.com/GenesisKernel/go-genesis/tools/update_server/storage"
	"github.com/GenesisKernel/go-genesis/tools/update_server/web"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Config string `long:"conf" description:"path to config.ini" default:"./resources/config.ini" required:"true"`
}

func main() {
	fp := flags.NewParser(&opts, flags.Default)
	if _, err := fp.Parse(); err != nil {
		os.Exit(1)
	}

	p := config.NewParser(opts.Config)
	c, err := p.Do()
	if err != nil {
		log.Fatalf("Config parsing error: %s", err.Error())
	}

	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

	db, err := storage.NewBoltStorage(storage.NewBinaryStorage(c.StorageDir), c.DatabaseDir)
	if err != nil {
		log.WithFields(log.Fields{"errType": consts.IOError, "err": err}).Fatal("Creation bolt database")
	}

	pk, err := ioutil.ReadFile(c.PubkeyPath)
	if err != nil {
		log.WithFields(log.Fields{"errType": consts.IOError, "err": err}).Fatal("Reading public key")
	}
	pubKey, err := hex.DecodeString(string(pk))
	if err != nil {
		log.WithFields(log.Fields{"errType": consts.CryptoError, "err": err}).Fatal("Decoding public key")
	}

	s := web.Server{
		Db:        &db,
		Conf:      &c,
		PublicKey: pubKey,
		Signer:    &crypto.BuildSigner{},
	}

	log.WithFields(log.Fields{"errType": consts.NetworkError, "err": s.Run()}).Error("Server running")
}
