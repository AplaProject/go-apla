package conf

import (
	"flag"
)

var (
	// run mode flags:

	// FlagReinstall rewrite config using comandline args
	FlagReinstall = flag.Bool("reinstall", false, "reset config, init database")

	// config flags:

	// FlagConfigPath - path to config file
	FlagConfigPath = flag.String("configPath", "", "full path to config file in toml format'")

	FlagDbName     = flag.String("dbName", "", "database name (default apla)")
	FlagDbHost     = flag.String("dbHost", "", "database host (default localhost)")
	FlagDbPort     = flag.Int("dbPort", 0, "database port (default 5432)")
	FlagDbUser     = flag.String("dbUser", "", "database user")
	FlagDbPassword = flag.String("dbPassword", "", "database password") // insecure! use env.PG_PASSWORD instead

	// FlagTCPHost daemon's host
	FlagTCPHost = flag.String("tcpHost", "", "tcpHost (e.g. 127.0.0.1)")
	// FlagTCPPort daemins's port bind to
	FlagTCPPort = flag.Int("tcpPort", 0, "tcpPort 7080 by default")

	// FlagHTTPHost http api endpoint host
	FlagHTTPHost = flag.String("httpHost", "", "http api bound to that host, use 0.0.0.0 to bind all addresses")
	// FlagHTTPPort http api endpoint port
	FlagHTTPPort = flag.Int("httpPort", 0, "http api port (7079)")

	FlagWorkDir = flag.String("workDir", "", "work directory")
	FlagDir     = flag.String("dDir", "", "work directory (deprecated")

	// FlagFirstBlockPath is a file (1block) where first block file will be stored
	FlagFirstBlockPath = flag.String("firstBlockPath", "", "pathname of '1block' file")

	// FlagPrivateDir - dirctory to store PrivateKey and NodePrivateKey
	FlagPrivateDir = flag.String("privateDir", "", "where privatekeys are stored")

	// // FirstBlockPublicKey is the private key
	// FirstBlockPublicKey = flag.String("firstBlockPublicKey", "", "FirstBlockPublicKey")
	// // FirstBlockNodePublicKey is the node private key
	// FirstBlockNodePublicKey = flag.String("firstBlockNodePublicKey", "", "FirstBlockNodePublicKey")
	// // FirstBlockHost is the host of the first block
	// FirstBlockHost = flag.String("firstBlockHost", "", "FirstBlockHost")
	// // WalletAddress is a wallet address for forging
	// WalletAddress = flag.String("walletAddress", "", "walletAddress for forging ")

	// FlagLogLevel set log level
	FlagLogLevel = flag.String("logLevel", "", "apla LogLevel")

	// // GenerateFirstBlock show if the first block must be generated
	// GenerateFirstBlock = flag.Int64("generateFirstBlock", 0, "generateFirstBlock")

	// // LogSQL show if we should display sql queries in logs
	// LogSQL = flag.Int64("logSQL", 0, "log sql")
	// // LogStackTrace show if we should display stack trace in logs
	// LogStackTrace = flag.Int64("logStackTrace", 0, "log stack trace")

	// // TestRollBack equals 1 for testing rollback
	// TestRollBack = flag.Int64("testRollBack", 0, "testRollBack")

	// // StartBlockID is the start block
	// StartBlockID = flag.Int64("startBlockId", 0, "Start block for blockCollection daemon")
	// // EndBlockID is the end block
	// EndBlockID = flag.Int64("endBlockId", 0, "End block for blockCollection daemon")
	// // RollbackToBlockID is the target block for rollback
	// RollbackToBlockID = flag.Int64("rollbackToBlockId", 0, "Rollback to block_id")
	// // TLS is a directory for .well-known and keys. It is required for https
	// TLS = flag.String("tls", "", "Support https. Specify directory for .well-known")
	// // DevTools switches on dev tools in thrust shell

	// // DltWalletID is the wallet identifier
	// KeyID = flag.Int64("keyID", 0, "keyID")

)

// ParseFlags from command line
func ParseFlags() {
	flag.Parse()
}

//.
