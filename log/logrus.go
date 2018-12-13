package log

import (
	"github.com/pkg/errors"
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

type Config struct {
	Format Format `env:"LOG_FORMAT" envDefault:"json"`
	Level  Level  `env:"LOG_LEVEL" envDefault:"debug"`
}

// New creates new instance of Logrus logger.
func New(c Config) (*logrus.Logger, error) {
	l := logrus.New()

	switch c.Format {
	case FormatJSON:
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case FormatText:
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	if c.Level == "" {
		c.Level = LevelError
	}
	lvl, err := logrus.ParseLevel(string(c.Level))
	if err != nil {
		return nil, errors.Errorf("unknown log level '%s', err: %s", c.Level, err)
	}
	l.SetLevel(lvl)

	return l, nil
}
