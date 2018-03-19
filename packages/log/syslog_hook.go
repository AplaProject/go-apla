// +build !windows,!nacl,!plan9

package log

import (
	"fmt"
	"log/syslog"
	"os"

	b_syslog "github.com/blackjack/syslog"
	"github.com/sirupsen/logrus"
)

// SyslogHook to send logs via syslog.
type SyslogHook struct {
	Writer        *b_syslog.Writer
	SyslogNetwork string
	SyslogRaddr   string
}

func NewSyslogHook(appName string, priority syslog.Priority) (*SyslogHook, error) {
	b_syslog.Openlog(appName, b_syslog.LOG_PID, b_syslog.Priority(priority))
	return &SyslogHook{nil, "", "localhost"}, nil
}

func (hook *SyslogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}

	switch entry.Level {
	case logrus.PanicLevel:
		{
			b_syslog.Crit(line)
			return nil
		}
	case logrus.FatalLevel:
		{
			b_syslog.Crit(line)
			return nil
		}
	case logrus.ErrorLevel:
		{
			b_syslog.Err(line)
			return nil
		}
	case logrus.WarnLevel:
		{
			b_syslog.Warning(line)
			return nil
		}
	case logrus.InfoLevel:
		{
			b_syslog.Info(line)
			return nil
		}
	case logrus.DebugLevel:
		{
			b_syslog.Debug(line)
			return nil
		}
	default:
		return nil
	}
}

func (hook *SyslogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
