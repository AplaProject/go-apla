package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"path/filepath"

	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Initial config generation",
	Run: func(cmd *cobra.Command, args []string) {
		// Error omitted because we have default flag value
		configPath, _ := cmd.Flags().GetString("path")

		err := conf.FillRuntimePaths()
		if err != nil {
			log.WithError(err).Fatal("Filling config")
		}

		if configPath == "" {
			configPath = filepath.Join(conf.Config.DataDir, consts.DefaultConfigFile)
		}

		err = conf.SaveConfig(configPath)
		if err != nil {
			log.WithError(err).Fatal("Saving config")
		}

		log.Infof("Config is saved to %s", configPath)
	},
}

func init() {
	// Command flags
	configCmd.Flags().String("path", "", "Generate config to (default dataDir/config.toml)")

	// TCP Server
	configCmd.Flags().StringVar(&conf.Config.TCPServer.Host, "tcpHost", "127.0.0.1", "Node TCP host")
	configCmd.Flags().IntVar(&conf.Config.TCPServer.Port, "tcpPort", 7078, "Node TCP port")
	viper.BindPFlag("TCPServer.Host", configCmd.Flags().Lookup("tcpHost"))
	viper.BindPFlag("TCPServer.Port", configCmd.Flags().Lookup("tcpPort"))

	// HTTP Server
	configCmd.Flags().StringVar(&conf.Config.HTTP.Host, "httpHost", "127.0.0.1", "Node HTTP host")
	configCmd.Flags().IntVar(&conf.Config.HTTP.Port, "httpPort", 7079, "Node HTTP port")
	viper.BindPFlag("HTTP.Host", configCmd.Flags().Lookup("httpHost"))
	viper.BindPFlag("HTTP.Port", configCmd.Flags().Lookup("httpPort"))

	// DB
	configCmd.Flags().StringVar(&conf.Config.DB.Host, "dbHost", "127.0.0.1", "DB host")
	configCmd.Flags().IntVar(&conf.Config.DB.Port, "dbPort", 5432, "DB port")
	configCmd.Flags().StringVar(&conf.Config.DB.Name, "dbName", "genesis", "DB name")
	configCmd.Flags().StringVar(&conf.Config.DB.User, "dbUser", "postgres", "DB username")
	configCmd.Flags().StringVar(&conf.Config.DB.Password, "dbPassword", "genesis", "DB password")
	viper.BindPFlag("DB.Name", configCmd.Flags().Lookup("dbName"))
	viper.BindPFlag("DB.Host", configCmd.Flags().Lookup("dbHost"))
	viper.BindPFlag("DB.Port", configCmd.Flags().Lookup("dbPort"))
	viper.BindPFlag("DB.User", configCmd.Flags().Lookup("dbUser"))
	viper.BindPFlag("DB.Password", configCmd.Flags().Lookup("dbPassword"))

	// StatsD
	configCmd.Flags().StringVar(&conf.Config.StatsD.Host, "statsdHost", "127.0.0.1", "StatsD host")
	configCmd.Flags().IntVar(&conf.Config.StatsD.Port, "statsdPort", 8125, "StatsD port")
	configCmd.Flags().StringVar(&conf.Config.StatsD.Name, "statsdName", "genesis", "StatsD name")
	viper.BindPFlag("StatsD.Host", configCmd.Flags().Lookup("statsdHost"))
	viper.BindPFlag("StatsD.Port", configCmd.Flags().Lookup("statsdPort"))
	viper.BindPFlag("StatsD.Name", configCmd.Flags().Lookup("statsdName"))

	// Centrifugo
	configCmd.Flags().StringVar(&conf.Config.Centrifugo.Secret, "centSecret", "127.0.0.1", "Centrifugo secret")
	configCmd.Flags().StringVar(&conf.Config.Centrifugo.URL, "centUrl", "127.0.0.1", "Centrifugo URL")
	viper.BindPFlag("Centrifugo.Secret", configCmd.Flags().Lookup("centSecret"))
	viper.BindPFlag("Centrifugo.URL", configCmd.Flags().Lookup("centUrl"))

	// Log
	configCmd.Flags().StringVar(&conf.Config.LogConfig.LogTo, "logTo", "stdout", "Send logs to stdout|(filename)|syslog")
	configCmd.Flags().StringVar(&conf.Config.LogConfig.LogLevel, "verbosity", "ERROR", "Log verbosity (DEBUG | INFO | WARN | ERROR)")
	configCmd.Flags().StringVar(&conf.Config.LogConfig.LogFormat, "logFormat", "text", "log format, could be text|json")
	configCmd.Flags().StringVar(&conf.Config.LogConfig.Syslog.Facility, "syslogFacility", "LOG_KERN", "syslog facility")
	configCmd.Flags().StringVar(&conf.Config.LogConfig.Syslog.Tag, "syslogTag", "go-genesis", "syslog program tag")
	viper.BindPFlag("Log.LogTo", configCmd.Flags().Lookup("logTo"))
	viper.BindPFlag("Log.Verbosity", configCmd.Flags().Lookup("verbosity"))
	viper.BindPFlag("Log.LogFormat", configCmd.Flags().Lookup("logFormat"))
	viper.BindPFlag("Log.Syslog.Facility", configCmd.Flags().Lookup("syslogFacility"))
	viper.BindPFlag("Log.Syslog.Tag", configCmd.Flags().Lookup("syslogTag"))

	// TokenMovement
	configCmd.Flags().StringVar(&conf.Config.TokenMovement.Host, "tmovHost", "", "Token movement host")
	configCmd.Flags().IntVar(&conf.Config.TokenMovement.Port, "tmovPort", 0, "Token movement port")
	configCmd.Flags().StringVar(&conf.Config.TokenMovement.Username, "tmovUser", "", "Token movement username")
	configCmd.Flags().StringVar(&conf.Config.TokenMovement.Password, "tmovPw", "", "Token movement password")
	configCmd.Flags().StringVar(&conf.Config.TokenMovement.To, "tmovTo", "", "Token movement to field")
	configCmd.Flags().StringVar(&conf.Config.TokenMovement.From, "tmovFrom", "", "Token movement from field")
	configCmd.Flags().StringVar(&conf.Config.TokenMovement.Subject, "tmovSubj", "", "Token movement subject")
	viper.BindPFlag("TokenMovement.Host", configCmd.Flags().Lookup("tmovHost"))
	viper.BindPFlag("TokenMovement.Port", configCmd.Flags().Lookup("tmovPort"))
	viper.BindPFlag("TokenMovement.Username", configCmd.Flags().Lookup("tmovUser"))
	viper.BindPFlag("TokenMovement.Password", configCmd.Flags().Lookup("tmovPw"))
	viper.BindPFlag("TokenMovement.To", configCmd.Flags().Lookup("tmovTo"))
	viper.BindPFlag("TokenMovement.From", configCmd.Flags().Lookup("tmovFrom"))
	viper.BindPFlag("TokenMovement.Subject", configCmd.Flags().Lookup("tmovSubj"))

	// Etc
	configCmd.Flags().StringVar(&conf.Config.PidFilePath, "pid", "",
		fmt.Sprintf("Genesis pid file name (default dataDir/%s)", consts.DefaultPidFilename),
	)
	configCmd.Flags().StringVar(&conf.Config.LockFilePath, "lock", "",
		fmt.Sprintf("Genesis lock file name (default dataDir/%s)", consts.DefaultLockFilename),
	)
	configCmd.Flags().StringVar(&conf.Config.KeysDir, "keysDir", "", "Keys directory (default dataDir)")
	configCmd.Flags().StringVar(&conf.Config.DataDir, "dataDir", "", "Data directory (default cwd/genesis-data)")
	configCmd.Flags().StringVar(&conf.Config.FirstBlockPath, "firstBlock", "", "First block path (default dataDir/1block)")
	configCmd.Flags().BoolVar(&conf.Config.TLS, "tls", false, "Enable https")
	configCmd.Flags().StringVar(&conf.Config.TLSCert, "tls-cert", "", "Filepath to the fullchain of certificates")
	configCmd.Flags().StringVar(&conf.Config.TLSKey, "tls-key", "", "Filepath to the private key")
	configCmd.Flags().Int64Var(&conf.Config.MaxPageGenerationTime, "mpgt", 1000, "Max page generation time in ms")
	configCmd.Flags().StringArrayVar(&conf.Config.NodesAddr, "nodesAddr", []string{}, "List of addresses for downloading blockchain")
	viper.BindPFlag("PidFilePath", configCmd.Flags().Lookup("pid"))
	viper.BindPFlag("LockFilePath", configCmd.Flags().Lookup("lock"))
	viper.BindPFlag("KeysDir", configCmd.Flags().Lookup("keysDir"))
	viper.BindPFlag("DataDir", configCmd.Flags().Lookup("dataDir"))
	viper.BindPFlag("FirstBlockPath", configCmd.Flags().Lookup("firstBlock"))
	viper.BindPFlag("TLS", configCmd.Flags().Lookup("tls"))
	viper.BindPFlag("TLSCert", configCmd.Flags().Lookup("tls-cert"))
	viper.BindPFlag("TLSKey", configCmd.Flags().Lookup("tls-key"))
	viper.BindPFlag("MaxPageGenerationTime", configCmd.Flags().Lookup("mpgt"))
}
