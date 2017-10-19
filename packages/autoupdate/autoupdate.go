package autoupdate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/AplaProject/go-apla/packages/migration"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/tools/update_client/client"
	"github.com/AplaProject/go-apla/tools/update_client/structs"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	version "github.com/hashicorp/go-version"
)

type UpdateTime struct {
	version     string
	blockNumber int64
}

type Updater struct {
	updateAddr  string
	pubkeyPath  string
	UpdateTimes []UpdateTime
}

func NewUpdater(updateAddr string, pubkeyPath string) *Updater {
	return &Updater{updateAddr: updateAddr, pubkeyPath: pubkeyPath, UpdateTimes: make([]UpdateTime, 0)}
}

func (u *Updater) TryUpdate(currentBlockNumber int64) {
	for _, update := range u.UpdateTimes {
		if update.blockNumber == -1 || currentBlockNumber+1 == update.blockNumber {
			utils.Stop()
			clientName, err := filepath.Glob("update_client*")
			if err != nil {
				//TODO add log
			}
			selfName, err := os.Executable()
			if err != nil {
				//TODO add log
			}
			exec.Command(*utils.Dir+"/"+clientName[0],
				fmt.Sprintf(`-command="u" -version="%s" -remove="%s -public="%s"`, update.version, selfName, u.pubkeyPath))
			os.Exit(0)
		}
	}
}

func (u *Updater) CheckUpdates(pubKeyPath string) error {
	versions, err := u.getNeededVersionsUpdates(u.updateAddr)
	u.UpdateTimes = make([]UpdateTime, len(versions))
	for _, vers := range versions {
		u.UpdateTimes = append(u.UpdateTimes, UpdateTime{version: vers.String()})
	}
	if err != nil {
		return err
	}
	for _, ver := range versions {
		err = u.checkUpdateFile(ver, pubKeyPath)
		if err != nil {
			err = u.downloadUpdate(u.updateAddr, ver, u.pubkeyPath)
			if err != nil {
				return err
			}
			u.checkUpdateFile(ver, pubKeyPath)
		}
	}
	time.Sleep(time.Hour)
	return nil
}

func (u *Updater) Migrate(vers *version.Version) error {
	if model.DBConn == nil {
		return errors.New("database disconnected")
	}
	return model.DBConn.Exec(migration.VersionedMigrations[vers.String()]).Error
}

func (u *Updater) getNeededVersionsUpdates(updateAddr string) ([]*version.Version, error) {
	client := &client.UpdateClient{}
	vers, err := client.GetVersionList(updateAddr)
	if err != nil {
		return nil, err
	}
	dbVersion, err := getLocalVersion()
	if err != nil {
		return nil, err
	}

	softwareVersion, _ := version.NewVersion(consts.VERSION)
	if dbVersion.LessThan(softwareVersion) {
		err = migration.Migrate(softwareVersion)
		if err != nil {
			return nil, err
		}
	}

	var neededVersions []*version.Version
	for _, ver := range vers {
		testVersion, err := version.NewVersion(ver)
		if err != nil {
			return nil, err
		}
		if testVersion.GreaterThan(currentVersion) {
			neededVersions = append(neededVersions, testVersion)
		}
	}
	sort.Sort(version.Collection(neededVersions))
	return neededVersions, nil
}

func (u *Updater) downloadUpdate(updateAddr string, version *version.Version, publicKey string) error {
	client := &client.UpdateClient{}
	err := client.GetBinary(updateAddr, publicKey, version.String())
	if err != nil {
		return err
	}
	return nil
}

func (u *Updater) checkUpdateFile(version *version.Version, publicKey string) error {
	file, err := os.Open("update" + version.String())
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	var binary structs.Binary
	err = json.Unmarshal(data, &binary)
	if err != nil {
		return err
	}

	pubFile, err := os.Open(publicKey)
	if err != nil {
		return err
	}
	pubData, err := ioutil.ReadAll(pubFile)
	if err != nil {
		return err
	}
	verified, err := binary.CheckSign(pubData)
	if err != nil {
		return err
	}
	if verified != true {
		return errors.New("can't vefiry sign")
	}

	for _, update := range u.UpdateTimes {
		if update.version == version.String() {
			update.blockNumber = binary.StartBlock
			break
		}
	}

	return nil
}

func getLocalVersion() (*version.Version, error) {
	migration := &model.MigrationHistory{}
	err := migration.Get()
	if err != nil {
		return nil, err
	}
	result, _ := version.NewVersion(migration.Version)
	return result, nil
}
