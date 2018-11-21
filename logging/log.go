package logging

// Logger defines our set of basic logging methods.
type Logger interface {
	Fatalf(format string, v ...interface{})
	Fatal(...interface{})

	Errorf(format string, v ...interface{})
	Error(...interface{})

	Warnf(format string, v ...interface{})
	Warn(...interface{})

	Infof(format string, v ...interface{})
	Info(...interface{})

	Debugf(format string, v ...interface{})
	Debug(...interface{})
}

// LoggerWithFields allows to create instances of loggers with additional fields.
// For example, you may define a set of tags for all logged messages.
type LoggerWithFields interface {
	Logger

	WithField(key string, value interface{}) LoggerWithFields
}
