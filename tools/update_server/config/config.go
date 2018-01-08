package config

import (
	config "github.com/astaxie/beego/config"
	"github.com/pkg/errors"
)

// Config is storing web & database credentials, public key for checking sign on apla's binary
type Config struct {
	Login      string
	Pass       string
	Host       string
	Port       string
	DBPath     string
	PubkeyPath string
}

type Parser struct {
	filepath string
}

func NewParser(filepath string) Parser {
	return Parser{filepath: filepath}
}

// Do parsing config from filepath to struct
func (p *Parser) Do() (Config, error) {
	var c Config
	cc, err := config.NewConfig("ini", p.filepath)
	if err != nil {
		return c, errors.Wrapf(err, "opening %s ini config", p.filepath)
	}

	configIni, err := cc.GetSection("default")
	if err != nil {
		return c, errors.Wrapf(err, "getting default section of config")
	}

	c.Login = configIni["login"]
	c.Pass = configIni["pass"]
	c.Host = configIni["host"]
	c.Port = configIni["port"]
	c.DBPath = configIni["dbpath"]
	c.PubkeyPath = configIni["pubkeypath"]
	return c, nil
}
