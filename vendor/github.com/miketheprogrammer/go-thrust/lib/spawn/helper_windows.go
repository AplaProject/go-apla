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
	return filepath.Join(base, "vendor", "windows", "ia32", thrustVersion)
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
	return filepath.Join(GetThrustDirectory(), "thrust_shell.exe")
}

/*
GetDownloadURL returns the interpolatable version of the Thrust download url
Differs between builds based on OS
*/
func GetDownloadURL() string {
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
	path, err := downloadFromUrl(GetDownloadURL(), base+"\\$V", thrustVersion)
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
