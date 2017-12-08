package conf

import (
	"flag"
)

var (
	// // // Config.toml // // //

	// FlagDbName database name
	FlagDbName = flag.String("dbName", "", "database name (default apla) or environment PGDATABASE")
	// FlagDbHost database host name
	FlagDbHost = flag.String("dbHost", "", "database host (default localhost) or environment PGHOST")
	// FlagDbPort database port
	FlagDbPort = flag.Int("dbPort", 0, "database port (default 5432) or environment PGPORT")
	// FlagDbUser database user name
	FlagDbUser = flag.String("dbUser", "", "database user or environment PGUSER")
	// FlagDbPassword database password
	FlagDbPassword = flag.String("dbPassword", "", "database password, use PGPASSWORD environment to be more secure")

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

	// InitDatabase recreate database
	InitDatabase = flag.Bool("initDatabase", false, "recreate database")

	// GenerateFirstBlock force regeneration of first block
	GenerateFirstBlock = flag.Bool("generateFirstBlock", false, "force init first block")

	// other flags

	// FirstBlockPublicKey is the private key
	FirstBlockPublicKey = flag.String("firstBlockPublicKey", "", "FirstBlockPublicKey")
	// FirstBlockNodePublicKey is the node private key
	FirstBlockNodePublicKey = flag.String("firstBlockNodePublicKey", "", "FirstBlockNodePublicKey")
	// FirstBlockHost is the host of the first block
	FirstBlockHost = flag.String("firstBlockHost", "127.0.0.1", "FirstBlockHost")

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
	TLS = flag.String("tls", "", "Support https. Specify directory for .well-known")
)

// ParseFlags from command line
func ParseFlags() {
	flag.Parse()
}

//.
