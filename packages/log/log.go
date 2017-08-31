package log

import (
	"runtime"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.New()
)

//logrus levels: debug, info, warn, error, fatal

func LogDebug(errorType consts.LogEventType, logData interface{}) {
	if logrus.GetLevel() < logrus.DebugLevel {
		return
	}

	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":    file,
			"line":    line,
			"errData": logData.(string)}).
			Debug(consts.LogEventsMap[errorType])
	} else {
		logger.WithFields(logrus.Fields{
			"errPlace": "?",
			"errData":  logData.(string)}).
			Debug(consts.LogEventsMap[errorType])
	}
}

func LogInfo(errorType consts.LogEventType, logData interface{}) {
	if logrus.GetLevel() < logrus.InfoLevel {
		return
	}

	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":    file,
			"line":    line,
			"errData": logData.(string)}).
			Info(consts.LogEventsMap[errorType])
	} else {
		logger.WithFields(logrus.Fields{
			"errPlace": "?",
			"errData":  logData.(string)}).
			Info(consts.LogEventsMap[errorType])
	}

}

func LogWarn(errorType consts.LogEventType, logData interface{}) {
	if logrus.GetLevel() < logrus.WarnLevel {
		return
	}

	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":    file,
			"line":    line,
			"errData": logData.(string)}).
			Warn(consts.LogEventsMap[errorType])
	} else {
		logger.WithFields(logrus.Fields{
			"errPlace": "?",
			"errData":  logData.(string)}).
			Warn(consts.LogEventsMap[errorType])
	}
}

func LogError(errorType consts.LogEventType, logData interface{}) {
	if logrus.GetLevel() < logrus.ErrorLevel {
		return
	}

	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":    file,
			"line":    line,
			"errData": logData.(string)}).
			Error(consts.LogEventsMap[errorType])
	} else {
		logger.WithFields(logrus.Fields{
			"errPlace": "?",
			"errData":  logData.(string)}).
			Error(consts.LogEventsMap[errorType])
	}
}

func LogFatal(errorType consts.LogEventType, logData interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":    file,
			"line":    line,
			"errData": logData.(string)}).
			Fatal(consts.LogEventsMap[errorType])
	} else {
		logger.WithFields(logrus.Fields{
			"errPlace": "?",
			"errData":  logData.(string)}).
			Fatal(consts.LogEventsMap[errorType])
	}
}
