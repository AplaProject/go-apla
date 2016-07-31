package spawn

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-thrust/lib/common"
)

/*
GetThrustDirectory returns the Directory where the unzipped thrust contents are.
Differs between builds based on OS
*/
func GetThrustDirectory() string {
	return base
}

/*
GetExecutablePath returns the path to the Thrust Executable
Differs between builds based on OS
*/
func GetExecutablePath() string {
	return filepath.Join(GetThrustDirectory(), "thrust_shell.exe")
}

/*
GetDownloadUrl returns the interpolatable version of the Thrust download url
Differs between builds based on OS
*/
func GetDownloadUrl() string {
	//https://github.com/breach/thrust/releases/download/v0.7.5/thrust-v0.7.5-win32-ia32.zip
	return "https://github.com/breach/thrust/releases/download/v$V/thrust-v$V-win32-ia32.zip"
}

func Bootstrap() error {
	if executableNotExist() == true {
		if err := prepareExecutable(); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func executableNotExist() bool {
	_, err := os.Stat(GetExecutablePath())
	return os.IsNotExist(err)
}

func prepareExecutable() error {
	common.Log.Print(base + "\\$V")
	_, err := downloadFromUrl(GetDownloadUrl(), base+"\\$V", thrustVersion)
	if err != nil {
		return err
	}
	if err = unzip(strings.Replace(base+"\\$V", "$V", thrustVersion, 1), GetThrustDirectory()); err != nil {
		return err
	}
	return nil
}
