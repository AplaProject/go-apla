// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package vdemanager

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/AplaProject/go-apla/packages/conf"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	pConf "github.com/rpoletaev/supervisord/config"
	"github.com/rpoletaev/supervisord/process"
	log "github.com/sirupsen/logrus"
)

const (
	childFolder        = "configs"
	createRoleTemplate = `CREATE ROLE %s WITH ENCRYPTED PASSWORD '%s' NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN`
	createDBTemplate   = `CREATE DATABASE %s OWNER %s`

	dropDBTemplate     = `DROP DATABASE IF EXISTS %s`
	dropOwnedTemplate  = `DROP OWNED BY %s CASCADE`
	dropDBRoleTemplate = `DROP ROLE IF EXISTS %s`
	commandTemplate    = `%s start --config=%s`

	alreadyExistsErrorTemplate = `vde '%s' already exists`
)

var (
	errWrongMode        = errors.New("node must be running as VDEMaster")
	errIncorrectVDEName = errors.New("the name cannot begit with a number and must contain alphabetical symbols and numbers")
)

// VDEManager struct
type VDEManager struct {
	processes        *process.ProcessManager
	execPath         string
	childConfigsPath string
}

var (
	Manager *VDEManager
)

func prepareWorkDir() (string, error) {
	childConfigsPath := path.Join(conf.Config.DataDir, childFolder)

	if _, err := os.Stat(childConfigsPath); os.IsNotExist(err) {
		if err := os.Mkdir(childConfigsPath, 0700); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating configs directory")
			return "", err
		}
	}

	return childConfigsPath, nil
}

// CreateVDE creates one instance of VDE
func (mgr *VDEManager) CreateVDE(name, dbUser, dbPassword string, port int) error {
	if err := checkVDEName(name); err != nil {
		log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err}).Error("on check VDE name")
		return errIncorrectVDEName
	}

	var err error
	var cancelChain []func()

	defer func() {
		if err == nil {
			return
		}

		for _, cancelFunc := range cancelChain {
			cancelFunc()
		}
	}()

	config := ChildVDEConfig{
		Executable:     mgr.execPath,
		Name:           name,
		Directory:      path.Join(mgr.childConfigsPath, name),
		DBUser:         dbUser,
		DBPassword:     dbPassword,
		ConfigFileName: consts.DefaultConfigFile,
		HTTPPort:       port,
		LogTo:          fmt.Sprintf("%s_%s", name, conf.Config.Log.LogTo),
		LogLevel:       conf.Config.Log.LogLevel,
	}

	if mgr.processes == nil {
		log.WithFields(log.Fields{"type": consts.WrongModeError, "error": errWrongMode}).Error("creating new VDE")
		return errWrongMode
	}

	if err = mgr.createVDEDB(name, dbUser, dbPassword); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on creating VDE DB")
		return fmt.Errorf(alreadyExistsErrorTemplate, name)
	}

	cancelChain = append(cancelChain, func() {
		dropDb(name, dbUser)
	})

	dirPath := path.Join(mgr.childConfigsPath, name)
	if directoryExists(dirPath) {
		err = fmt.Errorf(alreadyExistsErrorTemplate, name)
		log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err, "dirPath": dirPath}).Error("on check directory")
		return err
	}

	if err = mgr.initVDEDir(name); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "DirName": name, "error": err}).Error("on init VDE dir")
		return err
	}

	cancelChain = append(cancelChain, func() {
		dropVDEDir(mgr.childConfigsPath, name)
	})

	cmd := config.configCommand()
	if err = cmd.Run(); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "args": cmd.Args}).Error("on run config command")
		return err
	}

	if err = config.generateKeysCommand().Run(); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "args": cmd.Args}).Error("on run generateKeys command")
		return err
	}

	if err = config.initDBCommand().Run(); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "args": cmd.Args}).Error("on run initDB command")
		return err
	}

	procConfEntry := pConf.NewConfigEntry(config.Directory)
	procConfEntry.Name = "program:" + name
	command := fmt.Sprintf("%s start --config=%s", config.Executable, filepath.Join(config.Directory, consts.DefaultConfigFile))
	log.Infoln(command)
	procConfEntry.AddKeyValue("command", command)
	proc := process.NewProcess("vdeMaster", procConfEntry)

	mgr.processes.Add(name, proc)
	mgr.processes.Find(name).Start(true)
	return nil
}

// ListProcess returns list of process names with state of process
func (mgr *VDEManager) ListProcess() (map[string]string, error) {
	if mgr.processes == nil {
		log.WithFields(log.Fields{"type": consts.WrongModeError, "error": errWrongMode}).Error("get VDE list")
		return nil, errWrongMode
	}

	list := make(map[string]string)

	mgr.processes.ForEachProcess(func(p *process.Process) {
		list[p.GetName()] = p.GetState().String()
	})

	return list, nil
}

func (mgr *VDEManager) ListProcessWithPorts() (map[string]string, error) {
	list, err := mgr.ListProcess()
	if err != nil {
		return list, err
	}

	for name, status := range list {
		path := path.Join(mgr.childConfigsPath, name, consts.DefaultConfigFile)
		c := &conf.GlobalConfig{}
		if err := conf.LoadConfigToVar(path, c); err != nil {
			log.WithFields(log.Fields{"type": "dbError", "error": err, "path": path}).Warn("on loading child VDE config")
			continue
		}

		list[name] = fmt.Sprintf("%s %d", status, c.HTTP.Port)
	}

	return list, err
}

