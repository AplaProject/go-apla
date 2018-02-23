package modes

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	pConf "github.com/ochinchina/supervisord/config"
	"github.com/ochinchina/supervisord/process"
	log "github.com/sirupsen/logrus"
)

const (
	createRoleTemplate = `CREATE ROLE %s PASSWORD '%s' NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN`
	createDBTemplate   = `CREATE DATABASE %s OWNER %s`
)

// InitVDEMaster returns new master of VDE
func InitVDEMaster(config *conf.VDEMasterConfig) *VDEMaster {
	mode := &VDEMaster{
		VDEMasterConfig: config,
		VDE:             InitVDEMode(config.VDEConfig),
		configsPath:     path.Join(config.WorkDir, "configs"),
		processes:       process.NewProcessManager(),
	}

	return mode
}

// VDEMaster represents master of VDE mode
type VDEMaster struct {
	*conf.VDEMasterConfig
	*VDE
	configsPath string
	processes   *process.ProcessManager
}

// Start implements NodeMode interface
func (mode *VDEMaster) Start(exitFunc func(int), gormInit func(conf.DBConfig)) {
	mode.VDE.Start(exitFunc, gormInit)

	//TODO: load master implementations
	if err := mode.prepareWorkDir(); err != nil {
		exitFunc(1)
	}

	if err := mode.runProcesses(); err != nil {
		exitFunc(1)
	}
}

func (mode *VDEMaster) prepareWorkDir() error {
	if _, err := os.Stat(mode.configsPath); os.IsNotExist(err) {
		if err := os.Mkdir(mode.configsPath, 0700); err != nil {
			log.WithFields(log.Fields{"type": consts.VDEMasterError, "error": err}).Error("creating configs directory")
			return err
		}
	}

	return nil
}

// CreateVDE creates one instance of VDE
func (mode *VDEMaster) CreateVDE(name string, config *conf.VDEConfig) error {

	if err := mode.createVDEDB(name, config.DB.User, config.DB.Password); err != nil {
		return err
	}

	if err := mode.loadProcess(); err != nil {
		return err
	}

	if err := mode.initVDEDir(name, config); err != nil {

		return err
	}
	return nil
}

func (mode *VDEMaster) createVDEDB(vdeName, login, pass string) error {

	md5pas := getMD5Pass(pass)
	if err := model.DBConn.Exec(fmt.Sprintf(createRoleTemplate, login, md5pas)).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating VDE DB User")
		return err
	}

	if err := model.DBConn.Exec(fmt.Sprintf(createDBTemplate, vdeName, login)).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating VDE DB")
		return err
	}

	return nil
}

func (mode *VDEMaster) initVDEDir(vdeName string, config *conf.VDEConfig) error {

	vdeDirName := path.Join(mode.configsPath, vdeName)
	if _, err := os.Stat(vdeDirName); os.IsNotExist(err) {
		if err := os.Mkdir(vdeDirName, 0700); err != nil {
			log.WithFields(log.Fields{"type": consts.VDEMasterError, "error": err}).Error("creating VDE directory")
			return err
		}
	}

	configPath := path.Join(vdeDirName, consts.DefaultConfigFile)
	return saveConfigFile(configPath, config)
}

func getMD5Pass(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return "md5" + hex.EncodeToString(hasher.Sum(nil))
}

func saveConfigFile(path string, config interface{}) error {
	var cf *os.File
	var err error

	defer cf.Close()

	if _, err = os.Stat(path); os.IsNotExist(err) {
		if cf, err = os.Create(path); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Create VDE config file failed")
			return err
		}
	} else {
		if cf, err = os.Open(path); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Open VDE config file failed")
			return err
		}
	}

	if err = toml.NewEncoder(cf).Encode(config); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Saving config error")
		return err
	}

	return nil
}

func (mode *VDEMaster) loadProcess() error {

	list, err := ioutil.ReadDir(mode.configsPath)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Initialising VDE list")
		return err
	}

	for _, item := range list {
		if item.IsDir() {
			procDir := path.Join(mode.configsPath, item.Name())
			confEntry := pConf.NewConfigEntry(procDir)
			confEntry.Name = "program:" + item.Name()
			confEntry.
			proc := process.NewProcess("vdeMaster", confEntry)
			mode.processes.Add(item.Name(), proc)
		}
	}

	return nil
}
