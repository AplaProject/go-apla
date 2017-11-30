package spawn

/*
Package spawn implements methods and interfaces used in downloading and spawning the underlying thrust core binary.
*/
import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	. "github.com/miketheprogrammer/go-thrust/lib/common"
	"github.com/miketheprogrammer/go-thrust/lib/connection"
)

const (
	thrustVersion = "0.7.6"
)

var (
	// ApplicationName only functionally applies to OSX builds, otherwise it is only cosmetic
	ApplicationName = "Go Thrust"
	// base directory for storing the executable
	base = ""
)

/*
SetBaseDirectory sets the base directory used in the other helper methods
*/
func SetBaseDirectory(dir string) error {
	if len(dir) == 0 {
		usr, err := user.Current()
		if err != nil {
			fmt.Println(err)
		}
		dir = usr.HomeDir
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Println("Could not calculate absolute path", err)
		return err
	}
	base = dir

	return nil
}

/*
The SpawnThrustCore method is a bootstrap and run method.
It will try to detect an installation of thrust, if it cannot find it
it will download the version of Thrust detailed in the "common" package.
Once downloaded, it will launch a process.
Go-Thrust and all *-Thrust packages communicate with Thrust Core via Stdin/Stdout.
using -log=debug as a command switch will give you the most information about what is going on. -log=info will give you notices that stuff is happening.
Any log level higher than that will output nothing.
*/
func Run() {
	if Log == nil {
		InitLogger("debug")
	}
	if base == "" {
		SetBaseDirectory("") // Default to usr.homedir.
	}

	thrustExecPath := GetExecutablePath()
	if len(thrustExecPath) > 0 {

		if provisioner == nil {
			SetProvisioner(NewThrustProvisioner())
		}
		if err := provisioner.Provision(); err != nil {
			panic(err)
		}

		thrustExecPath = GetExecutablePath()

		Log.Print("Attempting to start Thrust Core")
		Log.Print("CMD:", thrustExecPath)
		cmd := exec.Command(thrustExecPath)
		cmdIn, e1 := cmd.StdinPipe()
		cmdOut, e2 := cmd.StdoutPipe()

		if e1 != nil {
			fmt.Println(e1)
			os.Exit(2) // need to improve exit codes
		}

		if e2 != nil {
			fmt.Println(e2)
			os.Exit(2)
		}

		if LogLevel != "none" {
			cmd.Stderr = os.Stdout
		}

		if err := cmd.Start(); err != nil {
			Log.Panic("Thrust Core not started.")
		}

		Log.Print("Thrust Core started.")

		// Setup our Connection.
		connection.Stdout = cmdOut
		connection.Stdin = cmdIn
		connection.ExecCommand = cmd
		connection.InitializeThreads()
		return
	} else {
		fmt.Println("===============WARNING================")
		fmt.Println("Current operating system not supported", runtime.GOOS)
		fmt.Println("===============END====================")
	}
	return
}
