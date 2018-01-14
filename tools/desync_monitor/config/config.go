package config

import (
	"github.com/BurntSushi/toml"
)

type Daemon struct {
	DaemonMode     bool `toml:"daemon"`
	QueryingPeriod int  `toml:"querying_period"`
}

type Alert struct {
	Email   string `toml:"email"`
	Message string `toml:"message"`
}

type Config struct {
	Daemon    Daemon   `toml:"daemon"`
	Alert     Alert    `toml:"alert"`
	NodesList []string `toml:"nodes_list"`
}

func (c *Config) Read(fileName string) error {
	_, err := toml.DecodeFile(fileName, c)
	return err
}
