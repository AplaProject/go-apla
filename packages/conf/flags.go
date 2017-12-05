package conf

import (
	"flag"
)

var (
	// // // Config.toml // // //

	// FlagDbName database name
	FlagDbName = flag.String("dbName", "", "database name (default apla)")
	// FlagDbHost database host name
	FlagDbHost = flag.String("dbHost", "", "database host (default localhost)")
	// FlagDbPort database port
	FlagDbPort = flag.Int("dbPort", 0, "database port (default 5432)")
	// FlagDbUser database user name
	FlagDbUser = flag.String("dbUser", "", "database user")
	// FlagDbPassword database password
	FlagDbPassword = flag.String("dbPassword", "", "database password, use PG_PASSWORD env to be more secure")

	// FlagTCPHost daemon's host
	FlagTCPHost = flag.String("tcpHost", "", "tcpHost (e.g. 127.0.0.1)")
	// FlagTCPPort daemins's port bind to
	FlagTCPPort = flag.Int("tcpPort", 0, "tcpPort 7080 by default")

	// FlagHTTPHost http api endpoint host
	FlagHTTPHost = flag.String("httpHost", "", "http api bound to that host, use 0.0.0.0 to bind all addresses")
	// FlagHTTPPort http api endpoint port
	FlagHTTPPort = flag.Int("httpPort", 0, "http api port (7079)")

	// FlagWorkDir application working directory
	FlagWorkDir = flag.String("workDir", "", "work directory")

	// FlagPrivateDir - dirctory to store PrivateKey and NodePrivateKey
	FlagPrivateDir = flag.String("privateDir", "", "where privatekeys are stored")

	// FlagLogLevel set log level
	FlagLogLevel = flag.String("logLevel", "", "LogLevel")

	// FlagLogFile log file
	FlagLogFile = flag.String("logFile", "", "log file")

	// FlagKeyID wallet id
	FlagKeyID = flag.Int64("keyID", 0, "wallet id")

	// // //

	// runtime paramters

	// ConfigPath - path to config file
	ConfigPath = flag.String("configPath", "", "full path to config file (toml format)")

	// FirstBlockPath is a file (1block) where first block file will be stored
	FirstBlockPath = flag.String("firstBlockPath", "", "pathname of '1block' file")

	// InitConfig rewrite config using comandline args
	InitConfig = flag.Bool("initConfig", false, "reset config")

	// InitDatabase recreate database
	InitDatabase = flag.Bool("initDatabase", false, "recreate database")

	// GenerateFirstBlock force regeneration of first block
	GenerateFirstBlock = flag.Bool("generateFirstBlock", false, "force init first block")
)

// ParseFlags from command line
func ParseFlags() {
	flag.Parse()
}

//.
