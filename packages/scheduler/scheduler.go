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
