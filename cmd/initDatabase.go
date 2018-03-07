package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/model"
)

// initDatabaseCmd represents the initDatabase command
var initDatabaseCmd = &cobra.Command{
	Use:   "initDatabase",
	Short: "Initializing database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(dbUser)
		dbPassword := os.Getenv("DB_PASSWORD")
		if err := model.InitDB(
			conf.DBConfig{
				Name: dbName,
				HostPort: conf.HostPort{
					Host: dbHost,
					Port: dbPort,
				},
				User:     dbUser,
				Password: dbPassword,
			},
		); err != nil {
			log.WithError(err).Fatal("init db")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initDatabaseCmd)

	initDatabaseCmd.Flags().IntVar(&dbPort, "dbPort", 5432, "genesis database port to rollback")
	initDatabaseCmd.Flags().StringVar(&dbHost, "dbHost", "localhost", "genesis database host to rollback")
	initDatabaseCmd.Flags().StringVar(&dbName, "dbName", "genesis", "genesis database name")
	initDatabaseCmd.Flags().StringVar(&dbUser, "dbUser", "postgres", "genesis database username")
}
