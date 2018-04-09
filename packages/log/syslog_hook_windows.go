// +build windows
package log

import (
	"github.com/sirupsen/logrus"
)

// SyslogHook to send logs via syslog.
type SyslogHook struct {
	SyslogNetwork string
	SyslogRaddr   string
}

func NewSyslogHook(appName, facility string) (*SyslogHook, error) {
	return &SyslogHook{"", "localhost"}, nil
}

func (hook *SyslogHook) Fire(entry *logrus.Entry) error {
	return nil
}

func (hook *SyslogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
