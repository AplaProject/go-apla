package conf

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/BurntSushi/toml"
	"github.com/GenesisKernel/go-genesis/packages/consts"
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

// TokenMovementConfig smtp config for token movement
type TokenMovementConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	To       string `toml:"to"`
	From     string `toml:"from"`
	Subject  string `toml:"subject"`
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

	TokenMovement TokenMovementConfig
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

// LoadConfig load config from default path
func LoadConfig() error {
	return LoadConfigFromPath(GetConfigPath())
}

// LoadConfigFromPath from configFile
// the function has side effect updating global var Config
func LoadConfigFromPath(path string) error {
	log.WithFields(log.Fields{"path": path}).Info("Loading config")
	_, err := toml.DecodeFile(path, &Config)
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

// SaveConfigByPath save config by path
func SaveConfigByPath(c SavedConfig, path string) error {
	var cf *os.File
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		cf, err = os.Create(path)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Create config file failed")
			return err
		}
	} else {
		cf, err = os.Open(path)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Create config file failed")
			return err
		}
	}

	defer cf.Close()
	return toml.NewEncoder(cf).Encode(c)
}

// NoConfig config file does not exist
func NoConfig() bool {
	_, err := os.Stat(GetConfigPath())
	return os.IsNotExist(err)
}
