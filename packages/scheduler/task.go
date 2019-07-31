// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

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
