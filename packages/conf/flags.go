package conf

import (
	"flag"
)

var (
	// FlagReinstall rewrite config using comandline args
	FlagReinstall = flag.Bool("reinstall", false, "reset config, init database")

	// FlagConfigPath - path to config file
	FlagConfigPath = flag.String("configPath", "", "full path to config file in toml format'")

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

	// FlagFirstBlockPath is a file (1block) where first block file will be stored
	FlagFirstBlockPath = flag.String("firstBlockPath", "", "pathname of '1block' file")

	// FlagPrivateDir - dirctory to store PrivateKey and NodePrivateKey
	FlagPrivateDir = flag.String("privateDir", "", "where privatekeys are stored")

	// FlagKeyID is the wallet identifier
	FlagKeyID = flag.Int64("keyID", 0, "keyID")

	// FlagLogLevel set log level
	FlagLogLevel = flag.String("logLevel", "", "LogLevel")

	// FlagLogFile log file
	FlagLogFile = flag.String("logFile", "", "log file")
)

// ParseFlags from command line
func ParseFlags() {
	flag.Parse()
}

//.
