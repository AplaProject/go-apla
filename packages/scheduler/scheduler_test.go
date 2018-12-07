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

package scheduler

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	cases := map[string]string{
		"60 * * * *":              "End of range (60) above maximum (59): 60",
		"0-59 0-23 1-31 1-12 0-6": "",
		"*/2 */2 */2 */2 */2":     "",
		"* * * * *":               "",
	}

	for cronSpec, expectedErr := range cases {
		_, err := Parse(cronSpec)
		if err != nil {
			if errStr := err.Error(); errStr != expectedErr {
				t.Errorf("cron: %s, expected: %s, got: %s\n", cronSpec, expectedErr, errStr)
			}

			continue
		}

		if expectedErr != "" {
			t.Errorf("cron: %s, error: %s\n", cronSpec, err)
		}
	}
}

type mockHandler struct {
	count int
}

func (mh *mockHandler) Run(t *Task) {
	mh.count++
}

// This test required timeout 60s
// go test -timeout 60s
func TestTask(t *testing.T) {
	var taskID = "task1"
	sch := NewScheduler()

	task := &Task{ID: taskID}

	nextTime := task.Next(time.Now())
	if nextTime != zeroTime {
		t.Error("error")
	}

	task = &Task{CronSpec: "60 * * * *"}
	err := sch.AddTask(task)
	if errStr := err.Error(); errStr != "End of range (60) above maximum (59): 60" {
		t.Error(err)
	}
	err = sch.UpdateTask(task)
	if errStr := err.Error(); errStr != "End of range (60) above maximum (59): 60" {
		t.Error(err)
	}

	err = sch.UpdateTask(&Task{ID: "task2"})
	if err != nil {
		t.Error(err)
	}

	handler := &mockHandler{}
	task = &Task{ID: taskID, CronSpec: "* * * * *", Handler: handler}
	sch.UpdateTask(task)

	now := time.Now()
	time.Sleep(task.Next(now).Sub(now) + time.Second)

	if handler.count == 0 {
		t.Error("task not running")
	}
}
