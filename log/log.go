package log

import (
	"log"

	"github.com/sirupsen/logrus"
)

type FieldsLogger logrus.FieldLogger

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
