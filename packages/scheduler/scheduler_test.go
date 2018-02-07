//MIT License
//
//Copyright (c) 2016-2018 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

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
