// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package obsmanager

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

// ChildOBSConfig struct to manage child entry
type ChildOBSConfig struct {
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

func (c ChildOBSConfig) configCommand() *exec.Cmd {

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
		"--obsMode=OBS",
	}

	return exec.Command(c.Executable, args...)
}

func (c ChildOBSConfig) initDBCommand() *exec.Cmd {
	return c.getCommand(inidDBCommand)
}

func (c ChildOBSConfig) generateKeysCommand() *exec.Cmd {
	return c.getCommand(genKeysCommand)
}

func (c ChildOBSConfig) startCommand() *exec.Cmd {
	return c.getCommand(startCommand)
}

func (c ChildOBSConfig) configPath() string {
	return filepath.Join(c.Directory, c.ConfigFileName)
}

func (c ChildOBSConfig) getCommand(commandName string) *exec.Cmd {
	args := []string{
		commandName,
		fmt.Sprintf("--config=%s", c.configPath()),
	}

	return exec.Command(c.Executable, args...)
}
