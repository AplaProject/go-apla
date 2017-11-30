package conf

import (
	"fmt"
	"os"

	// go get github.com/BurntSushi/toml
	toml "github.com/BurntSushi/toml"
)

const configFileName = "config.toml"

// HostPort tcp endpoint in form "host:port"
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

// SavedConfig config parameters saved in "config.toml"
type SavedConfig struct {
	LogLevel    string // ERROR, INFO, WARN, DEBUG
	InstallType string
	NodeStateID string // default "*"

	Daemon HostPort
	API    HostPort

	DB     DBConfig
	StatsD StatsDConfig

	WorkDir        string // application work dir (cwd by default)
	FirstBlockPath string // path to the first block file
	PrivateDir     string // place for private keys files: NodePrivateKey, PrivateKey
}

// Config - global immutable parameters
var Config = SavedConfig{
	LogLevel: "INFO",

	Daemon: HostPort{Host: "127.0.0.1", Port: 7078},
	API:    HostPort{Host: "127.0.0.1", Port: 7079},
	StatsD: StatsDConfig{Name: "apla", HostPort: HostPort{Host: "127.0.0.1", Port: 8125}},

	DB: DBConfig{
		Type:     "postgresql",
		Name:     "apla",
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "",
		Password: "",
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
