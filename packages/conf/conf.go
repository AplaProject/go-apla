package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AplaProject/go-apla/packages/consts"
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
	Name     string // default "apla"
	HostPort        // 127.0.0.1:8125
}

// CentrifugoConfig connection params
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
	TCPServer HostPort
	HTTP      HostPort
	DB        DBConfig
	StatsD    StatsDConfig
	//
	WorkDir    string // application work dir (cwd by default)
	PrivateDir string // place for private keys files: NodePrivateKey, PrivateKey
	//
	Centrifugo CentrifugoConfig
}

// Installed web UI installation mode
var Installed bool

// Config - global immutable parameters
var Config = *initialValues()

func initialValues() *SavedConfig {
	cwd, err := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Getcwd failed")
	}

	return &SavedConfig{
		LogLevel:     "INFO",
		InstallType:  "PRIVATE_NET", // ??? PUBLIC_NET ???
		NodeStateID:  "*",
		StartDaemons: "",
		//
		TCPServer: HostPort{Host: "127.0.0.1", Port: 7078},
		HTTP:      HostPort{Host: "127.0.0.1", Port: 7079},
		StatsD:    StatsDConfig{Name: "apla", HostPort: HostPort{Host: "127.0.0.1", Port: 8125}},

		DB: DBConfig{
			Name:     "apla",
			HostPort: HostPort{Host: "127.0.0.1", Port: 5432},
			User:     "",
			Password: "",
		},

		WorkDir:    cwd,
		PrivateDir: "",

		Centrifugo: CentrifugoConfig{
			Secret: "",
			URL:    "",
		},
	}

}

// GetConfigPath returns path from command line arg or default
func GetConfigPath() string {
	if *ConfigPath != "" {
		return *ConfigPath
	}
	return filepath.Join(*FlagWorkDir, consts.DefaultConfigFile)
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

// NoConfig - config file does not exist
func NoConfig() bool {
	_, err := os.Stat(GetConfigPath())
	return os.IsNotExist(err)
}

func flagOrEnv(confValue *string, flagValue string, envName string) {
	if flagValue != "" {
		*confValue = flagValue
		return
	}
	if env, ok := os.LookupEnv(envName); ok {
		*confValue = env
	}
}

func intFlagOrEnv(confValue *int, flagValue int, envName string) {
	if flagValue != 0 {
		*confValue = flagValue
		return
	}
	if env, ok := os.LookupEnv(envName); ok {
		i, err := strconv.Atoi(env)
		if err != nil {
			log.WithFields(
				log.Fields{"type": consts.ConfigError, "envName": envName, "error": err},
			).Error("Incorrect value in environment: " + envName)
			return
		}
		*confValue = i
	}
}

// OverrideFlags override default config values by environment or args
func OverrideFlags() {

	flagOrEnv(&Config.DB.Name, *FlagDbName, "PGDATABASE")
	flagOrEnv(&Config.DB.Host, *FlagDbHost, "PGHOST")
	intFlagOrEnv(&Config.DB.Port, *FlagDbPort, "PGPORT")
	flagOrEnv(&Config.DB.User, *FlagDbUser, "PGUSER")
	flagOrEnv(&Config.DB.Password, *FlagDbPassword, "PGPASSWORD")

	flagOrEnv(&Config.TCPServer.Host, *FlagTCPHost, "")
	intFlagOrEnv(&Config.TCPServer.Port, *FlagTCPPort, "")

	flagOrEnv(&Config.HTTP.Host, *FlagHTTPHost, "")
	intFlagOrEnv(&Config.HTTP.Port, *FlagHTTPPort, "")

	flagOrEnv(&Config.LogLevel, *FlagLogLevel, "")
	flagOrEnv(&Config.LogFileName, *FlagLogFile, "")

	flagOrEnv(&Config.WorkDir, *FlagWorkDir, "")
	flagOrEnv(&Config.PrivateDir, *FlagPrivateDir, "")

	if *FlagKeyID != 0 {
		Config.KeyID = *FlagKeyID
	}

	if Config.PrivateDir == "" {
		Config.PrivateDir = Config.WorkDir
	}

	if *FirstBlockPath == "" {
		*FirstBlockPath = filepath.Join(Config.PrivateDir, consts.FirstBlockFilename)
	}
}
