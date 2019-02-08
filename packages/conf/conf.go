// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package conf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/AplaProject/go-apla/packages/consts"
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
	Name        string
	Host        string // ipaddr, hostname, or "0.0.0.0"
	Port        int    // must be in range 1..65535
	User        string
	Password    string
	LockTimeout int // lock_timeout in milliseconds
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

// Syslog represents parameters of syslog
type Syslog struct {
	Facility string
	Tag      string
}

// Log represents parameters of log
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
	KeyID        int64  `toml:"-"`
	ConfigPath   string `toml:"-"`
	TestRollBack bool   `toml:"-"`
	FuncBench    bool   `toml:"-"`

	PidFilePath           string
	LockFilePath          string
	DataDir               string // application work dir (cwd by default)
	KeysDir               string // place for private keys files: NodePrivateKey, PrivateKey
	TempDir               string // temporary dir
	FirstBlockPath        string
	TLS                   bool   // TLS is on/off. It is required for https
	TLSCert               string // TLSCert is a filepath of the fullchain of certificate.
	TLSKey                string // TLSKey is a filepath of the private key.
	OBSMode               string
	HTTPServerMaxBodySize int64

	MaxPageGenerationTime int64 // in milliseconds

	TCPServer HostPort
	HTTP      HostPort

	DB            DBConfig
	StatsD        StatsDConfig
	Centrifugo    CentrifugoConfig
	Log           LogConfig
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
	return LoadConfigToVar(path, &Config)
}

func LoadConfigToVar(path string, v *GlobalConfig) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.Errorf("Unable to load config file %s", path)
	}

	viper.SetConfigFile(path)
	err = viper.ReadInConfig()
	if err != nil {
		return errors.Wrapf(err, "reading config")
	}

	err = viper.Unmarshal(v)
	if err != nil {
		return errors.Wrapf(err, "marshalling config to global struct variable")
	}
	return nil
}

// GetConfigFromPath read config from path and returns GlobalConfig struct
func GetConfigFromPath(path string) (*GlobalConfig, error) {
	log.WithFields(log.Fields{"path": path}).Info("Loading config")

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, errors.Errorf("Unable to load config file %s", path)
	}

	viper.SetConfigFile(path)
	err = viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "reading config")
	}

	c := &GlobalConfig{}
	err = viper.Unmarshal(c)
	if err != nil {
		return c, errors.Wrapf(err, "marshalling config to global struct variable")
	}

	return c, nil
}

// SaveConfig save global parameters to configFile
func SaveConfig(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0775)
		if err != nil {
			return errors.Wrapf(err, "creating dir %s", dir)
		}
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

// FillRuntimePaths fills paths from runtime parameters
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
		Config.TempDir = filepath.Join(os.TempDir(), consts.DefaultTempDirName)
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

// FillRuntimeKey fills parameters of keys from runtime parameters
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

// GetNodesAddr returns addreses of nodes
func GetNodesAddr() []string {
	return Config.NodesAddr[:]
}

// IsOBS check running mode
func (c GlobalConfig) IsOBS() bool {
	return RunMode(c.OBSMode).IsOBS()
}

// IsOBSMaster check running mode
func (c GlobalConfig) IsOBSMaster() bool {
	return RunMode(c.OBSMode).IsOBSMaster()
}

// IsSupportingOBS check running mode
func (c GlobalConfig) IsSupportingOBS() bool {
	return RunMode(c.OBSMode).IsSupportingOBS()
}

// IsNode check running mode
func (c GlobalConfig) IsNode() bool {
	return RunMode(c.OBSMode).IsNode()
}
