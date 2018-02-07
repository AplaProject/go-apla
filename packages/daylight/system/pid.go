package system

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/GenesisCommunity/go-genesis/packages/conf"
	"github.com/GenesisCommunity/go-genesis/packages/consts"

	log "github.com/sirupsen/logrus"
)

func CreatePidFile() error {
	pid := os.Getpid()
	data := []byte(strconv.Itoa(pid))
	return ioutil.WriteFile(conf.GetPidFile(), data, 0644)
}

func RemovePidFile() error {
	return os.Remove(conf.GetPidFile())
}

func ReadPidFile() (int, error) {
	pidPath := conf.GetPidFile()
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
