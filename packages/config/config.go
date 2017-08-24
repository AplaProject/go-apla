package config

import (
	"io/ioutil"
	"os"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/astaxie/beego/config"
	"github.com/op/go-logging"
)

var (
	ConfigIni map[string]string
	logger    = logging.MustGetLogger("config")
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
		return err
	} else {
		ConfigIni, err = fullConfigIni.GetSection("default")
		if err != nil {
			return err
		}
		*utils.Version2 = len(ConfigIni[`version2`]) > 0
	}
	return nil
}

func Save(logLevel, installType string, dbConf *DBConfig) error {
	path := *utils.Dir + "/" + configFileName
	if _, err := os.Stat(path); os.IsNotExist(err) {
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
	if *utils.Version2 {
		confIni.Set("version2", `true`)
	}
	confIni.Set("db_password", dbConf.Password)
	confIni.Set("db_name", dbConf.Name)
	confIni.Set("node_state_id", `*`)

	err = confIni.SaveConfigFile(path)
	if err != nil {
		Drop()
		return utils.ErrInfo(err)
	}
	return nil
}

func Drop() {
	os.Remove(*utils.Dir + "/" + configFileName)
}