// DeleteVDE stop VDE process and remove VDE folder
func (mgr *VDEManager) DeleteVDE(name string) error {

	if mgr.processes == nil {
		log.WithFields(log.Fields{"type": consts.WrongModeError, "error": errWrongMode}).Error("deleting VDE")
		return errWrongMode
	}

	mgr.StopVDE(name)
	mgr.processes.Remove(name)
	vdeDir := path.Join(mgr.childConfigsPath, name)
	vdeConfigPath := filepath.Join(vdeDir, consts.DefaultConfigFile)
	vdeConfig, err := conf.GetConfigFromPath(vdeConfigPath)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Errorf("Getting config from path %s", vdeConfigPath)
		return fmt.Errorf(`VDE '%s' is not exists`, name)
	}

	time.Sleep(1 * time.Second)
	if err := dropDb(vdeConfig.DB.Name, vdeConfig.DB.User); err != nil {
		return err
	}

	return os.RemoveAll(vdeDir)
}

// StartVDE find process and then start him
func (mgr *VDEManager) StartVDE(name string) error {

	if mgr.processes == nil {
		log.WithFields(log.Fields{"type": consts.WrongModeError, "error": errWrongMode}).Error("starting VDE")
		return errWrongMode
	}

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
		log.WithFields(log.Fields{"vde_name": name}).Info("VDE started")
		return nil
	}

	err := fmt.Errorf("VDE '%s' is %s", name, state)
	log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err}).Error("on starting VDE")
	return err
}

// StopVDE find process with definded name and then stop him
func (mgr *VDEManager) StopVDE(name string) error {

	if mgr.processes == nil {
		log.WithFields(log.Fields{"type": consts.WrongModeError, "error": errWrongMode}).Error("on stopping VDE process")
		return errWrongMode
	}

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
		log.WithFields(log.Fields{"vde_name": name}).Info("VDE is stoped")
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

		if err := model.GetDB(nil).Exec(fmt.Sprintf(dropDBRoleTemplate, login)).Error; err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err, "role": login}).Error("on deleting vde role")
			return err
		}
		return err
	}

	return nil
}

func (mgr *VDEManager) initVDEDir(vdeName string) error {

	vdeDirName := path.Join(mgr.childConfigsPath, vdeName)
	if _, err := os.Stat(vdeDirName); os.IsNotExist(err) {
		if err := os.Mkdir(vdeDirName, 0700); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("creating VDE directory")
			return err
		}
	}

	return nil
}

func InitVDEManager() {

	execPath, err := os.Executable()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err}).Fatal("on determine executable path")
	}

	childConfigsPath, err := prepareWorkDir()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.VDEManagerError, "error": err}).Fatal("on prepare child configs folder")
	}

	Manager = &VDEManager{
		processes:        process.NewProcessManager(),
		execPath:         execPath,
		childConfigsPath: childConfigsPath,
	}

	list, err := ioutil.ReadDir(childConfigsPath)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err, "path": childConfigsPath}).Fatal("on read child VDE directory")
	}

	for _, item := range list {
		if item.IsDir() {
			procDir := path.Join(Manager.childConfigsPath, item.Name())
			commandStr := fmt.Sprintf(commandTemplate, Manager.execPath, filepath.Join(procDir, consts.DefaultConfigFile))
			log.Info(commandStr)
			confEntry := pConf.NewConfigEntry(procDir)
			confEntry.Name = "program:" + item.Name()
			confEntry.AddKeyValue("command", commandStr)
			confEntry.AddKeyValue("redirect_stderr", "true")
			confEntry.AddKeyValue("autostart", "true")
			confEntry.AddKeyValue("autorestart", "true")

			proc := process.NewProcess("vdeMaster", confEntry)
			Manager.processes.Add(item.Name(), proc)
			proc.Start(true)
		}
	}
}

func dropDb(name, role string) error {
	if err := model.DropDatabase(name); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "db_name": name}).Error("Deleting vde db")
		return err
	}

	if err := model.GetDB(nil).Exec(fmt.Sprintf(dropDBRoleTemplate, role)).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "role": role}).Error("on deleting vde role")
	}
	return nil
}

func dropVDEDir(configsPath, vdeName string) error {
	path := path.Join(configsPath, vdeName)
	if directoryExists(path) {
		os.RemoveAll(path)
	}

	log.WithFields(log.Fields{"path": path}).Error("droping dir is not exists")
	return nil
}

func directoryExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func checkVDEName(name string) error {

	name = strings.ToLower(name)

	for i, c := range name {
		if unicode.IsDigit(c) && i == 0 {
			return fmt.Errorf("the name cannot begin with a number")
		}
		if !unicode.IsDigit(c) && !unicode.Is(unicode.Latin, c) {
			return fmt.Errorf("Incorrect symbol")
		}
	}

	return nil
}

func (mgr *VDEManager) configByName(name string) (*conf.GlobalConfig, error) {
	path := path.Join(mgr.childConfigsPath)
	c := &conf.GlobalConfig{}
	err := conf.LoadConfigToVar(path, c)
	return c, err
}
