//MIT License
//
//Copyright (c) 2016-2018 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package log

import (
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"

	"github.com/sirupsen/logrus"
)

// ContextHook storing nothing but behavior
type ContextHook struct{}

// Levels returns all log levels
func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire the log entry
func (hook ContextHook) Fire(entry *logrus.Entry) error {
	var pc []uintptr
	if *conf.LogStackTrace {
		pc = make([]uintptr, 15, 15)
	} else {
		pc = make([]uintptr, 4, 4)
	}
	cnt := runtime.Callers(6, pc)

	count := 0
	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !strings.Contains(name, "github.com/sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			if count == 0 {
				entry.Data["file"] = path.Base(file)
				entry.Data["func"] = path.Base(name)
				entry.Data["line"] = line
				entry.Data["time"] = time.Now().Format(time.RFC3339)
				if !*conf.LogStackTrace {
					break
				}
			}
			if count >= 1 {
				if count == 1 {
					entry.Data["from"] = []string{}
				}
				entry.Data["from"] = append(entry.Data["from"].([]string), path.Base(name))
			}
			count += 1
		}
	}
	return nil
}
