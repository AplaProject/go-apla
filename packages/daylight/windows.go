// +build windows
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

package daylight

import (
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

// KillPid kills the process with the specified pid
func KillPid(pid string) error {
	if model.DBConn != nil {
		sd := &model.StopDaemon{StopTime: time.Now().Unix()}
		err := sd.Create()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Error creating StopDaemon")
			return err
		}
	}
	rez, err := exec.Command("tasklist", "/fi", "PID eq "+pid).Output()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CommandExecutionError, "err": err, "cmd": "tasklist /fi PID eq" + pid}).Error("Error executing command")
		return err
	}
	if string(rez) == "" {
		return fmt.Errorf("null")
	}
	log.WithFields(log.Fields{"cmd": "tasklist /fi PID eq " + pid}).Debug("command execution result")
	if ok, _ := regexp.MatchString(`(?i)PID`, string(rez)); !ok {
		return fmt.Errorf("null")
	}
	return nil
}
