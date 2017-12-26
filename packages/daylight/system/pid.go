package system

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"

	log "github.com/sirupsen/logrus"
)

func CreatePidFile() error {
	pid := os.Getpid()
	data, err := json.Marshal(map[string]string{
		"pid":     converter.IntToStr(pid),
		"version": consts.VERSION,
	})
	if err != nil {
		log.WithFields(log.Fields{"pid": pid, "error": err, "type": consts.JSONMarshallError}).Error("marshalling pid to json")
		return err
	}
	return ioutil.WriteFile(conf.GetPidFile(), data, 0644)
}

func RemovePidFile() error {
	return os.Remove(conf.GetPidFile())
}

func ReadPidFile() (map[string]string, error) {
	pidPath := conf.GetPidFile()
	if _, err := os.Stat(pidPath); err != nil {
		return nil, nil
	}

	data, err := ioutil.ReadFile(pidPath)
	if err != nil {
		log.WithFields(log.Fields{"path": pidPath, "error": err, "type": consts.IOError}).Error("reading pid file")
		return nil, err
	}

	var pidMap map[string]string
	err = json.Unmarshal(data, &pidMap)
	if err != nil {
		log.WithFields(log.Fields{"data": data, "error": err, "type": consts.JSONUnmarshallError}).Error("unmarshalling pid map")
	}
	return pidMap, err
}
