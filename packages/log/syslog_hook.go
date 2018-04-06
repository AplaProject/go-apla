// +build !windows,!nacl,!plan9

package log

import (
	"encoding/json"
	"fmt"
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

// NewSyslogHook creats SyslogHook
func NewSyslogHook(appName, facility string) (*SyslogHook, error) {
	b_syslog.Openlog(appName, b_syslog.LOG_PID, syslogFacility(facility))
	return &SyslogHook{nil, "", "localhost"}, nil
}

// Fire the log entry
func (hook *SyslogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	jsonMap := map[string]interface{}{}
	if err := json.Unmarshal([]byte(line), &jsonMap); err == nil {
		delete(jsonMap, "time")
		delete(jsonMap, "level")
		delete(jsonMap, "fields.time")
		if bString, err := json.Marshal(jsonMap); err == nil {
			line = string(bString)
		}
	}
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

// Levels returns list of levels
func (hook *SyslogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func syslogFacility(facility string) b_syslog.Priority {
	switch facility {
	case "kern":
		return b_syslog.LOG_KERN
	case "user":
		return b_syslog.LOG_USER
	case "mail":
		return b_syslog.LOG_MAIL
	case "daemon":
		return b_syslog.LOG_DAEMON
	case "auth":
		return b_syslog.LOG_AUTH
	case "syslog":
		return b_syslog.LOG_SYSLOG
	case "lpr":
		return b_syslog.LOG_LPR
	case "news":
		return b_syslog.LOG_NEWS
	case "uucp":
		return b_syslog.LOG_UUCP
	case "cron":
		return b_syslog.LOG_CRON
	case "authpriv":
		return b_syslog.LOG_AUTHPRIV
	case "ftp":
		return b_syslog.LOG_FTP
	case "local0":
		return b_syslog.LOG_LOCAL0
	case "local1":
		return b_syslog.LOG_LOCAL1
	case "local2":
		return b_syslog.LOG_LOCAL2
	case "local3":
		return b_syslog.LOG_LOCAL3
	case "local4":
		return b_syslog.LOG_LOCAL4
	case "local5":
		return b_syslog.LOG_LOCAL5
	case "local6":
		return b_syslog.LOG_LOCAL6
	case "local7":
		return b_syslog.LOG_LOCAL7
	}
	return 0
}
