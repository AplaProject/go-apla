package vdemanager

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

const (
	inidDBCommand  = "initDatabase"
	genKeysCommand = "generateKeys"
	startCommand   = "start"
)

// ChildVDEConfig struct to manage child entry
type ChildVDEConfig struct {
	Executable     string
	Name           string
	Directory      string
	DBUser         string
	DBPassword     string
	ConfigFileName string
	LogTo          string
	LogLevel       string
	HTTPPort       int
}

func (c ChildVDEConfig) configCommand() *exec.Cmd {

	args := []string{
		"config",
		fmt.Sprintf("--path=%s", c.configPath()),
		fmt.Sprintf("--dbUser=%s", c.DBUser),
		fmt.Sprintf("--dbPassword=%s", c.DBPassword),
		fmt.Sprintf("--dbName=%s", c.Name),
		fmt.Sprintf("--httpPort=%d", c.HTTPPort),
		fmt.Sprintf("--dataDir=%s", c.Directory),
		fmt.Sprintf("--keysDir=%s", c.Directory),
		fmt.Sprintf("--logTo=%s", c.LogTo),
		fmt.Sprintf("--logLevel=%s", c.LogLevel),
		"--runMode=VDE",
	}

	return exec.Command(c.Executable, args...)
}

func (c ChildVDEConfig) initDBCommand() *exec.Cmd {
	return c.getCommand(inidDBCommand)
}

func (c ChildVDEConfig) generateKeysCommand() *exec.Cmd {
	return c.getCommand(genKeysCommand)
}

func (c ChildVDEConfig) startCommand() *exec.Cmd {
	return c.getCommand(startCommand)
}

func (c ChildVDEConfig) configPath() string {
	return filepath.Join(c.Directory, c.ConfigFileName)
}

func (c ChildVDEConfig) getCommand(commandName string) *exec.Cmd {
	args := []string{
		commandName,
		fmt.Sprintf("--config=%s", c.configPath()),
	}

	return exec.Command(c.Executable, args...)
}
