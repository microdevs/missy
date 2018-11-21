package logging

import (
	"log"

	"github.com/sirupsen/logrus"
)

// Format allows to define logger format.
type Format string

// Level allows to define output message format.
type Level string

const (
	// FormatJSON defines JSON output format.
	FormatJSON Format = "json"
	// FormatText defines output format as plaintext.
	FormatText Format = "text"

	// LevelPanic defines logger message level as 'panic'.
	LevelPanic Level = "panic"
	// LevelFatal defines logger message level as 'fatal'.
	LevelFatal Level = "fatal"
	// LevelError defines logger message level as 'error'.
	LevelError Level = "error"
	// LevelWarning defines logger message level as 'warn'.
	LevelWarning Level = "warn"
	// LevelInfo defines logger message level as 'info'.
	LevelInfo Level = "info"
	// LevelDebug defines logger message level as 'debug'.
	LevelDebug Level = "debug"
	// LevelTrace defines logger message level as 'trace'.
	LevelTrace Level = "trace"
)

// New creates new instance of Logrus logger.
func New(format Format, level Level) *logrus.Logger {
	l := logrus.New()

	switch format {
	case FormatJSON:
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case FormatText:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	if level == "" {
		level = LevelError
	}
	lvl, err := logrus.ParseLevel(string(level))
	if err != nil {
		log.Fatalf("unknown log level '%s', err: %s", level, err)
	}
	l.SetLevel(lvl)

	return l
}
