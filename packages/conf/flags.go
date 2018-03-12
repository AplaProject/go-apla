package conf

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
)

// default flag values
const (
	defaultTCPHost  = "127.0.0.1"
	defaultTCPPort  = 7078
	defaultHTTPHost = "127.0.0.1"
	defaultHTTPPort = 7079

	defaultDBName = "apla"
	defaultDBHost = "127.0.0.1"
	defaultDBPort = 5432

	defaultFirstBlockHost = "127.0.0.1"

	defaultUpdateInterval      = int64(time.Hour / time.Second)
	defaultUpdateServer        = "http://127.0.0.1:12345"
	defaultUpdatePublicKeyPath = "update.pub"
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

	"logLevel":   &flagStr{confVar: &Config.LogLevel, defVal: "ERROR", flagBase: flagBase{help: "log level - ERROR,WARN,INFO,DEBUG"}},
	"logFile":    &flagStr{confVar: &Config.LogFileName, flagBase: flagBase{help: "log file name"}},
	"privateDir": &flagStr{confVar: &Config.PrivateDir, flagBase: flagBase{help: "directory for public/private keys"}},

	"updateServer":        &flagStr{confVar: &Config.Autoupdate.ServerAddress, defVal: defaultUpdateServer, flagBase: flagBase{help: "server address for autoupdates"}},
	"updatePublicKeyPath": &flagStr{confVar: &Config.Autoupdate.PublicKeyPath, defVal: defaultUpdatePublicKeyPath, flagBase: flagBase{help: "public key path for autoupdates"}},
}

var (
	// ConfigPath path to config file
	ConfigPath = flag.String("configPath", "", "full path to config file (toml format)")
	NoStart    = flag.Bool("noStart", false, "do not start daemon, just do all necessary job")

	// WorkDirectory application working directory
	WorkDirectory = flag.String("workDir", "", "work directory")

	// FirstBlockPath is a file (1block) where first block file will be stored
	FirstBlockPath = flag.String("firstBlockPath", "", "pathname of '1block' file")

	// InitConfig initialize config
	InitConfig = flag.Bool("initConfig", false, "write config parameters to file")

	// InitDatabase initialize database
	InitDatabase = flag.Bool("initDatabase", false, "initialize database")

	// GenerateFirstBlock force regeneration of first block
	GenerateFirstBlock = flag.Bool("generateFirstBlock", false, "generate firstblock and keys")

	// keyID wallet id
	keyID = flag.Int64("keyID", 0, "wallet id")

	// FirstBlockPublicKey is the private key
	FirstBlockPublicKey = flag.String("firstBlockPublicKey", "", "FirstBlockPublicKey")

	// FirstBlockNodePublicKey is the node private key
	FirstBlockNodePublicKey = flag.String("firstBlockNodePublicKey", "", "FirstBlockNodePublicKey")

	// FirstBlockHost is the host of the first block
	FirstBlockHost = flag.String("firstBlockHost", defaultFirstBlockHost, "FirstBlockHost")

	// WalletAddress is a wallet address for forging
	WalletAddress = flag.String("walletAddress", "", "walletAddress for forging ")

	// LogSQL show if we should display sql queries in logs
	LogSQL = flag.Bool("logSQL", false, "set DBConn.LogMode")

	// LogStackTrace show if we should display stack trace in logs
	LogStackTrace = flag.Bool("logStackTrace", false, "log stack trace")

	// TestRollBack starts special set of daemons
	TestRollBack = flag.Bool("testRollBack", false, "starts special set of daemons")

	// StartBlockID is the start block
	StartBlockID = flag.Int64("startBlockId", 0, "Start block for blockCollection daemon")

	// EndBlockID is the end block
	EndBlockID = flag.Int64("endBlockId", 0, "End block for blockCollection daemon")

	// RollbackToBlockID is the target block for rollback
	RollbackToBlockID = flag.Int64("rollbackToBlockId", 0, "Rollback to block_id")

	// TLS is a directory for .well-known and keys. It is required for https
	TLS = flag.String("tls", "", "Enable https. Ddirectory for .well-known and keys")

	// UpdateInterval is interval in seconds for checking updates
	UpdateInterval = flag.Int64("updateInterval", defaultUpdateInterval, "Interval in seconds for checking updates, default 3600 seconds (1 hour)")

	// CheckReadAccess access check for reading, is used only for VDE
	CheckReadAccess = flag.Bool("checkReadAccess", true, "Check access for reading, only used for VDE")

	IsVDEMasterMode = flag.Bool("VDEMaster", false, "Setup mode as VDEMaster, allow manage child VDE systems")

	IsVDEMode = flag.Bool("VDEMode", false, "Setup mode as VDE")
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

	if Config.PrivateDir == "" {
		Config.PrivateDir = Config.WorkDir
	}

	if *FirstBlockPath == "" {
		*FirstBlockPath = filepath.Join(Config.PrivateDir, consts.FirstBlockFilename)
	}

	if *keyID != 0 {
		Config.KeyID = *keyID
	}
}
