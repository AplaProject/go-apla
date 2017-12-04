package conf

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/BurntSushi/toml"
)

const defaultConfigFile = "config.toml"

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
	Type     string
	Name     string
	Host     string
	Port     int
	User     string
	Password string
}

// StatsDConfig statd connection parameters
type StatsDConfig struct {
	Name     string // default "apla"
	HostPort        // 127.0.0.1:8125
}

// Centrifugo connection params
type CentrifugoConfig struct {
	Secret string
	URL    string
}

// SavedConfig parameters saved in "config.toml"
type SavedConfig struct {
	LogLevel    string // DEBUG, INFO, WARN, ERROR
	LogFileName string // log file name relative to cwd or empty for stdout
	InstallType string
	NodeStateID string // default "*"
	//
	TestMode bool // ??? used in daemons/confirmations
	//
	StartDaemons string // comma separated list of daemons to start or empty for all or 'null'
	//
	KeyID       int64
	EcosystemID int64
	//
	BadBlocks              string // ??? accessed once as json map
	FirstLoadBlockchainURL string // ??? install -> blocks_colletcion
	FirstLoadBlockchain    string // ??? install -> blocks_collection == 'file'
	//
	Daemon HostPort
	HTTP   HostPort
	DB     DBConfig
	StatsD StatsDConfig
	//
	WorkDir        string // application work dir (cwd by default)
	FirstBlockPath string // path to the first block file
	PrivateDir     string // place for private keys files: NodePrivateKey, PrivateKey
	//
	Centrifugo CentrifugoConfig
}

// WebInstall web UI installation mode
var WebInstall bool

// Config - global immutable parameters
var Config = *initialValues()

func initialValues() *SavedConfig {
	cwd, _ := os.Getwd()

	return &SavedConfig{
		LogLevel:     "INFO",
		InstallType:  "PRIVATE_NET", // ??? PUBLIC_NET ???
		NodeStateID:  "*",
		StartDaemons: "",
		//
		Daemon: HostPort{Host: "127.0.0.1", Port: 7078},
		HTTP:   HostPort{Host: "127.0.0.1", Port: 7079},
		StatsD: StatsDConfig{Name: "apla", HostPort: HostPort{Host: "127.0.0.1", Port: 8125}},

		DB: DBConfig{
			Type:     "postgresql", // the only one supported
			Name:     "apla",
			Host:     "127.0.0.1",
			Port:     5432,
			User:     "",
			Password: "",
		},

		WorkDir: cwd,
		//
		FirstBlockPath: "",
		PrivateDir:     "",

		Centrifugo: CentrifugoConfig{
			Secret: "",
			URL:    "",
		},
	}

}

// GetConfigPath returns path from command line arg or default
func GetConfigPath() string {
	if *FlagConfigPath != "" {
		return *FlagConfigPath
	}
	return defaultConfigFile
}

// GetPidFile returns path to pid file
func GetPidFile() string {
	return filepath.Join(Config.WorkDir, "apla.pid")
}

// LoadConfig from configFile
// the function has side effect updating global var Config
func LoadConfig() error {
	_, err := toml.DecodeFile(GetConfigPath(), &Config)
	return err
}

// SaveConfig save global parameters to configFile
func SaveConfig() error {
	cf, err := os.Create(GetConfigPath())
	if err != nil {
		return err
	}
	defer cf.Close()
	return toml.NewEncoder(cf).Encode(Config)
}

// NoConfig - config file does not exist
func NoConfig() bool {
	_, err := os.Stat(GetConfigPath())
	return os.IsNotExist(err)
}

// MergeFlags override default config values by environment or args
func MergeFlags() {

	if *FlagDbName != "" {
		Config.DB.Name = *FlagDbName
	}

	if *FlagDbHost != "" {
		Config.DB.Host = *FlagDbHost
	}

	if *FlagDbPort != 0 {
		Config.DB.Port = *FlagDbPort
	}

	if *FlagDbUser != "" {
		Config.DB.User = *FlagDbUser
	}

	p := os.Getenv("PG_PASSWORD")
	if p != "" {
		Config.DB.Password = p
	}
	if *FlagDbPassword != "" {
		Config.DB.Password = *FlagDbPassword
	}

	// tcp
	if *FlagTCPHost != "" {
		Config.Daemon.Host = *FlagTCPHost
	}
	if *FlagTCPPort != 0 {
		Config.Daemon.Port = *FlagTCPPort
	}

	// http
	if *FlagHTTPHost != "" {
		Config.HTTP.Host = *FlagHTTPHost
	}
	if *FlagHTTPPort != 0 {
		Config.HTTP.Port = *FlagHTTPPort
	}

	if *FlagWorkDir != "" {
		Config.WorkDir = *FlagWorkDir
	} else if *FlagDir != "" {
		// NOTE: old flag deprecated!
		Config.WorkDir = *FlagDir
	}

	if *FlagFirstBlockPath == "" {
		Config.FirstBlockPath = filepath.Join(Config.WorkDir, "1block")
	} else {
		Config.FirstBlockPath = *FlagFirstBlockPath
	}

	if *FlagPrivateDir != "" {
		Config.PrivateDir = *FlagPrivateDir
	} else {
		Config.PrivateDir = Config.WorkDir
	}

	if *FlagLogLevel != "" {
		Config.LogLevel = *FlagLogLevel
	}
}
