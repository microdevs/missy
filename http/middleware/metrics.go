package middleware

import (
	"net/http"
	"time"
)

// MetricRequest should allow to Collect metrics about the HTTP request.
type MetricRequest interface {
	CollectDuration(method, name string, status int, duration time.Duration)
}

// Metrics is a middleware for sending metrics about HTTP requests.
func Metrics(name string, m MetricRequest) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if m == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			mw := &responseWriter{ResponseWriter: w}

			timeStart := time.Now()
			next.ServeHTTP(mw, req)
			timeEnd := time.Now()

			m.CollectDuration(req.Method, name, mw.code(), timeEnd.Sub(timeStart))
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *responseWriter) code() int {
	code := w.status
	if code == 0 {
		code = http.StatusOK
	}
	return code
}
