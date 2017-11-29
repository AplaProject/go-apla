package conf

import (
	"os"

	// go get github.com/BurntSushi/toml
	toml "github.com/BurntSushi/toml"
)

const configFileName = "config.toml"

// DBConfig database connection parameters
type DBConfig struct {
	Type     string
	Name     string
	Host     string
	Port     string
	User     string
	Password string
}

// StatsDConfig statd connection parameters
type StatsDConfig struct {
	Host string // default "127.0.0.1"
	Port string // default 8125
	Name string // default "apla"
}

// SavedConfig config parameters saved in "config.toml"
type SavedConfig struct {
	Version     string
	LogLevel    string // ERROR, INFO, WARN, DEBUG
	InstallType string
	NodeStateID string // default "*"

	TCPHost  string
	TCPPort  string // must be in range 1..65535
	HTTPHost string
	HTTPPort string // must be in range 1..65535
	DB       DBConfig
	StatsD   StatsDConfig

	WorkDir        string // application work dir (cwd by default)
	FirstBlockPath string // path to the first block file
	PrivateDir     string // place for private keys files: NodePrivateKey, PrivateKey
}

// Config - global immutable parameters
var Config = SavedConfig{
	Version:  "v0.1",
	LogLevel: "INFO",

	StatsD: StatsDConfig{
		Host: "127.0.0.1",
		Port: "8125",
		Name: "apla",
	},
}

// LoadConfig from configFileName ("config.toml")
// the function has side effect updating global var Config
func LoadConfig() error {
	_, err := toml.DecodeFile(configFileName, &Config)
	return err
}

// SaveConfig save global parameters to configFileName ("config.toml")
func SaveConfig() error {
	cf, err := os.Create(configFileName)
	if err != nil {
		return err
	}
	defer cf.Close()
	return toml.NewEncoder(cf).Encode(Config)
}

/*
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
confIni.Set("version2", `true`)
confIni.Set("db_password", dbConf.Password)
confIni.Set("db_name", dbConf.Name)
confIni.Set("node_state_id", `*`)
*/
