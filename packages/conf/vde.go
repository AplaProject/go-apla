package conf

import (
	"os"

	toml "github.com/BurntSushi/toml"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
)

// LoadVDEConfig from configFile
func LoadVDEConfig(config interface{}) error {
	log.WithFields(log.Fields{"path": GetConfigPath()}).Info("Loading config")
	if _, err := toml.DecodeFile(GetConfigPath(), config); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("LoadConfig")
		return err
	}
	return nil
}

// SaveVDEConfig save global parameters to configFile
func SaveVDEConfig(path string, config interface{}) error {
	cf, err := os.Create(path)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Create config file failed")
		return err
	}
	defer cf.Close()
	return toml.NewEncoder(cf).Encode(config)
}
