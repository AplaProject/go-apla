package utils

import "time"

type Clock interface {
	Now() time.Time
}

type ClockWrapper struct {
}

func (cw *ClockWrapper) Now() time.Time { return time.Now() }
