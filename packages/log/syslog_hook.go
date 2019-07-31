// +build !windows,!nacl,!plan9

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

package log

import (
	"encoding/json"
	"fmt"
	"os"

	b_syslog "github.com/blackjack/syslog"
	"github.com/sirupsen/logrus"
)

var syslogFacilityPriority map[string]b_syslog.Priority

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
	return syslogFacilityPriority[facility]
}

func init() {
	syslogFacilityPriority = map[string]b_syslog.Priority{
		"kern":     b_syslog.LOG_KERN,
		"user":     b_syslog.LOG_USER,
		"mail":     b_syslog.LOG_MAIL,
		"daemon":   b_syslog.LOG_DAEMON,
		"auth":     b_syslog.LOG_AUTH,
		"syslog":   b_syslog.LOG_SYSLOG,
		"lpr":      b_syslog.LOG_LPR,
		"news":     b_syslog.LOG_NEWS,
		"uucp":     b_syslog.LOG_UUCP,
		"cron":     b_syslog.LOG_CRON,
		"authpriv": b_syslog.LOG_AUTHPRIV,
		"ftp":      b_syslog.LOG_FTP,
		"local0":   b_syslog.LOG_LOCAL0,
		"local1":   b_syslog.LOG_LOCAL1,
		"local2":   b_syslog.LOG_LOCAL2,
		"local3":   b_syslog.LOG_LOCAL3,
		"local4":   b_syslog.LOG_LOCAL4,
		"local5":   b_syslog.LOG_LOCAL5,
		"local6":   b_syslog.LOG_LOCAL6,
		"local7":   b_syslog.LOG_LOCAL7,
	}
}
