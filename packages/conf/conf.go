//MIT License
//
//Copyright (c) 2016-2018 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	toml "github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

// HostPort endpoint in form "str:int"
type HostPort struct {
	Host string // ipaddr, hostname, or "0.0.0.0"
	Port int    // must be in range 1..65535
}

// Str converts HostPort pair to string format
func (h HostPort) Str() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

// DBConfig database connection parameters
type DBConfig struct {
	Name string
	HostPort
	User     string
	Password string
}

// StatsDConfig statd connection parameters
type StatsDConfig struct {
	Name string
	HostPort
}

// CentrifugoConfig connection params
type CentrifugoConfig struct {
	Secret string
	URL    string
}

// AutoupdateConfig is autoupdate params
type AutoupdateConfig struct {
	ServerAddress string
	PublicKeyPath string
}

// SavedConfig parameters saved in "config.toml"
type SavedConfig struct {
	LogLevel    string
	LogFileName string
	InstallType string
	NodeStateID string
	TestMode    bool

	StartDaemons string // comma separated list of daemons to start or empty for all or 'null'

	KeyID       int64
	EcosystemID int64

	BadBlocks              string
	FirstLoadBlockchainURL string
	FirstLoadBlockchain    string

	MaxPageGenerationTime int64 // in milliseconds

	TCPServer HostPort
	HTTP      HostPort
	DB        DBConfig
	StatsD    StatsDConfig

	WorkDir    string // application work dir (cwd by default)
	PrivateDir string // place for private keys files: NodePrivateKey, PrivateKey

	Centrifugo CentrifugoConfig

	Autoupdate AutoupdateConfig
}

// Installed web UI installation mode
var Installed bool

// Config global parameters
var Config = SavedConfig{
	InstallType:  "PRIVATE_NET",
	NodeStateID:  "*",
	StartDaemons: "",
	StatsD:       StatsDConfig{Name: "apla", HostPort: HostPort{Host: "127.0.0.1", Port: 8125}},
}

// GetConfigPath returns path from command line arg or default
func GetConfigPath() string {
	if *ConfigPath != "" {
		return *ConfigPath
	}
	return filepath.Join(*WorkDirectory, consts.DefaultConfigFile)
}

// GetPidFile returns path to pid file
func GetPidFile() string {
	return filepath.Join(Config.WorkDir, consts.PidFilename)
}

// LoadConfig from configFile
// the function has side effect updating global var Config
func LoadConfig() error {
	log.WithFields(log.Fields{"path": GetConfigPath()}).Info("Loading config")
	_, err := toml.DecodeFile(GetConfigPath(), &Config)
	return err
}

// SaveConfig save global parameters to configFile
func SaveConfig() error {
	cf, err := os.Create(GetConfigPath())
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Create config file failed")
		return err
	}
	defer cf.Close()
	return toml.NewEncoder(cf).Encode(Config)
}

// NoConfig config file does not exist
func NoConfig() bool {
	_, err := os.Stat(GetConfigPath())
	return os.IsNotExist(err)
}
