package server

import (
	"fmt"
	"net/http"
	"runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/gorilla/mux"
	"github.com/microdevs/missy/log"
	"os"
)

type Server struct {
	name         string
	Host 	     string
	Port         string
	prometheus   *PrometheusHolder
	timer        *Timer
	router       *mux.Router
}

var listenPort = "8080"
var listenHost = "localhost"

// get a new server object
func NewServer(name string) *Server {

	if _, present := os.LookupEnv("LISTEN_HOST"); present {
		listenHost = os.Getenv("LISTEN_HOST")
	}

	if _, present := os.LookupEnv("LISTEN_PORT"); present {
		listenPort = os.Getenv("LISTEN_PORT")
	}

	return &Server{
		name: name,
		Host: listenHost,
		Port: listenPort,
		prometheus: NewPrometheus(name),
		router: mux.NewRouter()}
}

// start http server
func (s *Server) StartServer() error {
	log.Warnf("Starting service %s listening on %s:%s ...", s.name, s.Host, s.Port)
	s.prepareBeforeStart()
	listen := s.Host + ":" + s.Port
	err := http.ListenAndServe(listen, nil); if (err != nil) {
		log.Fatalf("Error starting Server due to %v", err)
	}
	log.Warn("...shutting down.")
	return err
}

func (s *Server) prepareBeforeStart() {
	s.timer = NewTimer()
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	s.router.HandleFunc("/health", s.healthHandler).Methods("GET")
	s.router.HandleFunc("/info", s.infoHandler).Methods("GET")

	http.Handle("/", s.router)
}

// handle func wrapper with token validation, logging recovery and metrics
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	h := func(originalResponseWriter http.ResponseWriter, r *http.Request) {
		timer := NewTimer()
		// use our response writer
		w := &ResponseWriter{ResponseWriter: originalResponseWriter, status: 200}
		defer s.finalizeRequest(w, r, timer)
		// call custom handler
		handler(w, r)
	}
	return s.router.HandleFunc(pattern, h)
}

func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	info := fmt.Sprintf("Name %s\nUptime %s", s.name, s.timer.Uptime())
	w.Write([]byte(info))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (s Server) finalizeRequest(w *ResponseWriter, r *http.Request, timer *Timer) {
	if err := recover(); err != nil {
		stack := make([]byte, 1024 * 8)
		stack = stack[:runtime.Stack(stack, false)]
		log.Error("PANIC: %s\n%s", err, stack)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		//log.Info("%s \"%s %s %s\" %d", r.RemoteAddr, r.Method, r.URL, r.Proto, w.status)
	}
	s.prometheus.OnRequestFinished(r.Method, r.URL.Path, w.status, timer.durationMillis())
	// access log
	log.Infof("%s \"%s %s %s\" %d - %s", r.RemoteAddr, r.Method, r.URL, r.Proto, w.status, r.UserAgent())
}

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}