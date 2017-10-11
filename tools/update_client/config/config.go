package config

import "github.com/astaxie/beego/config"

const configFileName = "config.ini"

type Config struct {
	Login      string
	Pass       string
	Host       string
	Port       string
	DBPath     string
	PubkeyPath string
}

var ConfigIni map[string]string

func (dbc *Config) Read() error {
	fullConfigIni, err := config.NewConfig("ini", "./config.ini")
	if err != nil {
		return err
	}
	configIni, err := fullConfigIni.GetSection("default")
	if err != nil {
		return err
	}
	dbc.Login = configIni["login"]
	dbc.Pass = configIni["pass"]
	dbc.Host = configIni["host"]
	dbc.Port = configIni["port"]
	dbc.DBPath = configIni["dbPath"]
	dbc.PubkeyPath = configIni["pubkeyPath"]
	return nil
}
