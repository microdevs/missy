package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/microdevs/missy/http/metrics"
	"github.com/microdevs/missy/http/middleware"
	"github.com/microdevs/missy/log"
)

type Router chi.Router

// Server defines the http server
type Server interface {
	ListenAndServe(context.Context, context.CancelFunc)
}

type Servers struct {
	internal *serverInternal
	external *server

	tlsCertFile string
	tlsKeyFile  string

	shutdown time.Duration
	l        log.FieldsLogger
}

type Config struct {
	Listen         string `env:"HTTP_SERVER_LISTEN" envDefault:"localhost:8080"`
	InternalListen string `env:"HTTP_SERVER_INTERNAL_LISTEN" envDefault:"localhost:5000"`

	TLSCertFile string `env:"HTTP_SERVER_CERTFILE"`
	TLSKeyFile  string `env:"HTTP_SERVER_KEYFILE"`

	Metrics  bool          `env:"HTTP_SERVER_METRICS" envDefault:"true"`
	Shutdown time.Duration `env:"HTTP_SERVER_SHUTDOWN_TIMEOUT" envDefault:"30s"`
}

// New returns a new HTTP server object.
func New(name string, c Config, l log.FieldsLogger) *Servers {
	s := &Servers{
		internal: newServerInternal(c, l.WithField("http", []string{"server", "internal"})),
		external: newServer(c, l.WithField("http", []string{"server", "external"})),
		shutdown: c.Shutdown,
		l:        l,
	}
	go func() {
		err := s.internal.ListenAndServe()
		if err != http.ErrServerClosed {
			l.Fatalf("internal server listen and serve err: %s", err)
		}
	}()
	return s
}

func (s *Servers) ListenAndServe(ctx context.Context, cancel context.CancelFunc) {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	s.SetReadiness(true)
	go func() {
		err := s.external.ListenAndServe()
		if err != http.ErrServerClosed {
			s.l.Errorf("external server listen/serve err: %s", err)
			s.SetHealth(false)
			s.SetReadiness(false)
		}
	}()

	<-signals
	s.SetReadiness(false)
	cancel()

	sthCtx, sthCancel := context.WithTimeout(context.Background(), s.shutdown)
	defer sthCancel()
	s.external.Shutdown(sthCtx)
	s.internal.Shutdown(sthCtx)
}

// RoutesRaw method allows to define routes without any standard builtin middlewares.
func (s *Servers) RoutesRaw(f func(r Router) error) error {
	return s.external.Routes(f)
}

// Routes automatically add logger, recover and metrics functionalities.
// If you don't want them, please use RoutesRaw method.
func (s *Servers) Routes(f func(Router) error) error {
	return s.external.Routes(func(r Router) error {
		r.Use(chimiddleware.RealIP)
		r.Use(middleware.Tracing)
		r.Use(middleware.Logger(s.l))
		r.Use(middleware.Recoverer(s.l))
		return f(RouterWithMetrics(r, metrics.NewHTTPRequestsCollector()))
	})
}

func (s *Servers) SetHealth(health bool) {
	s.internal.SetHealth(health)
}

func (s *Servers) SetReadiness(ready bool) {
	s.internal.SetReadiness(ready)
}
