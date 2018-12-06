package system

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"

	log "github.com/sirupsen/logrus"
)

// CreatePidFile creats pid file
func CreatePidFile() error {
	pid := os.Getpid()
	data := []byte(strconv.Itoa(pid))
	return ioutil.WriteFile(conf.Config.GetPidPath(), data, 0644)
}

// RemovePidFile removes pid file
func RemovePidFile() error {
	return os.Remove(conf.Config.GetPidPath())
}

// ReadPidFile reads pid file
func ReadPidFile() (int, error) {
	pidPath := conf.Config.GetPidPath()
	if _, err := os.Stat(pidPath); err != nil {
		return 0, nil
	}

	data, err := ioutil.ReadFile(pidPath)
	if err != nil {
		log.WithFields(log.Fields{"path": pidPath, "error": err, "type": consts.IOError}).Error("reading pid file")
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		log.WithFields(log.Fields{"data": data, "error": err, "type": consts.ConversionError}).Error("pid file data to int")
	}
	return pid, err
}
