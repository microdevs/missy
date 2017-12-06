package service

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/gorilla/mux"
	"github.com/microdevs/missy/log"
	"os"
	"github.com/microdevs/missy/data"
	"github.com/microdevs/missy/config"
	"flag"
	"encoding/json"
	"bytes"
	"os/signal"
	"time"
	gctx "github.com/gorilla/context"
)

type key int

const PrometheusInstance key = 0
const RouterInstance key = 1
const RequestTimer key = 2

type Service struct {
	name         string
	Host 	     string
	Port         string
	Prometheus   *PrometheusHolder
	timer        *Timer
	Router       *mux.Router
}

var listenPort = "8080"
var listenHost = "localhost"
var controllerAddr string

const FlagMissyControllerAddressDefault = "http://missy-controller"
const FlagMissyControllerUsage = "The address of the MiSSy controller"

func init() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initCmd.StringVar(&controllerAddr, "addr", FlagMissyControllerAddressDefault, FlagMissyControllerUsage)
	initCmd.StringVar(&controllerAddr, "a", FlagMissyControllerAddressDefault,  FlagMissyControllerUsage + " (Shorthand)")

	if len(os.Args) > 1 && os.Args[1] == "init" {
		initCmd.Parse(os.Args[2:])
		c := config.GetInstance()
		cjson, jsonErr := json.Marshal(c)
		if jsonErr != nil {
			fmt.Println("Error marshalling config to json.")
			os.Exit(1)
		}
		log.Infof("Registering service %s with MiSSy controller at %s", c.Name, controllerAddr)
		_, err := http.Post(controllerAddr + "/registerService", "application/json", bytes.NewReader(cjson))
		// todo: check response for return status
		if err != nil {
			fmt.Printf("Can not reach missy controller: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

}

// get a new Service object
func New() *Service {

	if _, present := os.LookupEnv("LISTEN_HOST"); present {
		listenHost = os.Getenv("LISTEN_HOST")
	}

	if _, present := os.LookupEnv("LISTEN_PORT"); present {
		listenPort = os.Getenv("LISTEN_PORT")
	}

	c := config.GetInstance()

	return &Service{
		name: c.Name,
		Host: listenHost,
		Port: listenPort,
		Prometheus: NewPrometheus(c.Name),
		Router: mux.NewRouter()}
}

// start http server
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
		err := h.ListenAndServe();
		if (err != nil) {
			log.Fatalf("Error starting Service due to %v", err)
		}
	}()

	//wait for SIGTERM
	<- stop
	// we linebreak here just to get the log message pringted nicely
	fmt.Print("\n")
	log.Warnf("Service shutting down...")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	//TODO: build some connection drainer for websockets
	h.Shutdown(ctx)
	log.Infof("Service stopped gracefully.")
}

func (s *Service) prepareBeforeStart() {
	s.timer = NewTimer()
	s.Router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	s.Router.HandleFunc("/health", s.healthHandler).Methods("GET")
	s.Router.HandleFunc("/info", s.infoHandler).Methods("GET")

	http.Handle("/", s.Router)
}

// handle func wrapper with token validation, logging recovery and metrics
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
		chain.ServeHTTP(w,r)

	}
	return s.Router.HandleFunc(pattern, h)
}

func (s *Service) infoHandler(w http.ResponseWriter, r *http.Request) {
	info := fmt.Sprintf("Name %s\nUptime %s", s.name, s.timer.Uptime())
	w.Write([]byte(info))
}

func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (s Service) finalizeRequest(w *ResponseWriter, r *http.Request, timer *Timer) {
	if err := recover(); err != nil {
		stack := make([]byte, 1024 * 8)
		stack = stack[:runtime.Stack(stack, false)]
		log.Error("PANIC: %s\n%s", err, stack)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		w.Status = http.StatusInternalServerError
	}
	s.Prometheus.OnRequestFinished(r.Method, r.URL.Path, w.Status, timer.durationMillis())
	// access log
	log.Infof("%s \"%s %s %s\" %d - %s", r.RemoteAddr, r.Method, r.URL, r.Proto, w.Status, r.UserAgent())
}

type ResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriter) Marshal(r *http.Request, subject interface{}) {
	resp, err := data.MarshalResponse(w, r, subject)

	if err != nil {
		http.Error(w, "Unexpected Error: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(resp)
}

func (w *ResponseWriter) Error(error string, code int) {
	http.Error(w, error, code)
	w.Status = code
}

type HandlerFunc func(*ResponseWriter, *http.Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w *ResponseWriter, r *http.Request) {
	f(w, r)
}

type Handler interface {
	ServeHTTP(w *ResponseWriter, r *http.Request)
}
