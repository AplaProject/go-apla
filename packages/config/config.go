package config

import (
	"io/ioutil"
	"os"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/astaxie/beego/config"
)

var (
	ConfigIni map[string]string
)

const configFileName = "config.ini"

type DBConfig struct {
	Type     string
	User     string
	Host     string
	Port     string
	Password string
	Name     string
}

func Read() error {
	ConfigIni = map[string]string{}
	fullConfigIni, err := config.NewConfig("ini", *utils.Dir+"/"+configFileName)
	if err != nil {
		logger.LogError(consts.ConfigError, err)
		return err
	} else {
		ConfigIni, err = fullConfigIni.GetSection("default")
		if err != nil {
			logger.LogError(consts.ConfigError, err)
			return err
		}
	}
	return nil
}

func Save(logLevel, installType string, dbConf *DBConfig) error {
	logger.LogDebug(consts.FuncStarted, "")
	path := *utils.Dir + "/" + configFileName
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.LogError(consts.IOError, err)
		ioutil.WriteFile(path, []byte(``), 0644)
	}
	confIni, err := config.NewConfig("ini", path)
	confIni.Set("log_level", logLevel)
	confIni.Set("install_type", installType)
	confIni.Set("dir", *utils.Dir)
	confIni.Set("tcp_host", *utils.TCPHost)
	confIni.Set("http_port", *utils.ListenHTTPPort)
	confIni.Set("first_block_dir", *utils.FirstBlockDir)
	confIni.Set("db_type", dbConf.Type)
	confIni.Set("db_user", dbConf.User)
	confIni.Set("db_host", dbConf.Host)
	confIni.Set("db_port", dbConf.Port)
	confIni.Set("db_password", dbConf.Password)
	confIni.Set("db_name", dbConf.Name)
	confIni.Set("node_state_id", `*`)

	err = confIni.SaveConfigFile(path)
	if err != nil {
		logger.LogError(consts.IOError, err)
		Drop()
		return utils.ErrInfo(err)
	}
	return nil
}

func Drop() {
	err := os.Remove(*utils.Dir + "/" + configFileName)
	if err != nil {
		logger.LogError(consts.IOError, err)
	}
}
