package service

import (
	"net/http"
	"github.com/gorilla/context"
	"github.com/microdevs/missy/log"
)


func AccessLogHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := ResponseWriter{w, http.StatusOK}
		h(&rec,r)
		log.Infof("%s \"%s %s %s\" %d - %s", r.RemoteAddr, r.Method, r.URL, r.Proto, rec.Status, r.UserAgent())
	}
}

func MeasureTimeHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := NewTimer()
		rec := ResponseWriter{w, http.StatusOK}
		h(&rec,r)
		prometheus := context.Get(r, PrometheusInstance).(*PrometheusHolder)
		prometheus.OnRequestFinished(r.Method, r.URL.Path, rec.Status, t.durationMillis())
	}
}
