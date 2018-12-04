// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

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

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package log

import (
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"

	"github.com/sirupsen/logrus"
)

// ContextHook storing nothing but behavior
type ContextHook struct{}

// Levels returns all log levels
func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire the log entry
func (hook ContextHook) Fire(entry *logrus.Entry) error {
	var pc []uintptr
	if conf.Config.Log.LogLevel == "DEBUG" {
		pc = make([]uintptr, 15, 15)
	} else {
		pc = make([]uintptr, 4, 4)
	}
	cnt := runtime.Callers(6, pc)

	count := 0
	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !strings.Contains(name, "github.com/sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			if count == 0 {
				entry.Data["file"] = path.Base(file)
				entry.Data["func"] = path.Base(name)
				entry.Data["line"] = line
				entry.Data["time"] = time.Now().Format(time.RFC3339)
				if conf.Config.Log.LogLevel != "DEBUG" {
					break
				}
			}
			if count >= 1 {
				if count == 1 {
					entry.Data["from"] = []string{}
				}
				entry.Data["from"] = append(entry.Data["from"].([]string), path.Base(name))
			}
			count += 1
		}
	}
	return nil
}
