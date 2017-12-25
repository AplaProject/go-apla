package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/tools/update_server/config"
	"github.com/AplaProject/go-apla/tools/update_server/crypto"
	"github.com/AplaProject/go-apla/tools/update_server/storage"
	"github.com/AplaProject/go-apla/tools/update_server/web"
)

func main() {
	p := config.NewParser(filepath.Join(".", "resources", "config.ini"))
	c, err := p.Do()
	if err != nil {
		log.Fatalf("Config parsing error: %s", err.Error())
	}

	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

	db, err := storage.NewBoltStorage(c.DBPath)
	if err != nil {
		log.WithFields(log.Fields{"errType": consts.IOError, "err": err}).Fatal("Creation bolt database")
	}

	pk, err := ioutil.ReadFile(c.PubkeyPath)
	if err != nil {
		log.WithFields(log.Fields{"errType": consts.IOError, "err": err}).Fatal("Reading public key")
	}

	s := web.Server{
		Db:        &db,
		Conf:      &c,
		PublicKey: pk,
		Signer:    &crypto.BuildSigner{},
	}

	log.WithFields(log.Fields{"errType": consts.NetworkError, "err": s.Run()}).Error("Server running")
}
