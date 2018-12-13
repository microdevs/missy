package middleware

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/microdevs/missy/log"
)

// Logger allows to send logs about HTTP requests.
func Logger(l log.FieldsLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return middleware.RequestLogger(&middleware.DefaultLogFormatter{
			Logger:  l,
			NoColor: false,
		})(next)
	}
}
