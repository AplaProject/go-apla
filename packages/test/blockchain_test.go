package test

import (
	"context"
	"testing"
	"time"

	"github.com/AplaProject/go-apla/packages/apiv2"
	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/daemons"
	"github.com/AplaProject/go-apla/packages/utils"
)

func reInstall() error {
	*utils.Dir = "."
	params := &apiv2.InstallParams{
		GenerateFirstBlock: false,
		InstallType:        "TESTNET_URL",
		DbHost:             "localhost",
		DbName:             "aplatest",
		DbUsername:         "aplatest",
		LogLevel:           "DEBUG",
	}
	if err := apiv2.InstallCommon(params); err != nil {
		return err
	}

	if err := syspar.SysUpdate(); err != nil {
		return err
	}

	return nil
}

func TestOld(t *testing.T) {
	err := reInstall()
	if err != nil {
		t.Fatalf("install failed: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := daemons.LoadFromFile(ctx, "./blockchain"); err != nil {
		t.Fatalf("error while loading: %s", err)
	}
}
