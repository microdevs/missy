package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/microdevs/missy/log"
)

// Recoverer logs panics if any occurred during processing HTTP request.
func Recoverer(l log.FieldsLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					l.Errorf("Panic: %+v, Stack: %s", rvr, string(debug.Stack()))
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
