package translators_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CockroachSuite struct {
	suite.Suite
}

type PostgreSQLSuite struct {
	suite.Suite
}

type MySQLSuite struct {
	suite.Suite
}

type MariaDBSuite struct {
	suite.Suite
}

type MsSqlServerSQLSuite struct {
	suite.Suite
}

type SQLiteSuite struct {
	suite.Suite
}

type SchemaSuite struct {
	suite.Suite
}

func TestSpecificSuites(t *testing.T) {
	switch os.Getenv("SODA_DIALECT") {
	case "postgres":
		suite.Run(t, &PostgreSQLSuite{})
	case "cockroach":
		suite.Run(t, &CockroachSuite{})
	case "mysql", "mysql_travis":
		suite.Run(t, &MySQLSuite{})
	case "mariadb":
		suite.Run(t, &MariaDBSuite{})
	case "sqlserver":
		suite.Run(t, &MsSqlServerSQLSuite{})
	case "sqlite":
		suite.Run(t, &SQLiteSuite{})
	}

	suite.Run(t, &SchemaSuite{})
}

func getEnv(key, defaultValue string) string {
	if v, found := os.LookupEnv(key); found {
		return v
	}
	return defaultValue
}
