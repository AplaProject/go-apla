package log

import (
	"fmt"
	"os"
	"runtime"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/sirupsen/logrus"
)

type LogLevel uint32

const (
	Debug LogLevel = iota
	Info           = iota
	Warn           = iota
	Error          = iota
	Fatal          = iota
)

var (
	logger = logrus.New()
)

func WriteToFile(fileName string) error {
	openMode := os.O_APPEND
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		openMode = os.O_CREATE
	}

	f, err := os.OpenFile(fileName, os.O_WRONLY|openMode, 0755)
	if err != nil {
		fmt.Println("Can't open log file ", fileName)
		return err
	}
	logrus.SetOutput(f)
	return nil
}

func SetLevel(level LogLevel) {
	logrus.SetLevel(logrus.Level(level))
}

func WriteToConsole() {
	logrus.SetOutput(os.Stdout)
}

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
