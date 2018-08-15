// MIT License
//
// Copyright (c) 2016 GenesisCommunity
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
	"github.com/GenesisCommunity/go-genesis/packages/consts"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

var scheduler *Scheduler

func init() {
	scheduler = NewScheduler()
}

// Scheduler represents wrapper over the cron library
type Scheduler struct {
	cron *cron.Cron
}

// AddTask adds task to cron
func (s *Scheduler) AddTask(t *Task) error {
	err := t.ParseCron()
	if err != nil {
		return err
	}

	s.cron.Schedule(t, t)
	log.WithFields(log.Fields{"task": t.String()}).Info("task added")

	return nil
}

// UpdateTask updates task
func (s *Scheduler) UpdateTask(t *Task) error {
	err := t.ParseCron()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParseError, "error": err}).Error("parse cron format")
		return err
	}

	s.cron.Stop()
	defer s.cron.Start()

	entries := s.cron.Entries()
	for _, entry := range entries {
		task := entry.Schedule.(*Task)
		if task.ID == t.ID {
			*task = *t
			log.WithFields(log.Fields{"task": t.String()}).Info("task updated")
			return nil
		}

		continue
	}

	s.cron.Schedule(t, t)
	log.WithFields(log.Fields{"task": t.String()}).Info("task added")

	return nil
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	s := &Scheduler{cron: cron.New()}
	s.cron.Start()
	return s
}

// AddTask adds task to global scheduler
func AddTask(t *Task) error {
	return scheduler.AddTask(t)
}

// UpdateTask updates task in global scheduler
func UpdateTask(t *Task) error {
	return scheduler.UpdateTask(t)
}

// Parse parses cron format
func Parse(cronSpec string) (cron.Schedule, error) {
	sch, err := cron.ParseStandard(cronSpec)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParseError, "error": err}).Error("parse cron format")
		return nil, err
	}
	return sch, nil
}
