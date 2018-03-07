package autoupdate

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/tools/update_client/client"
	"github.com/GenesisKernel/go-genesis/tools/update_client/params"
	updateModel "github.com/GenesisKernel/go-genesis/tools/update_server/model"

	version "github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

const (
	// updates not tied to block number
	noBlockNumber = 0
)

var updater *Updater

type updateVersion struct {
	version     string
	blockNumber uint64
}

// Updater is updater
type Updater struct {
	server     string
	pubkeyPath string

	client   *client.UpdateClient
	versions []*updateVersion
}

func (u *Updater) tryUpdate(currentBlockNumber uint64) error {
	for _, v := range u.versions {
		if v.blockNumber == noBlockNumber || currentBlockNumber+1 <= v.blockNumber {
			err := model.SetStopNow()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.DBError, "err": err}).Error("create stop daemon record")
			}

			executablePath, err := os.Executable()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("get application executable path")
				return err
			}

			build, err := u.client.GetBinary(
				params.ServerParams{Server: u.server},
				params.KeyParams{PublicKeyPath: u.pubkeyPath},
				params.BinaryParams{Version: v.version},
			)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("get binary data")
				return err
			}

			err = ioutil.WriteFile(executablePath, build.Body, 0755)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "err": err}).Error("update executable file")
				return err
			}

			err = u.restart(executablePath)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("restart")
				return err
			}

			return nil
		}
	}

	return nil
}

func (u *Updater) restart(executablePath string) error {
	cmd := exec.Command(executablePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	log.Info("Restart")
	os.Exit(0)

	return nil
}

func (u *Updater) checkUpdates() error {
	versions, err := u.versionsForUpdate()
	if err != nil {
		return err
	}

	u.versions = versions
	return nil
}

// migrate executes database migrations
func (u *Updater) migrate() error {
	err := model.ExecSchema()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "err": err}).Errorf("apply migrations")
		return err
	}

	return nil
}

// versionsForUpdate receives a list of versions for the operating system,
// followed by filtering versions that are greater current version of the application
func (u *Updater) versionsForUpdate() ([]*updateVersion, error) {
	verList, err := u.client.GetVersionList(
		params.ServerParams{Server: u.server},
		params.BinaryParams{Version: currentVersion()},
	)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("get version list")
		return nil, err
	}

	appVer, err := version.NewVersion(consts.VERSION)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("parse version")
	}

	var neededVersions []*updateVersion
	for _, v := range verList {
		ver, err := version.NewVersion(v.Number)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("parse version")
			return nil, err
		}
		if ver.GreaterThan(appVer) {
			neededVersions = append(neededVersions, &updateVersion{
				version:     v.String(),
				blockNumber: v.StartBlock,
			})
			log.WithFields(log.Fields{"version": v.Number}).Info("Update available")
		}
	}

	return neededVersions, nil
}

// update startup scheduler every checkUpdatesInterval
func (u *Updater) scheduler() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			u.checkUpdates()
			u.tryUpdate(noBlockNumber)
		}
	}
}

func currentVersion() string {
	v := updateModel.Version{
		Number: consts.VERSION,
		OS:     runtime.GOOS,
		Arch:   runtime.GOARCH,
	}

	return v.String()
}

// InitUpdater initializes the update
func InitUpdater(server string, pubkeyPath string) {
	updater = &Updater{
		server:     server,
		pubkeyPath: pubkeyPath,

		client:   &client.UpdateClient{},
		versions: make([]*updateVersion, 0),
	}
}

// Run starting updater
func Run() error {
	err := updater.checkUpdates()
	if err != nil {
		return err
	}

	err = updater.tryUpdate(noBlockNumber)
	if err != nil {
		return err
	}

	err = updater.migrate()
	if err != nil {
		return err
	}

	go updater.scheduler()

	return nil
}

// TryUpdate tries to update for the current block number
func TryUpdate(currentBlockNumber uint64) error {
	return updater.tryUpdate(currentBlockNumber)
}
