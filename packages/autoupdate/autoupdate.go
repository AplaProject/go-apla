package autoupdate

import (
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/tools/update_client/client"

	version "github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

const (
	// updates not tied to block number
	noBlockNumber = -1

	// interval value of checking updates
	checkUpdatesInterval = time.Hour
)

var updaterInstance *updater

type updateVersion struct {
	version     string
	blockNumber int64
}

type updater struct {
	updateAddr string
	pubkeyPath string

	client *client.UpdateClient

	updateVersions []updateVersion
}

func (u *updater) tryUpdate(currentBlockNumber int64) error {
	for _, update := range u.updateVersions {
		if update.blockNumber == noBlockNumber || currentBlockNumber+1 <= update.blockNumber {
			model.StopAll()

			appBinaryPath, err := os.Executable()
			if err != nil {
				log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("get application executable path")
				return err
			}

			err = u.client.UpdateFile(update.version, appBinaryPath, u.pubkeyPath)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("update executable file")
				return err
			}

			u.restart(appBinaryPath)
		}
	}

	return nil
}

func (u *updater) restart(appBinaryPath string) error {
	cmd := exec.Command(appBinaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("restart")
		return err
	}

	log.Info("restart")
	os.Exit(0)

	return nil
}

func (u *updater) checkUpdates() error {
	versions, err := u.getVersionsForUpdates()
	if err != nil {
		return err
	}

	u.updateVersions = make([]updateVersion, len(versions))
	for _, v := range versions {
		// TODO client will change
		err := u.client.GetBinary(u.updateAddr, u.pubkeyPath, v.String())
		if err != nil {
			return err
		}

		// TODO save block number
		u.updateVersions = append(u.updateVersions, updateVersion{version: v.String()})
	}

	return nil
}

// migrate executes database migrations
func (u *updater) migrate() error {
	err := model.ExecSchema()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MigrationError, "err": err}).Errorf("apply migrations")
		return err
	}

	return nil
}

// getVersionsForUpdates receives a list of versions for the operating system,
// followed by filtering versions that are greater current version of the application
func (u *updater) getVersionsForUpdates() ([]*version.Version, error) {
	verList, err := u.client.GetVersionList(u.updateAddr)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("get version list")
		return nil, err
	}

	appVer, err := version.NewVersion(consts.VERSION)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("parse version")
	}

	var neededVersions []*version.Version
	for _, v := range verList {
		ver, err := version.NewVersion(v)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.AutoupdateError, "err": err}).Error("parse version")
			return nil, err
		}
		if ver.GreaterThan(appVer) {
			neededVersions = append(neededVersions, ver)
		}
	}
	sort.Sort(version.Collection(neededVersions))
	return neededVersions, nil
}

// update startup scheduler every checkUpdatesInterval
func (u *updater) scheduler() {
	ticker := time.NewTicker(checkUpdatesInterval)
	for {
		select {
		case <-ticker.C:
			u.checkUpdates()
			u.tryUpdate(noBlockNumber)
		}
	}
}

// InitUpdater initializes the update
func InitUpdater(updateAddr string, pubkeyPath string) {
	updaterInstance = &updater{
		updateAddr: updateAddr,
		pubkeyPath: pubkeyPath,

		client:         &client.UpdateClient{},
		updateVersions: make([]updateVersion, 0),
	}
}

// Run starting updater
func Run() error {
	err := updaterInstance.checkUpdates()
	if err != nil {
		return err
	}

	err = updaterInstance.tryUpdate(noBlockNumber)
	if err != nil {
		return err
	}

	err = updaterInstance.migrate()
	if err != nil {
		return err
	}

	go updaterInstance.scheduler()

	return nil
}

// TryUpdate tries to update for the current block number
func TryUpdate(currentBlockNumber int64) error {
	return updaterInstance.tryUpdate(currentBlockNumber)
}
