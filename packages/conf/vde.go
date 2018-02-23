package conf

import (
	"os"

	toml "github.com/BurntSushi/toml"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
)

// VDEConfig config for VDE mode
type VDEConfig struct {
	DB          DBConfig
	HTTP        HostPort
	Centrifugo  CentrifugoConfig
	Autoupdate  AutoupdateConfig
	WorkDir     string
	LogLevel    string
	LogFileName string
}

// VDEMasterConfig config for VDE master mode
type VDEMasterConfig struct {
	*VDEConfig
	Login    string
	Password string
}

// LoadVDEConfig from configFile
// the function has side effect updating global var Config
func LoadVDEConfig(config interface{}) error {
	log.WithFields(log.Fields{"path": GetConfigPath()}).Info("Loading config")
	_, err := toml.DecodeFile(GetConfigPath(), config)
	return err
}

// SaveVDEConfig save global parameters to configFile
func SaveVDEConfig(config interface{}) error {
	cf, err := os.Create(GetConfigPath())
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Create config file failed")
		return err
	}
	defer cf.Close()
	return toml.NewEncoder(cf).Encode(config)
}
