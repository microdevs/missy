package log

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

// DummyLogger returns a logger which logs nothing. It is ideal for testing purposes.
func DummyLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(ioutil.Discard)
	return l
}
