package config

import (
	"github.com/BurntSushi/toml"
)

type Daemon struct {
	DaemonMode     bool `toml:"daemon"`
	QueryingPeriod int  `toml:"querying_period"`
}

type AlertMessage struct {
	To      string `toml:"to"`
	From    string `toml:"from"`
	Subject string `toml:"subject"`
}

type Smtp struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type Config struct {
	Daemon       Daemon       `toml:"daemon"`
	AlertMessage AlertMessage `toml:"alert_message"`
	Smtp         Smtp         `toml:"smtp"`
	NodesList    []string     `toml:"nodes_list"`
}

func (c *Config) Read(fileName string) error {
	_, err := toml.DecodeFile(fileName, c)
	return err
}
