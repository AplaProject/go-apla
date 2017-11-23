package spawn

import (
	"os"
	"path/filepath"
	"strings"
)

/*
GetThrustDirectory returns the Directory where the unzipped thrust contents are.
Differs between builds based on OS
*/
func GetThrustDirectory() string {
	return filepath.Join(base, "vendor", "linux", "x64", thrustVersion)
}

/*
GetDownloadDirectory gets the download or extract directory for Thrust
*/
func GetDownloadPath() string {
	return strings.Replace(filepath.Join(base, "$V"), "$V", thrustVersion, 1)
}

/*
GetExecutablePath returns the path to the Thrust Executable
Differs between builds based on OS
*/
func GetExecutablePath() string {
	return GetThrustDirectory() + "/thrust_shell"
}

/*
GetDownloadURL returns the interpolatable version of the Thrust download url
Differs between builds based on OS
*/
func GetDownloadURL() string {
	return "https://github.com/breach/thrust/releases/download/v$V/thrust-v$V-linux-x64.zip"
}

/*
Bootstrap executes the default bootstrapping plan for this system and returns an error if failed
*/
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
	path, err := downloadFromUrl(GetDownloadURL(), base+"/$V", thrustVersion)
	if err != nil {
		return err
	}

	return UnzipExecutable(path)
}

func UnzipExecutable(path string) error {
	return unzip(path, GetThrustDirectory())
}

func PathNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
