package conf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/GenesisKernel/go-genesis/packages/consts"
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
	Name     string
	Host     string // ipaddr, hostname, or "0.0.0.0"
	Port     int    // must be in range 1..65535
	User     string
	Password string
}

// StatsDConfig statd connection parameters
type StatsDConfig struct {
	Host string // ipaddr, hostname, or "0.0.0.0"
	Port int    // must be in range 1..65535
	Name string
}

// CentrifugoConfig connection params
type CentrifugoConfig struct {
	Secret string
	URL    string
}

type Syslog struct {
	Facility string
	Tag      string
}

type LogConfig struct {
	LogTo     string
	LogLevel  string
	LogFormat string
	Syslog    Syslog
}

// TokenMovementConfig smtp config for token movement
type TokenMovementConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	To       string
	From     string
	Subject  string
}

// GlobalConfig is storing all startup config as global struct
type GlobalConfig struct {
	Installed    bool   `toml:"-"`
	KeyID        int64  `toml:"-"`
	ConfigPath   string `toml:"-"`
	TestRollBack bool   `toml:"-"`

	PidFilePath    string
	LockFilePath   string
	DataDir        string // application work dir (cwd by default)
	KeysDir        string // place for private keys files: NodePrivateKey, PrivateKey
	TempDir        string // temporary dir
	FirstBlockPath string
	TLS            bool   // TLS is on/off. It is required for https
	TLSCert        string // TLSCert is a filepath of the fullchain of certificate.
	TLSKey         string // TLSKey is a filepath of the private key.

	MaxPageGenerationTime int64 // in milliseconds

	TCPServer HostPort
	HTTP      HostPort

	DB            DBConfig
	StatsD        StatsDConfig
	Centrifugo    CentrifugoConfig
	LogConfig     LogConfig
	TokenMovement TokenMovementConfig

	NodesAddr []string
}

// Config global parameters
var Config GlobalConfig

// GetPidPath returns path to pid file
func (c *GlobalConfig) GetPidPath() string {
	return c.PidFilePath
}

// LoadConfig from configFile
// the function has side effect updating global var Config
func LoadConfig(path string) error {
	log.WithFields(log.Fields{"path": path}).Info("Loading config")

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.Errorf("Unable to load config file %s", path)
	}

	viper.SetConfigFile(path)
	err = viper.ReadInConfig()
	if err != nil {
		return errors.Wrapf(err, "reading config")
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		return errors.Wrapf(err, "marshalling config to global struct variable")
	}

	return nil
}

// SaveConfig save global parameters to configFile
func SaveConfig(path string) error {
	if err := makeDir(filepath.Dir(path)); err != nil {
		return err
	}

	cf, err := os.Create(path)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Create config file failed")
		return err
	}
	defer cf.Close()

	err = toml.NewEncoder(cf).Encode(Config)
	if err != nil {
		return err
	}
	return nil
}

func FillRuntimePaths() error {
	if Config.DataDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.Wrapf(err, "getting current wd")
		}

		Config.DataDir = filepath.Join(cwd, consts.DefaultWorkdirName)
	}

	if Config.KeysDir == "" {
		Config.KeysDir = Config.DataDir
	}

	if Config.TempDir == "" {
		Config.TempDir = path.Join(os.TempDir(), consts.DefaultTempDirName)
	}

	if Config.FirstBlockPath == "" {
		Config.FirstBlockPath = filepath.Join(Config.DataDir, consts.FirstBlockFilename)
	}

	if Config.PidFilePath == "" {
		Config.PidFilePath = filepath.Join(Config.DataDir, consts.DefaultPidFilename)
	}

	if Config.LockFilePath == "" {
		Config.LockFilePath = filepath.Join(Config.DataDir, consts.DefaultLockFilename)
	}

	return nil
}

func FillRuntimeKey() error {
	keyIDFileName := filepath.Join(Config.KeysDir, consts.KeyIDFilename)
	keyIDBytes, err := ioutil.ReadFile(keyIDFileName)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err, "path": keyIDFileName}).Fatal("reading KeyID file")
		return err
	}

	Config.KeyID, err = strconv.ParseInt(string(keyIDBytes), 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": string(keyIDBytes)}).Fatal("converting keyID to int")
		return errors.New("converting keyID to int")
	}

	return nil
}

func GetNodesAddr() []string {
	return Config.NodesAddr[:]
}

func MakeDirs() error {
	dirs := []string{
		Config.DataDir,
		Config.KeysDir,
		Config.TempDir,
	}

	for _, dir := range dirs {
		err := makeDir(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0775)
		if err != nil {
			return errors.Wrapf(err, "creating dir %s", dir)
		}
	}

	return nil
}
