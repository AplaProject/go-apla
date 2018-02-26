package modes

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/install"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/julienschmidt/httprouter"
	pConf "github.com/rpoletaev/supervisord/config"
	"github.com/rpoletaev/supervisord/process"
	log "github.com/sirupsen/logrus"
)

const (
	createRoleTemplate = `CREATE ROLE %s WITH ENCRYPTED PASSWORD '%s' NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN`
	createDBTemplate   = `CREATE DATABASE %s OWNER %s`
	commandTemplate    = `./go-genesis -VDEMode=true -configPath=%s -workDir=%s`
)

// InitVDEMaster returns new master of VDE
func InitVDEMaster(config *conf.VDEMasterConfig) *VDEMaster {
	mode := &VDEMaster{
		VDEMasterConfig: config,
		VDE:             InitVDEMode(config.VDEConfig),
		configsPath:     path.Join(config.WorkDir, "configs"),
		processes:       process.NewProcessManager(),
	}

	mode.registerHandlers(mode.VDE.api)
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
func (mode *VDEMaster) Start(exitFunc func(int), gormInit func(conf.DBConfig), listenerFunc func(string, *httprouter.Router)) {

	mode.VDE.Start(exitFunc, gormInit, listenerFunc)

	//TODO: load master implementations
	if err := mode.prepareWorkDir(); err != nil {
		exitFunc(1)
	}

	if err := mode.initProcessManager(); err != nil {
		exitFunc(1)
	}
}

func (mode *VDEMaster) DaemonList() []string {
	return mode.VDE.DaemonList()
}

func (mode *VDEMaster) Stop() {
	mode.processes.StopAllProcesses()
	log.Infoln("VDEMaster mode stopped")
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
func (mode *VDEMaster) CreateVDE(name, dbUser, dbPassword string) error {

	if err := mode.createVDEDB(name, dbUser, dbPassword); err != nil {
		return err
	}

	if err := mode.initVDEDir(name, mode.VDE.VDEConfig); err != nil {
		return err
	}

	vdeDir := path.Join(mode.configsPath, name)
	privFile := filepath.Join(vdeDir, consts.PrivateKeyFilename)
	pubFile := filepath.Join(vdeDir, consts.PublicKeyFilename)
	_, _, err := install.CreateKeyPair(privFile, pubFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error on creating keys")
		return err
	}

	vdeConfigPath := filepath.Join(vdeDir, "config.toml")
	vdeConfig := *mode.VDE.VDEConfig
	vdeConfig.WorkDir = vdeDir
	vdeConfig.DB.User = dbUser
	vdeConfig.DB.Password = dbPassword
	vdeConfig.DB.Name = name
	vdeConfig.HTTP.Port = 7081

	conf.SaveVDEConfig(vdeConfigPath, &vdeConfig)

	confEntry := pConf.NewConfigEntry(vdeDir)
	confEntry.Name = "program:" + name
	command := fmt.Sprintf("./go-genesis -VDEMode=true -initDatabase=true -configPath=%s -workDir=%s", vdeConfigPath, vdeDir)
	confEntry.AddKeyValue("command", command)
	proc := process.NewProcess("vdeMaster", confEntry)

	mode.processes.Add(name, proc)
	mode.processes.Find(name).Start(true)
	log.Infoln(command)
	return nil
}

func (mode *VDEMaster) createVDEDB(vdeName, login, pass string) error {

	// md5pas := getMD5Pass(login, pass)
	if err := model.DBConn.Exec(fmt.Sprintf(createRoleTemplate, login, pass)).Error; err != nil {
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

func getMD5Pass(login, pass string) string {
	hasher := md5.New()
	hasher.Write([]byte(login + pass))
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

func (mode *VDEMaster) initProcessManager() error {

	list, err := ioutil.ReadDir(mode.configsPath)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Initialising VDE list")
		return err
	}

	for _, item := range list {
		if item.IsDir() {
			procDir := path.Join(mode.configsPath, item.Name())
			commandStr := fmt.Sprintf(commandTemplate, filepath.Join(procDir, "config.toml"), procDir)

			confEntry := pConf.NewConfigEntry(procDir)
			confEntry.Name = "program:" + item.Name()
			confEntry.AddKeyValue("command", commandStr)
			confEntry.AddKeyValue("redirect_stderr", "true")
			proc := process.NewProcess("vdeMaster", confEntry)

			mode.processes.Add(item.Name(), proc)
		}
	}

	return nil
}

// ListProcess returns list of process names with state of process
func (mode *VDEMaster) ListProcess() map[string]string {
	list := make(map[string]string)

	mode.processes.ForEachProcess(func(p *process.Process) {
		list[p.GetName()] = p.GetState().String()
	})

	return list
}

// DeleteVDE stop VDE process and remove VDE folder
func (mode *VDEMaster) DeleteVDE(name string) error {
	p := mode.processes.Find(name)
	if p != nil {
		p.Stop(true)
	}

	vdeDir := path.Join(mode.configsPath, name)
	return os.RemoveAll(vdeDir)
}
