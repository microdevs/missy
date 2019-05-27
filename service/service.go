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
	"strings"
	"sync"
	"syscall"
	"time"

	"bufio"
	"net"

	"github.com/gorilla/mux"
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
	name          string
	Host          string
	Port          string
	MetricsPort   string
	Prometheus    *PrometheusHolder
	timer         *Timer
	Router        *mux.Router
	MetricsRouter *mux.Router
	Stop          chan os.Signal

	StateProbes struct {
		IsHealthy bool
		MuHealthy sync.Mutex
		IsReady   bool
		MuReady   sync.Mutex
	}
}

var controllerAddr string

// FlagMissyControllerAddressDefault is a default for the missy-controller url used in the during service initialisation when given the init flag
const FlagMissyControllerAddressDefault = "http://missy-controller"

// FlagMissyControllerUsage is a usage message for the missy-controller url used in the during service initialisation when given the init flag
const FlagMissyControllerUsage = "The address of the MiSSy controller"

const (
	listenHost        = "service.listen.host"
	listenPort        = "service.listen.port"
	metricsListenPort = "service.metrics.listen.port"
)

// init checks for init flag and executes the service registration with the missy controller if applicable
func init() {

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initCmd.StringVar(&controllerAddr, "addr", FlagMissyControllerAddressDefault, FlagMissyControllerUsage)
	initCmd.StringVar(&controllerAddr, "a", FlagMissyControllerAddressDefault, FlagMissyControllerUsage+" (Shorthand)")

	if len(os.Args) > 1 && os.Args[1] == "init" {
		initCmd.Parse(os.Args[2:])
		cjson, jsonErr := json.Marshal(Config())
		if jsonErr != nil {
			fmt.Println("Error marshalling config to json.")
			os.Exit(1)
		}
		log.Infof("Registering service %s with MiSSy controller at %s", Config().Name, controllerAddr)
		_, err := http.Post(controllerAddr+"/registerService", "application/json", bytes.NewReader(cjson))
		// todo: check response for return status
		if err != nil {
			fmt.Printf("Can not reach missy controller: %s", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	//todo: refactor this to LISTEN_ADDRESS
	config := Config()
	config.RegisterOptionalParameter("LISTEN_HOST", "0.0.0.0", listenHost, "The address the service listens on")
	config.RegisterOptionalParameter("LISTEN_PORT", "8080", listenPort, "The port the service listens on")
	config.RegisterOptionalParameter("METRICS_LISTEN_PORT", "8090", metricsListenPort, "The port the service metrics listens on")
	config.Parse()
}

// New returns a new Service object
func New(name string) *Service {

	// check if name has at least one character
	if len(name) < 1 {
		log.Fatal("Unnamed services are not allowed, passed an empty string as name")
	}
	config := Config()
	config.Name = name

	s := &Service{
		name:          name,
		Host:          config.Get(listenHost),
		Port:          config.Get(listenPort),
		MetricsPort:   config.Get(metricsListenPort),
		Prometheus:    NewPrometheus(name),
		Router:        mux.NewRouter(),
		MetricsRouter: mux.NewRouter(),
	}

	s.StateProbes.IsHealthy = true
	s.StateProbes.IsReady = true
	s.prepareBeforeStart()
	s.Stop = make(chan os.Signal, 1)
	return s
}

// Start starts the http server
func (s *Service) Start() {
	// start the server
	log.Infof("Starting service %s", s.name)
	log.Infof("listening on %s:%s ...", s.Host, s.Port)
	log.Infof("listening for metrics on %s:%s ...", s.Host, s.MetricsPort)
	// set service host and port to listen to
	listen := s.Host + ":" + s.Port
	h := &http.Server{Addr: listen, Handler: s.Router}
	// set service metrics host and port to listen to
	metricsListen := s.Host + ":" + s.MetricsPort
	m := &http.Server{Addr: metricsListen, Handler: s.MetricsRouter}
	// run server in background
	go func() {
		certFile, keyFile, useTLS := prepareTLS()

		errChan := make(chan error, 1)
		if useTLS {
			// listen for main service with TLS
			go listenAndServeTLS(h, certFile, keyFile, errChan)
			// listen for service metrics with TLS
			go listenAndServeTLS(m, certFile, keyFile, errChan)
		} else {
			log.Warnf("WARNING! This server starts without transport layer security (TLS) to use it set TLS_CERTFILE and TLS_KEYFILE in environment")
			// listen for main service without TLS
			go listenAndServe(h, errChan)
			// listen for service metrics without TLS
			go listenAndServe(m, errChan)
		}

		for {
			select {
			case err := <-errChan:
				logFunc := log.Fatalf
				// if server is closed, it's not a fatal but info
				// it means that there was a call to Shutdown or Close
				if err == http.ErrServerClosed {
					logFunc = log.Infof
				}
				logFunc("server shut down due to %v", err)
			}
		}
	}()
	s.prepareShutdown(h, m)
}

func (s *Service) prepareShutdown(h Server, m Server) {
	signal.Notify(s.Stop, os.Interrupt, syscall.SIGTERM)
	<-s.Stop
	s.StatusNotReady()
	// shutdown main service
	shutdown(h)
	// shutdown metrics service
	shutdown(m)
}

// Shutdown allows to stop the HTTP Server gracefully
func (s *Service) Shutdown() {
	s.Stop <- os.Signal(os.Interrupt)
}

func listenAndServeTLS(server *http.Server, certFile, keyFile string, errChan chan error) {
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		errChan <- err
		return
	}
}

func listenAndServe(server *http.Server, errChan chan error) {
	if err := server.ListenAndServe(); err != nil {
		errChan <- err
		return
	}
}

func shutdown(s Shutdowner) {
	if s == nil {
		return
	}

	// todo: make configurable
	timeout := time.Second * 5

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Printf(""+
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
	s.MetricsRouter.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	s.MetricsRouter.HandleFunc("/health", s.healthHandler).Methods(http.MethodGet)
	s.MetricsRouter.HandleFunc("/ready", s.readinessHandler).Methods(http.MethodGet)
	s.MetricsRouter.HandleFunc("/info", s.infoHandler).Methods(http.MethodGet)
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
	h := s.makeHandler(originalHandler, pattern, false)
	return s.Router.Handle(pattern, h)
}

// UnsafeHandle is a wrapper around the original Go handle func with logging recovery and metrics
func (s *Service) UnsafeHandle(pattern string, originalHandler http.Handler) *mux.Route {
	h := s.makeHandler(originalHandler, pattern, false)
	return s.Router.Handle(pattern, h)
}

// SecureHandle is a wrapper around the original Go handle func with logging recovery and metrics
func (s *Service) SecureHandle(pattern string, originalHandler http.Handler) *mux.Route {
	initPublicKey()
	h := s.makeHandler(originalHandler, pattern, true)
	return s.Router.Handle(pattern, h)
}

// Makes a handler that wraps Missy specific functionality and returns either a secure or insecure chain
// a secure chain includes the auth handler
func (s *Service) makeHandler(originalHandler http.Handler, pattern string, secure bool) http.Handler {
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
		ctx := r.Context()
		ctx = context.WithValue(ctx, PrometheusInstance, s.Prometheus)
		ctx = context.WithValue(ctx, RouterInstance, s.Router)
		r = r.WithContext(ctx)
		// call custom handler
		chain := NewChain(StartTimerHandler, FinalHandler(pattern)).Then(originalHandler)
		if secure {
			chain = NewChain(StartTimerHandler, AuthHandler, FinalHandler(pattern)).Then(originalHandler)
		}
		chain.ServeHTTP(w, r)
	})
}

// ResponseWriter is the MiSSy owned response writer object
type ResponseWriter struct {
	http.ResponseWriter
	status    int
	headerSet bool
}

// HijackerResponseWriter is the MiSSy owned response writer which also handles Hijacker
type HijackerResponseWriter struct {
	*ResponseWriter
}

// WriteHeader overrides the original WriteHeader function to keep the status code
func (w *ResponseWriter) WriteHeader(code int) {
	w.status = code
	w.headerSet = true
	w.ResponseWriter.WriteHeader(code)
}

// WriteMetricsHeader just sets the status code for metrics and logging
// Useful for example for hijacked ResponseWriter where WriteHeader and Write are bypassed
func (w *ResponseWriter) WriteMetricsHeader(code int) {
	w.status = code
}

// ResponseWriter wrapper for http.ResponseWriter interface
func (w *ResponseWriter) Write(b []byte) (int, error) {
	if !w.headerSet {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Status is a getter for status
func (w *ResponseWriter) Status() int {
	return w.status
}

// Header wrapper for http.ResponseWriter interface
func (w *ResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Hijack wrapper for http.Hijacker interface
func (w *HijackerResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.ResponseWriter.(http.Hijacker).Hijack()
}

// prepareTLS will look in the os environment for a certfile and a keyfile
// both values will be trimmed for trailing or leading spaces and the files will be
// checked for existence and accessibility. If all is good both file names will be returned and
// the boolean value useTLS will return true
func prepareTLS() (certFile string, keyFile string, useTLS bool) {
	certFile = strings.Trim(os.Getenv("TLS_CERTFILE"), " ")
	if certFile == "" {
		log.Debug("TLS Certfile was not set")
		return "", "", false
	}
	if _, err := os.OpenFile(certFile, os.O_RDONLY, 0600); os.IsNotExist(err) || os.IsPermission(err) {
		log.Warnf("TLS Certfile was set to %s, but cannot be accessed: %s", certFile, err)
		return "", "", false
	}
	keyFile = strings.Trim(os.Getenv("TLS_KEYFILE"), " ")
	if keyFile == "" {
		log.Debug("TLS Keyfile was not set")
		return "", "", false
	}
	if _, err := os.OpenFile(keyFile, os.O_RDONLY, 0600); os.IsNotExist(err) || os.IsPermission(err) {
		log.Warnf("TLS Keyfile was set to %s, but cannot be accessed: %s", keyFile, err)
		return "", "", false
	}

	return certFile, keyFile, true
}

func (s *Service) StatusUnhealthy() {
	s.statusHealthy(false)
}

func (s *Service) StatusHealthy() {
	s.statusHealthy(true)
}

func (s *Service) statusHealthy(v bool) {
	s.StateProbes.MuHealthy.Lock()
	defer s.StateProbes.MuHealthy.Unlock()
	s.StateProbes.IsHealthy = v
}

func (s *Service) StatusNotReady() {
	s.statusReady(false)
}

func (s *Service) StatusReady() {
	s.statusReady(true)
}

func (s *Service) statusReady(v bool) {
	s.StateProbes.MuReady.Lock()
	defer s.StateProbes.MuReady.Unlock()
	s.StateProbes.IsReady = v
}
