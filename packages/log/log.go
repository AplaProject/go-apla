package log

import (
	"runtime"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.New()
)

func LogInfo(errorType int, logData interface{}) {
	switch errorType {
	case consts.StrtoInt64Error:
		if _, file, line, ok := runtime.Caller(1); ok {
			logger.WithFields(logrus.Fields{
				"file":    file,
				"line":    line,
				"errData": logData.(string)}).
				Info(consts.StrToInt64Message)
		} else {
			logger.WithFields(logrus.Fields{
				"errPlace": "?",
				"errData":  logData.(string)}).
				Info(consts.StrToInt64Message)
		}
	}
}
