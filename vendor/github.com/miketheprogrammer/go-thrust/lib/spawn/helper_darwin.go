package spawn

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

/*
GetThrustDirectory returns the Directory where the unzipped thrust contents are.
Differs between builds based on OS
*/
func GetThrustDirectory() string {
	return base + "/vendor/darwin/x64/v" + thrustVersion
}

/*
GetDownloadPath gets the download or extract directory for Thrust
*/
func GetDownloadPath() string {
	return strings.Replace(filepath.Join(base, "$V"), "$V", thrustVersion, 1)
}

/*
Get the path to the .app directory (only OSX)
*/
func getAppDirectory() string {
	return base + "/vendor/darwin/x64/v" + thrustVersion + "/" + ApplicationName + ".app"
}

/*
GetExecutablePath returns the path to the Thrust Executable
Differs between builds based on OS
*/
func GetExecutablePath() string {
	return GetThrustDirectory() + "/" + ApplicationName + ".app/Contents/MacOS/" + ApplicationName
}

/*
GetDownloadURL returns the interpolatable version of the Thrust download url
Differs between builds based on OS
*/
func GetDownloadURL() string {
	return "https://github.com/breach/thrust/releases/download/v$V/thrust-v$V-darwin-x64.zip"
}

/*
SetThrustApplicationTitle sets the title in the Info.plist. This method only exists on Darwin.
*/
func Bootstrap() error {
	if ExecutableNotExist() == true {
		var err error
		if err = prepareExecutable(); err != nil {
			return err
		}
		if err = prepareInfoPropertiesListTemplate(); err != nil {
			return err
		}

		return writeInfoPropertiesList()
	}

	return nil
}

/*
ExecutableNotExist checks if the executable does not exist
*/
func ExecutableNotExist() bool {
	_, err := os.Stat(GetExecutablePath())
	fmt.Println(err)
	return os.IsNotExist(err)
}

func PathNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

/*
prepareExecutable dowloads, unzips and does alot of other magic to prepare our thrust core build.
*/
func prepareExecutable() error {
	path, err := downloadFromUrl(GetDownloadURL(), base+"/$V", thrustVersion)
	if err != nil {
		return err
	}

	return UnzipExecutable(path)
}

func UnzipExecutable(path string) error {
	if err := unzip(path, GetThrustDirectory()); err != nil {
		return err
	}
	os.Rename(GetThrustDirectory()+"/ThrustShell.app/Contents/MacOS/ThrustShell", GetThrustDirectory()+"/ThrustShell.app/Contents/MacOS/"+ApplicationName)
	os.Rename(GetThrustDirectory()+"/ThrustShell.app", GetThrustDirectory()+"/"+ApplicationName+".app")

	if err := applySymlinks(); err != nil {
		panic(err)
	}

	return nil
}

/*
ApplySymLinks exists because our unzip utility does not respect deferred symlinks. It applies all the neccessary symlinks to make the thrust core exe connect to the thrust core libs.
*/
func applySymlinks() error {
	fmt.Println("Removing bad symlinks")
	var err error
	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current") == false {
		if err = os.Remove(getAppDirectory() + "/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Frameworks") == false {
		if err = os.Remove(getAppDirectory() + "/Contents/Frameworks/ThrustShell Framework.framework/Frameworks"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Libraries") == false {
		if err = os.Remove(getAppDirectory() + "/Contents/Frameworks/ThrustShell Framework.framework/Libraries"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Resources") == false {
		if err = os.Remove(getAppDirectory() + "/Contents/Frameworks/ThrustShell Framework.framework/Resources"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/ThrustShell Framework") == false {
		if err = os.Remove(getAppDirectory() + "/Contents/Frameworks/ThrustShell Framework.framework/ThrustShell Framework"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current/Libraries") == false {
		if err = os.Remove(getAppDirectory() + "/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current/Libraries"); err != nil {
			return err
		}
	}

	fmt.Println("Applying Symlinks")

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current") == true {
		if err = os.Symlink(
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/A",
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Frameworks") == true {
		if err = os.Symlink(
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current/Frameworks",
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Frameworks"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Libraries") == true {
		if err = os.Symlink(
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current/Libraries",
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Libraries"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Resources") == true {
		if err = os.Symlink(
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current/Resources",
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Resources"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/ThrustShell Framework") == true {
		if err = os.Symlink(
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current/ThrustShell Framework",
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/ThrustShell Framework"); err != nil {
			return err
		}
	}

	if PathNotExist(getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current/Libraries") == true {
		if err = os.Symlink(
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/A/Libraries/Libraries",
			getAppDirectory()+"/Contents/Frameworks/ThrustShell Framework.framework/Versions/Current/Libraries"); err != nil {
			return err
		}
	}

	return nil
}

func prepareInfoPropertiesListTemplate() error {
	plistPath := getInfoPropertiesListDirectory() + "/Info.plist"

	// Not an OSX user, but perhaps we should build this on each invocation anyways?
	// Might help prevent stale issues if the plist gets out of date/sync, although
	// I'm not entirely sure how important that is.
	if _, err := os.Stat(plistPath + ".tmpl"); os.IsNotExist(err) {
		plist, err := ioutil.ReadFile(plistPath)

		if err != nil {
			fmt.Println(err)
			return err
		}

		plistTmpl := strings.Replace(string(plist), "ThrustShell", "$$", -1)
		plistTmpl = strings.Replace(plistTmpl, "org.breach.$$", "org.breach.ThrustShell", 1)
		plistTmpl = strings.Replace(plistTmpl, "$$Application", "ThrustShellApplication", 1)
		//func WriteFile(filename string, data []byte, perm os.FileMode) error

		return ioutil.WriteFile(plistPath+".tmpl", []byte(plistTmpl), 0775)
	}

	return nil
}

func writeInfoPropertiesList() error {
	plistPath := getInfoPropertiesListDirectory() + "/Info.plist"
	if err := prepareInfoPropertiesListTemplate(); err == nil {
		plistTmpl, err := ioutil.ReadFile(plistPath + ".tmpl")

		if err != nil {
			fmt.Println(err)
			return err
		}

		plist := strings.Replace(string(plistTmpl), "$$", ApplicationName, -1)

		err = ioutil.WriteFile(plistPath, []byte(plist), 0775)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Could not change title")
		return nil
	}
	return nil
}

func getInfoPropertiesListDirectory() string {
	return GetThrustDirectory() + "/" + ApplicationName + ".app/Contents"
}
