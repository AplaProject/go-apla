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
	"github.com/AplaProject/go-apla/packages/consts"

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
