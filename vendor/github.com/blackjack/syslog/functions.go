package syslog

// Writes syslog message with log level EMERG
func Emerg(msg string) {
	Syslog(LOG_EMERG, msg)
}

// Formats according to a format specifier and writes to syslog with log level EMERG
func Emergf(format string, a ...interface{}) {
	Syslogf(LOG_EMERG, format, a...)
}

// Writes syslog message with log level ALERT
func Alert(msg string) {
	Syslog(LOG_ALERT, msg)
}

// Formats according to a format specifier and writes to syslog with log level ALERT
func Alertf(format string, a ...interface{}) {
	Syslogf(LOG_ALERT, format, a...)
}

// Writes syslog message with log level CRIT
func Crit(msg string) {
	Syslog(LOG_CRIT, msg)
}

// Formats according to a format specifier and writes to syslog with log level CRIT
func Critf(format string, a ...interface{}) {
	Syslogf(LOG_CRIT, format, a...)
}

// Writes syslog message with log level ERR
func Err(msg string) {
	Syslog(LOG_ERR, msg)
}

// Formats according to a format specifier and writes to syslog with log level ERR
func Errf(format string, a ...interface{}) {
	Syslogf(LOG_ERR, format, a...)
}

// Writes syslog message with log level WARNING
func Warning(msg string) {
	Syslog(LOG_WARNING, msg)
}

// Formats according to a format specifier and writes to syslog with log level WARNING
func Warningf(format string, a ...interface{}) {
	Syslogf(LOG_WARNING, format, a...)
}

// Writes syslog message with log level NOTICE
func Notice(msg string) {
	Syslog(LOG_NOTICE, msg)
}

// Formats according to a format specifier and writes to syslog with log level NOTICE
func Noticef(format string, a ...interface{}) {
	Syslogf(LOG_NOTICE, format, a...)
}

// Writes syslog message with log level INFO
func Info(msg string) {
	Syslog(LOG_INFO, msg)
}

// Formats according to a format specifier and writes to syslog with log level INFO
func Infof(format string, a ...interface{}) {
	Syslogf(LOG_INFO, format, a...)
}

// Writes syslog message with log level DEBUG
func Debug(msg string) {
	Syslog(LOG_DEBUG, msg)
}

// Formats according to a format specifier and writes to syslog with log level DEBUG
func Debugf(format string, a ...interface{}) {
	Syslogf(LOG_DEBUG, format, a...)
}
