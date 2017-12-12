package service

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	gctx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/microdevs/missy/config"
	"github.com/microdevs/missy/data"
	"github.com/microdevs/missy/log"
	"github.com/microdevs/missy/resource"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"
)

type key int

// PrometheusInstance is a key for the context value of the PrometheusHolder
const PrometheusInstance key = 0

// RouterInstance is the key for the goilla/mux router instance in the context
const RouterInstance key = 1

// RequestTimer is the key for the request timer in context
const RequestTimer key = 2

// Service type provides a HTTP/Rest service
type Service struct {
	name       string
	Host       string
	Port       string
	Prometheus *PrometheusHolder
	timer      *Timer
	Router     *mux.Router
}

var listenPort = "8080"
var listenHost = "localhost"
var controllerAddr string

// FlagMissyControllerAddressDefault is a default for the missy-controller url used in the during service initialisation when given the init flag
const FlagMissyControllerAddressDefault = "http://missy-controller"

// FlagMissyControllerUsage is a usage message for the missy-controller url used in the during service initialisation when given the init flag
const FlagMissyControllerUsage = "The address of the MiSSy controller"

// init checks for init flag and executes the service registration with the missy controller if applicable
func init() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initCmd.StringVar(&controllerAddr, "addr", FlagMissyControllerAddressDefault, FlagMissyControllerUsage)
	initCmd.StringVar(&controllerAddr, "a", FlagMissyControllerAddressDefault, FlagMissyControllerUsage+" (Shorthand)")

	if len(os.Args) > 1 && os.Args[1] == "init" {
		initCmd.Parse(os.Args[2:])
		c := config.GetInstance()
		cjson, jsonErr := json.Marshal(c)
		if jsonErr != nil {
			fmt.Println("Error marshalling config to json.")
			os.Exit(1)
		}
		log.Infof("Registering service %s with MiSSy controller at %s", c.Name, controllerAddr)
		_, err := http.Post(controllerAddr+"/registerService", "application/json", bytes.NewReader(cjson))
		// todo: check response for return status
		if err != nil {
			fmt.Printf("Can not reach missy controller: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

}

// New returns a new Service object
func New() *Service {

	if _, present := os.LookupEnv("LISTEN_HOST"); present {
		listenHost = os.Getenv("LISTEN_HOST")
	}

	if _, present := os.LookupEnv("LISTEN_PORT"); present {
		listenPort = os.Getenv("LISTEN_PORT")
	}

	c := config.GetInstance()
	resource.Setup(c)
	c.ParseEnv()

	return &Service{
		name:       c.Name,
		Host:       listenHost,
		Port:       listenPort,
		Prometheus: NewPrometheus(c.Name),
		Router:     mux.NewRouter()}
}

// Start starts the http server
func (s *Service) Start() {
	// Open a channel to capture ^C signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	// start the server
	log.Infof("Starting service %s listening on %s:%s ...", s.name, s.Host, s.Port)
	s.prepareBeforeStart()
	// set host and port to listen to
	listen := s.Host + ":" + s.Port
	h := &http.Server{Addr: listen, Handler: s.Router}
	// run server in background
	go func() {
		err := h.ListenAndServe()
		if err != nil {
			log.Fatalf("Error starting Service due to %v", err)
		}
	}()

	//wait for SIGTERM
	<-stop
	// we linebreak here just to get the log message pringted nicely
	fmt.Print("\n")
	log.Warnf("Service shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Cancel ctx as soon as handleSearch returns.
	//TODO: build some connection drainer for websockets
	h.Shutdown(ctx)
	log.Infof("Service stopped gracefully.")
}

// prepareBeforeStart sets up the standard handlers
func (s *Service) prepareBeforeStart() {
	s.timer = NewTimer()
	s.Router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	s.Router.HandleFunc("/health", s.healthHandler).Methods("GET")
	s.Router.HandleFunc("/info", s.infoHandler).Methods("GET")

	http.Handle("/", s.Router)
}

// HandleFunc wrapper with token validation, logging recovery and metrics
func (s *Service) HandleFunc(pattern string, handler func(*ResponseWriter, *http.Request)) *mux.Route {
	h := func(originalResponseWriter http.ResponseWriter, r *http.Request) {
		// build context
		gctx.Set(r, PrometheusInstance, s.Prometheus)
		gctx.Set(r, RouterInstance, s.Router)
		// use our response writer
		w := &ResponseWriter{originalResponseWriter, http.StatusOK}
		// call custom handler
		handleFunc := HandlerFunc(handler)
		chain := NewChain(StartTimerHandler, AccessLogHandler).Final(StopTimerHandler).Then(handleFunc)
		chain.ServeHTTP(w, r)

	}
	return s.Router.HandleFunc(pattern, h)
}

// finalizeRequest
func (s Service) finalizeRequest(w *ResponseWriter, r *http.Request, timer *Timer) {
	if err := recover(); err != nil {
		stack := make([]byte, 1024*8)
		stack = stack[:runtime.Stack(stack, false)]
		log.Error("PANIC: %s\n%s", err, stack)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		w.Status = http.StatusInternalServerError
	}
	s.Prometheus.OnRequestFinished(r.Method, r.URL.Path, w.Status, timer.durationMillis())
	// access log
	log.Infof("%s \"%s %s %s\" %d - %s", r.RemoteAddr, r.Method, r.URL, r.Proto, w.Status, r.UserAgent())
}

// ResponseWriter is the MiSSy owned response writer object
type ResponseWriter struct {
	http.ResponseWriter
	Status int
}

// WriteHeader overrides the original WriteHeader function to keep the status code
func (w *ResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

// Marshal an object to json or xml according the request Accept header
func (w *ResponseWriter) Marshal(r *http.Request, subject interface{}) {
	resp, err := data.MarshalResponse(w, r, subject)

	if err != nil {
		http.Error(w, "Unexpected Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

// Error adds a shorthand to the http.Error function
func (w *ResponseWriter) Error(error string, code int) {
	http.Error(w, error, code)
	w.Status = code
}

// HandlerFunc provides MiSSys own handler func
type HandlerFunc func(*ResponseWriter, *http.Request)

// ServeHTTP calls f(w, r)
func (f HandlerFunc) ServeHTTP(w *ResponseWriter, r *http.Request) {
	f(w, r)
}

// Handler MiSSy's own Handler interface taking the custom ResponseWriter
type Handler interface {
	ServeHTTP(w *ResponseWriter, r *http.Request)
}
