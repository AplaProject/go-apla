// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron"
)

var zeroTime time.Time

// Handler represents interface of task handler
type Handler interface {
	Run(*Task)
}

// Task represents task
type Task struct {
	ID       string
	CronSpec string

	Handler Handler

	schedule cron.Schedule
}

// String returns description of task
func (t *Task) String() string {
	return fmt.Sprintf("%s %s", t.ID, t.CronSpec)
}

// ParseCron parsed cron format
func (t *Task) ParseCron() error {
	if len(t.CronSpec) == 0 {
		return nil
	}

	var err error
	t.schedule, err = Parse(t.CronSpec)
	return err
}

// Next returns time for next task
func (t *Task) Next(tm time.Time) time.Time {
	if len(t.CronSpec) == 0 {
		return zeroTime
	}
	return t.schedule.Next(tm)
}

// Run executes task
func (t *Task) Run() {
	t.Handler.Run(t)
}
