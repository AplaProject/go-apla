package utils

import "time"

type ClockWrapper struct {
}

func (cw *ClockWrapper) Now() time.Time { return time.Now() }
