package vdemanager

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/crypto"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	pConf "github.com/rpoletaev/supervisord/config"
	"github.com/rpoletaev/supervisord/process"
	log "github.com/sirupsen/logrus"
)

const (
	childFolder        = "configs"
	createRoleTemplate = `CREATE ROLE %s WITH ENCRYPTED PASSWORD '%s' NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN`
	createDBTemplate   = `CREATE DATABASE %s OWNER %s`

	dropDBTemplate     = `DROP OWNED BY %s CASCADE`
	dropDBRoleTemplate = `DROP ROLE IF EXISTS %s`
	commandTemplate    = `%s -VDEMode=true -configPath=%s -workDir=%s`
)

// VDEManager struct
type VDEManager struct {
	processes *process.ProcessManager
}

var (
	Manager          *VDEManager
	childConfigsPath string
)

// InitVDEManager create init instance of VDEManager
func InitVDEManager() error {
	if err := prepareWorkDir(); err != nil {
		return err
	}

	return initProcessManager()
}

func prepareWorkDir() error {
	childConfigsPath = path.Join(conf.Config.WorkDir, childFolder)

	if _, err := os.Stat(childConfigsPath); os.IsNotExist(err) {
		if err := os.Mkdir(childConfigsPath, 0700); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating configs directory")
			return err
		}
	}

	return nil
}

// CreateVDE creates one instance of VDE
func (mgr *VDEManager) CreateVDE(name, dbUser, dbPassword string, port int) error {

	if err := mgr.createVDEDB(name, dbUser, dbPassword); err != nil {
		return err
	}

	if err := mgr.initVDEDir(name); err != nil {
		return err
	}

	vdeDir := path.Join(childConfigsPath, name)
	privFile := filepath.Join(vdeDir, consts.PrivateKeyFilename)
	pubFile := filepath.Join(vdeDir, consts.PublicKeyFilename)
	_, _, err := createKeyPair(privFile, pubFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("error on creating keys")
		return err
	}

	vdeConfigPath := filepath.Join(vdeDir, consts.DefaultConfigFile)
	vdeConfig := conf.Config
	vdeConfig.WorkDir = vdeDir
	vdeConfig.DB.User = dbUser
	vdeConfig.DB.Password = dbPassword
	vdeConfig.DB.Name = name
	vdeConfig.HTTP.Port = port

	if err := conf.SaveConfigByPath(vdeConfig, vdeConfigPath); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("saving VDE config")
		return err
	}

	confEntry := pConf.NewConfigEntry(vdeDir)
	confEntry.Name = "program:" + name
	command := fmt.Sprintf("%s -VDEMode=true -initDatabase=true -configPath=%s -workDir=%s", bin(), vdeConfigPath, vdeDir)
	confEntry.AddKeyValue("command", command)
	proc := process.NewProcess("vdeMaster", confEntry)

	mgr.processes.Add(name, proc)
	mgr.processes.Find(name).Start(true)
	log.Infoln(command)
	return nil
}

// ListProcess returns list of process names with state of process
func (mgr *VDEManager) ListProcess() map[string]string {
	list := make(map[string]string)

	mgr.processes.ForEachProcess(func(p *process.Process) {
		list[p.GetName()] = p.GetState().String()
	})

	return list
}

// DeleteVDE stop VDE process and remove VDE folder
func (mgr *VDEManager) DeleteVDE(name string) error {
	p := mgr.processes.Find(name)
	if p != nil {
		p.Stop(true)
	}

	vdeDir := path.Join(childConfigsPath, name)
	vdeConfigPath := filepath.Join(vdeDir, consts.DefaultConfigFile)
	vdeConfig, err := conf.GetConfigFromPath(vdeConfigPath)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Errorf("Getting config from path %s", vdeConfigPath)
		return err
	}

	dropDBquery := fmt.Sprintf(dropDBTemplate, vdeConfig.DB.User)
	if err := model.DBConn.Exec(dropDBquery).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Deleting vde db")
		return err
	}

	dropVDERoleQuery := fmt.Sprintf(dropDBRoleTemplate, vdeConfig.DB.User)
	if err := model.DBConn.Exec(dropVDERoleQuery).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Deleting vde db user")
		return err
	}

	return os.RemoveAll(vdeDir)
}

// StartVDE find process and then start him
func (mgr *VDEManager) StartVDE(name string) error {

	proc := mgr.processes.Find(name)
	if proc == nil {
		err := fmt.Errorf(`VDE '%s' is not exists`, name)
		log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err}).Error("on find VDE process")
		return err
	}

	state := proc.GetState()
	if state == process.STOPPED ||
		state == process.EXITED ||
		state == process.FATAL {
		proc.Start(true)
		log.Infof("VDE '%s' is started", name)
		return nil
	}

	err := fmt.Errorf("VDE '%s' is %s", name, state)
	log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err}).Error("on starting VDE")
	return err
}

// StopVDE find process with definded name and then stop him
func (mgr *VDEManager) StopVDE(name string) error {
	proc := mgr.processes.Find(name)
	if proc == nil {
		err := fmt.Errorf(`VDE '%s' is not exists`, name)
		log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err}).Error("on find VDE process")
		return err
	}

	state := proc.GetState()
	if state == process.RUNNING ||
		state == process.STARTING {
		proc.Stop(true)
		log.Infof("VDE '%s' is stoped", name)
		return nil
	}

	err := fmt.Errorf("VDE '%s' is %s", name, state)
	log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err}).Error("on stoping VDE")
	return err
}

func (mgr *VDEManager) createVDEDB(vdeName, login, pass string) error {

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

func (mgr *VDEManager) initVDEDir(vdeName string) error {

	vdeDirName := path.Join(childConfigsPath, vdeName)
	if _, err := os.Stat(vdeDirName); os.IsNotExist(err) {
		if err := os.Mkdir(vdeDirName, 0700); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating VDE directory")
			return err
		}
	}

	return nil
}

func initProcessManager() error {
	Manager = &VDEManager{
		processes: process.NewProcessManager(),
	}

	list, err := ioutil.ReadDir(childConfigsPath)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err, "path": childConfigsPath}).Error("Initialising VDE list")
		return err
	}

	for _, item := range list {
		if item.IsDir() {
			procDir := path.Join(childConfigsPath, item.Name())
			commandStr := fmt.Sprintf(commandTemplate, bin(), filepath.Join(procDir, consts.DefaultConfigFile), procDir)
			log.Errorln("commandStr: ", commandStr)
			confEntry := pConf.NewConfigEntry(procDir)
			confEntry.Name = "program:" + item.Name()
			confEntry.AddKeyValue("command", commandStr)
			confEntry.AddKeyValue("redirect_stderr", "true")
			confEntry.AddKeyValue("autostart", "true")
			confEntry.AddKeyValue("autorestart", "true")

			proc := process.NewProcess("vdeMaster", confEntry)
			Manager.processes.Add(item.Name(), proc)
		}
	}

	return nil
}

func createKeyPair(privFilename, pubFilename string) (priv, pub []byte, err error) {
	priv, pub, err = crypto.GenBytesKeys()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("generate keys")
		return
	}

	err = createFile(privFilename, []byte(hex.EncodeToString(priv)))
	if err != nil {
		return
	}

	err = createFile(pubFilename, []byte(hex.EncodeToString(pub)))
	if err != nil {
		return
	}

	return
}

func createFile(filename string, data []byte) error {
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing file")
		return err
	}
	return nil
}

func bin() string {
	return path.Join(conf.Config.WorkDir, "go-genesis")
}
