package vdemanager

import (
	"fmt"
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/model"
)

func prepare(configPath string) error {
	if err := conf.LoadConfigFromPath(configPath); err != nil {
		return err
	}

	if err := model.GormInit(conf.Config.DB.Host, conf.Config.DB.Port, conf.Config.DB.User, conf.Config.DB.Password, conf.Config.DB.Name); err != nil {
		return err
	}

	return InitVDEManager()
}
func TestCreateVDE(t *testing.T) {

	if err := prepare("/home/losaped/go/src/github.com/GenesisKernel/go-genesis/config.toml"); err != nil {
		t.Error(err)
		return
	}

	if err := Manager.CreateVDE("one9", "one9", "one9", 8006); err != nil {
		t.Error(err)
		return
	}

}

func TestListVDE(t *testing.T) {

	if err := prepare("/home/losaped/go/src/github.com/GenesisKernel/go-genesis/config.toml"); err != nil {
		t.Error(err)
		return
	}

	list := Manager.ListProcess()
	fmt.Printf("%v\n", list)
}

func TestListRunningVDE(t *testing.T) {

	if err := prepare("/home/losaped/go/src/github.com/GenesisKernel/go-genesis/config.toml"); err != nil {
		t.Error(err)
		return
	}

	list := Manager.ListProcess()
	fmt.Printf("%v\n", list)

	if err := Manager.StartVDE("one3"); err != nil {
		t.Error(err)
		return
	}

	list = Manager.ListProcess()
	fmt.Printf("%v\n", list)
}
