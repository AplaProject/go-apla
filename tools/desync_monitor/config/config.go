// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

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
