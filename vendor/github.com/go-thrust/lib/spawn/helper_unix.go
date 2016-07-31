// +build linux freebsd

package spawn

import (
	"os"
	"strings"
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
	return GetThrustDirectory() + "/thrust_shell"
}

/*
GetDownloadUrl returns the interpolatable version of the Thrust download url
Differs between builds based on OS
*/
func GetDownloadUrl() string {
	return "https://github.com/breach/thrust/releases/download/v$V/thrust-v$V-linux-x64.zip"
}

func Bootstrap() error {
	if executableNotExist() == true {
		return prepareExecutable()
	}
	return nil
}

func executableNotExist() bool {
	_, err := os.Stat(GetExecutablePath())
	return os.IsNotExist(err)
}

func prepareExecutable() error {
	_, err := downloadFromUrl(GetDownloadUrl(), base+"/$V", thrustVersion)
	if err != nil {
		return err
	}
	return unzip(strings.Replace(base+"/$V", "$V", thrustVersion, 1), GetThrustDirectory())
}
