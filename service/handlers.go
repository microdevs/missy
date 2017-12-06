package service

import (
	"net/http"
	"github.com/gorilla/context"
	"github.com/microdevs/missy/log"
)

func StartTimerHandler(h Handler) Handler {
	return HandlerFunc(func(w *ResponseWriter, r *http.Request) {
		context.Set(r, RequestTimer, NewTimer())
		h.ServeHTTP(w,r)
	})
}

func AccessLogHandler(h Handler) Handler {
	return HandlerFunc(func(w *ResponseWriter, r *http.Request) {
		log.Infof("%s \"%s %s %s\" %d - %s", r.RemoteAddr, r.Method, r.URL, r.Proto, w.Status, r.UserAgent())
		h.ServeHTTP(w,r)
	})
}

func StopTimerHandler(h Handler) Handler {
	return HandlerFunc(func(w *ResponseWriter, r *http.Request) {
		h.ServeHTTP(w,r)
		timer := context.Get(r, RequestTimer).(*Timer)
		prometheus := context.Get(r, PrometheusInstance).(*PrometheusHolder)
		prometheus.OnRequestFinished(r.Method, r.URL.Path, w.Status, timer.durationMillis())
	})
}
