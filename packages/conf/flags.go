package conf

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
)

// default flag values
const (
	defaultTCPHost  = "127.0.0.1"
	defaultTCPPort  = 7078
	defaultHTTPHost = "127.0.0.1"
	defaultHTTPPort = 7079

	defaultDBName = "genesis"
	defaultDBHost = "127.0.0.1"
	defaultDBPort = 5432
)

type flagBase struct {
	env  string
	help string
}

type flagStr struct {
	flagBase
	confVar *string
	flagVar string
	defVal  string
}

type flagInt struct {
	flagBase
	confVar *int
	flagVar int
	defVal  int
}

var configFlagMap = map[string]interface{}{
	"tcpHost":  &flagStr{confVar: &Config.TCPServer.Host, defVal: defaultTCPHost, flagBase: flagBase{help: "tcp server host"}},
	"tcpPort":  &flagInt{confVar: &Config.TCPServer.Port, defVal: defaultTCPPort, flagBase: flagBase{help: "tcp server port"}},
	"httpHost": &flagStr{confVar: &Config.HTTP.Host, defVal: defaultHTTPHost, flagBase: flagBase{help: "http server host"}},
	"httpPort": &flagInt{confVar: &Config.HTTP.Port, defVal: defaultHTTPPort, flagBase: flagBase{help: "http server port"}},

	"dbName":     &flagStr{confVar: &Config.DB.Name, defVal: defaultDBName, flagBase: flagBase{env: "PGDATABASE", help: "database name"}},
	"dbHost":     &flagStr{confVar: &Config.DB.Host, defVal: defaultDBHost, flagBase: flagBase{env: "PGHOST", help: "database host"}},
	"dbPort":     &flagInt{confVar: &Config.DB.Port, defVal: defaultDBPort, flagBase: flagBase{env: "PGPORT", help: "database port"}},
	"dbUser":     &flagStr{confVar: &Config.DB.User, flagBase: flagBase{env: "PGUSER", help: "database user"}},
	"dbPassword": &flagStr{confVar: &Config.DB.Password, flagBase: flagBase{env: "PGPASSWORD", help: "database password"}},

	"logLevel": &flagStr{confVar: &Config.LogConfig.LogLevel, defVal: "ERROR", flagBase: flagBase{help: "log level - ERROR,WARN,INFO,DEBUG"}},
	"logTo":    &flagStr{confVar: &Config.LogConfig.LogTo, flagBase: flagBase{help: "log to"}},
	"keysDir":  &flagStr{confVar: &Config.KeysDir, flagBase: flagBase{help: "directory for public/private keys"}},
}

var (
	// ConfigPath path to config file
	ConfigPath = flag.String("configPath", "", "full path to config file (toml format)")

	// WorkDirectory application working directory
	WorkDirectory = flag.String("dataDir", "", "work directory")

	// InitConfig initialize config
	CreateConfig = flag.Bool("createConfig", false, "write config parameters to file")

	// FirstBlockPath is a file (1block) where first block file will be stored
	FirstBlockPath = flag.String("firstBlockPath", "", "pathname of '1block' file")

	// keyID wallet id
	keyID = flag.Int64("keyID", 0, "wallet id")

	// TestRollBack starts special set of daemons
	TestRollBack = flag.Bool("testRollBack", false, "starts special set of daemons")

	// TLS is a directory for .well-known and keys. It is required for https
	TLS = flag.String("tls", "", "Enable https. Ddirectory for .well-known and keys")
)

func envStr(envName string, val *string) bool {
	if env, ok := os.LookupEnv(envName); ok {
		*val = env
		return true
	}
	return false
}

func envInt(envName string, val *int) bool {
	var strval string
	if !envStr(envName, &strval) {
		return false
	}
	i, err := strconv.Atoi(strval)
	if err != nil {
		log.WithFields(
			log.Fields{"type": consts.ConfigError, "envName": envName, "error": err},
		).Error("Incorrect value in environment")
		return false
	}
	*val = i
	return true
}

// InitConfigFlags initialize config flags
func InitConfigFlags() {
	for name, paramsPtr := range configFlagMap {
		switch flagParams := paramsPtr.(type) {
		case *flagStr:
			*flagParams.confVar = flagParams.defVal
			envStr(flagParams.env, flagParams.confVar)
			flag.StringVar(&flagParams.flagVar, name, flagParams.defVal, flagParams.help)
		case *flagInt:
			*flagParams.confVar = flagParams.defVal
			envInt(flagParams.env, flagParams.confVar)
			flag.IntVar(&flagParams.flagVar, name, flagParams.defVal, flagParams.help)
		default:
			log.WithFields(log.Fields{
				"type": consts.ConfigError, "flag": name,
			}).Error("Unexpected type in configFlagMap")
			os.Exit(1)
		}
	}
	flag.Parse()
}

// SetConfigParams set config parameters from command line
func SetConfigParams() {
	flag.Visit(func(f *flag.Flag) {
		paramsPtr, ok := configFlagMap[f.Name]
		if ok {
			switch flagParams := paramsPtr.(type) {
			case *flagStr:
				fp := *flagParams
				*fp.confVar = fp.flagVar
			case *flagInt:
				fp := *flagParams
				*fp.confVar = fp.flagVar
			}
		}
	})

	if *WorkDirectory != "" {
		Config.WorkDir = *WorkDirectory
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("Getcwd failed")
		}
		Config.WorkDir = cwd
	}

	if Config.KeysDir == "" {
		Config.KeysDir = Config.WorkDir
	}

	if Config.KeyID == 0 {
		keyIDFileName := filepath.Join(Config.KeysDir, consts.KeyIDFilename)
		if keyIDBytes, err := ioutil.ReadFile(keyIDFileName); err != nil {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err, "path": keyIDFileName}).Fatal("reading KeyID file")
		} else {
			if keyID, err := strconv.ParseInt(string(keyIDBytes), 10, 64); err != nil {
				log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": string(keyIDBytes)}).Fatal("converting keyID to int")
			} else {
				Config.KeyID = keyID
			}
		}
	}

	if *FirstBlockPath == "" {
		*FirstBlockPath = filepath.Join(Config.KeysDir, consts.FirstBlockFilename)
	}

	if *keyID != 0 {
		Config.KeyID = *keyID
	}
}
