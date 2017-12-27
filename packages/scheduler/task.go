package scheduler

import (
	"time"
)

type Task interface {
	String() string

	Equal(Task) bool
	ParseCron() error
	Update(Task)

	Next(time.Time) time.Time
	Run()
}
