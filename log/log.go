package log

import (
	l "github.com/sirupsen/logrus"
	"os"
)

// init logrus with environment values for log level and log format
func init() {
	// set log format
	format := os.Getenv("LOG_FORMAT")
	switch format {
	case "json":
		l.SetFormatter(&l.JSONFormatter{})
		break
	case "text":
		l.SetFormatter(&l.TextFormatter{})
		break
	case "":
		// nothing
		break
	default:
		Fatalf("Unkown log format \"%s\"", format)
	}

	// setting level, default debug
	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel == "" {
		loglevel = "debug"
	}
	level, error := l.ParseLevel(loglevel)
	if error != nil {
		Fatalf("Unknown loglevel %s, allowed levels: debug, info, warn, error, fatal, panic")
	}
	l.SetLevel(level)
	l.Debugf("Setting Loglevel to %s", level.String())

}

// Wrapper around logrus
func Debug(args ...interface{}) {
	l.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

func Debugln(args ...interface{}) {
	l.Debugln(args...)
}

func Info(args ...interface{}) {
	l.Info(args...)
}

func Infof(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func Infoln(args ...interface{}) {
	l.Infoln(args...)
}

func Warn(args ...interface{}) {
	l.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

func Warnln(args ...interface{}) {
	l.Warnln(args...)
}

func Error(args ...interface{}) {
	l.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

func Errorln(args ...interface{}) {
	l.Errorln(args...)
}

func Fatal(args ...interface{}) {
	l.Error(args...)
}

func Fatalf(format string, args ...interface{}) {
	l.Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	l.Fatalln(args...)
}

func Panic(args ...interface{}) {
	l.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	l.Panicf(format, args...)
}

func Panicln(args ...interface{}) {
	l.Panicln(args...)
}

func Print(args ...interface{}) {
	l.Print(args...)
}

func Printf(format string, args ...interface{}) {
	l.Printf(format, args...)
}

func Println(args ...interface{}) {
	l.Println(args...)
}
