package log

import (
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/utils"

	"github.com/sirupsen/logrus"
)

type ContextHook struct{}

func (hook ContextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook ContextHook) Fire(entry *logrus.Entry) error {
	var pc []uintptr
	if *utils.LogStackTrace == 1 {
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
				if *utils.LogStackTrace == 0 {
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
