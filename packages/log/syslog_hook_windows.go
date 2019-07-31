// +build windows

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
