package service

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/microdevs/missy/log"
	"net/http"
)

// StartTimerHandler is a middleware to start a timer for the request benchmark
func StartTimerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, RequestTimer, NewTimer())
		h.ServeHTTP(w, r)
	})
}

// AccessLogHandler writes an acccess logon after the response has been sent
func AccessLogHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw := &ResponseWriter{w, http.StatusOK}
		h.ServeHTTP(mw, r)
		log.Infof("%s \"%s %s %s\" %d - %s", r.RemoteAddr, r.Method, r.URL, r.Proto, mw.Status, r.UserAgent())
	})
}

// StopTimerHandler measures the time of the request with the help of the timestamp taken in StartTimerHandler and writes it to a Prometheus metric
func StopTimerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw := &ResponseWriter{w, http.StatusOK}
		h.ServeHTTP(mw, r)
		timer := context.Get(r, RequestTimer).(*Timer)
		prometheus := context.Get(r, PrometheusInstance).(*PrometheusHolder)
		prometheus.OnRequestFinished(r.Method, r.URL.Path, mw.Status, timer.durationMillis())
	})
}

// infoHandler
func (s *Service) infoHandler(w http.ResponseWriter, r *http.Request) {
	info := fmt.Sprintf("Name %s\nUptime %s", s.name, s.timer.Uptime())
	w.Write([]byte(info))
}

// healthHandler
func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
