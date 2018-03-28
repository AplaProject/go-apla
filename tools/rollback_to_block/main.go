package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/smart"
)

var blockID *int64 = flag.Int64("blockID", 1, "blockID to rollback")
var dbHost *string = flag.String("dbHost", "localhost", "genesis database host to rollback")
var dbPort *int = flag.Int("dbPort", 5432, "genesis database port to rollback")
var dbName *string = flag.String("dbName", "genesis", "genesis database name")
var dbUser *string = flag.String("dbUser", "genesis", "genesis database username")

func main() {
	dbPassword := os.Getenv("DB_PASSWORD")
	flag.Parse()
	if err := model.GormInit(*dbHost, *dbPort, *dbUser, dbPassword, *dbName); err != nil {
		log.WithError(err).Fatal("init db")
		return
	}
	if err := syspar.SysUpdate(nil); err != nil {
		log.WithError(err).Error("can't read system parameters")
	}
	if err := smart.LoadContracts(nil); err != nil {
		log.WithError(err).Fatal("loading contracts")
		return
	}
	parser := new(parser.Parser)
	err := parser.RollbackToBlockID(*blockID)
	if err != nil {
		log.WithError(err).Fatal("rollback to block id")
		return
	}

	// block id = 1, is a special case for full rollback
	if *blockID != 1 {
		log.Info("Not full rollback, finishing work without checking")
		return
	}
}
