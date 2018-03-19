//Package service contains the logic to build a HTTP/Rest Service for the MiSSy runtime environment
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	gctx "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/microdevs/missy/config"
	"github.com/microdevs/missy/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server is implemented by *http.Server
type Server interface {
	ListenAndServe() error
	Shutdowner
}

// TLSServer is implemented by *http.Server
type TLSServer interface {
	ListenAndServeTLS(string, string) error
	Shutdowner
}

// Shutdowner is implemented by *http.Server, and optionally by *http.Server.Handler
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

// signals is the channel used to signal shutdown
var signals chan os.Signal

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
	Stop       chan os.Signal
	ServeMux   *http.ServeMux
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

	s := &Service{
		name:       c.Name,
		Host:       listenHost,
		Port:       listenPort,
		Prometheus: NewPrometheus(c.Name),
		Router:     mux.NewRouter(),
		ServeMux:   http.NewServeMux(),
	}
	s.prepareBeforeStart()
	return s
}

// Start starts the http server
func (s *Service) Start() {
	// start the server
	log.Infof("Starting service %s listening on %s:%s ...", s.name, s.Host, s.Port)
	// set host and port to listen to
	listen := s.Host + ":" + s.Port
	h := &http.Server{Addr: listen, Handler: s.Router}
	// run server in background
	go func() {
		err := h.ListenAndServe()
		if err != nil {
			log.Fatalf("Error starting Service due to %v", err)
		}
		log.Debugf("server shut down")
	}()

	s.prepareShutdown(h)
}

func (s *Service) prepareShutdown(h Server) {
	s.Stop = make(chan os.Signal, 1)
	signal.Notify(s.Stop, os.Interrupt, syscall.SIGTERM)
	<-s.Stop
	shutdown(h)
}

// Shutdown allows to stop the HTTP Server gracefully
func (s *Service) Shutdown() {
	s.Stop <- os.Signal(os.Interrupt)
}

func shutdown(s Shutdowner) {
	if s == nil {
		return
	}

	// todo: make configurable
	timeout := time.Second * 5

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Printf("" +
		"Server shutdown with timeout: %s", timeout)

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Error: %v", err)
	} else {
		if hs, ok := s.(*http.Server); ok {
			log.Printf("Finished all in-flight HTTP requests")

			if hss, ok := hs.Handler.(Shutdowner); ok {
				select {
				case <-ctx.Done():
					if err := ctx.Err(); err != nil {
						log.Printf("Error: %v", err)
						return
					}
				default:
					if deadline, ok := ctx.Deadline(); ok {
						secs := (time.Until(deadline) + time.Second/2) / time.Second
						log.Printf("Shutting down handler with timeout: %ds", secs)
					}

					done := make(chan error)

					go func() {
						<-ctx.Done()
						done <- ctx.Err()
					}()

					go func() {
						done <- hss.Shutdown(ctx)
					}()

					if err := <-done; err != nil {
						log.Printf("Error: %v", err)
						return
					}
				}
			}
		}

		if deadline, ok := ctx.Deadline(); ok {
			secs := (time.Until(deadline) + time.Second/2) / time.Second
			log.Printf("Shutdown finished %ds before deadline", secs)
		}
	}
}

// prepareBeforeStart sets up the standard handlers
func (s *Service) prepareBeforeStart() {
	s.timer = NewTimer()
	s.Router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	s.Router.HandleFunc("/health", s.healthHandler).Methods("GET")
	s.Router.HandleFunc("/info", s.infoHandler).Methods("GET")

	s.ServeMux.Handle("/", s.Router)

	initPublicKey()
}

// HandleFunc excepts a HanderFunc an converts it to a handler, then registers this handler
// Deprecated: Developers should use SecureHandleFunc() or UnsafeHandleFunc() explicitly
func (s *Service) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	return s.UnsafeHandle(pattern, http.HandlerFunc(handler))
}

// UnsafeHandleFunc excepts a HanderFunc an converts it to a handler, then registers this handler
func (s *Service) UnsafeHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	return s.UnsafeHandle(pattern, http.HandlerFunc(handler))
}

// SecureHandleFunc excepts a HanderFunc an converts it to a handler, then registers this handler
func (s *Service) SecureHandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	return s.SecureHandle(pattern, http.HandlerFunc(handler))
}

// Handle is a wrapper around the original Go handle func with logging recovery and metrics
// Deprecated: Developers should use SecureHandle() or UnsafeHandle() explicitly
func (s *Service) Handle(pattern string, originalHandler http.Handler) *mux.Route {
	h := s.makeHandler(originalHandler, false)
	return s.Router.Handle(pattern, h)
}

// UnsafeHandle is a wrapper around the original Go handle func with logging recovery and metrics
func (s *Service) UnsafeHandle(pattern string, originalHandler http.Handler) *mux.Route {
	h := s.makeHandler(originalHandler, false)
	return s.Router.Handle(pattern, h)
}

// SecureHandle is a wrapper around the original Go handle func with logging recovery and metrics
func (s *Service) SecureHandle(pattern string, originalHandler http.Handler) *mux.Route {
	initPublicKey()
	h := s.makeHandler(originalHandler, true)
	return s.Router.Handle(pattern, h)
}

// Makes a handler that wraps Missy specific functionality and returns either a secure or insecure chain
// a secure chain includes the auth handler
func (s *Service) makeHandler(originalHandler http.Handler, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]
				log.Errorf("PANIC: %s\n%s", err, stack)
				http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			}
		}()
		// build context
		gctx.Set(r, PrometheusInstance, s.Prometheus)
		gctx.Set(r, RouterInstance, s.Router)
		// call custom handler
		chain := NewChain(StartTimerHandler).Final(FinalHandler).Then(originalHandler)
		if secure {
			chain = NewChain(StartTimerHandler, AuthHandler).Final(FinalHandler).Then(originalHandler)
		}
		chain.ServeHTTP(w, r)
	})
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

// ResponseWriter wrapper for http.ResponseWriter interface
func (w *ResponseWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

// Header wrapper for http.ResponseWriter interface
func (w *ResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}
