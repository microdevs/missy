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
	// TODO we should split listen and serve and notify about listen by chan
	ListenAndServe(context.Context, context.CancelFunc)
}

type Servers struct {
	name string

	internal *internal
	external *server

	tlsCertFile string
	tlsKeyFile  string

	shutdown time.Duration
	l        log.FieldsLogger
}

type Config struct {
	Name string `env:"HTTP_SERVER_NAME" envDefault:"missy"`

	Listen         string `env:"HTTP_SERVER_LISTEN" envDefault:"localhost:8080"`
	InternalListen string `env:"HTTP_SERVER_INTERNAL_LISTEN" envDefault:"localhost:5000"`

	Metrics bool `env:"HTTP_SERVER_METRICS_ENABLED" envDefault:"true"`

	TLSCertFile string `env:"HTTP_SERVER_CERT_FILE"`
	TLSKeyFile  string `env:"HTTP_SERVER_KEY_FILE"`

	Shutdown time.Duration `env:"HTTP_SERVER_SHUTDOWN_TIMEOUT" envDefault:"30s"`
}

// New returns new Servers component.
func New(c Config, l log.FieldsLogger) *Servers {
	s := &Servers{
		name:     c.Name,
		internal: newInternal(c, l.WithField("http", []string{c.Name, "server", "internal"})),
		external: newServer(c, l.WithField("http", []string{c.Name, "server", "external"})),
		shutdown: c.Shutdown,
		l:        l,
	}
	return s
}

func (s *Servers) ListenAndServe(ctx context.Context, cancel context.CancelFunc) {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// internal server
	go func() {
		err := s.internal.ListenAndServe()
		if err != http.ErrServerClosed {
			s.l.Fatalf("internal server listen and serve err: %s", err)
		}
	}()

	// external server
	go func() {
		// We should set this after external is actually listening
		s.SetReadiness(true)
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
		return f(RouterWithMetrics(r, metrics.NewHTTPRequestsCollector(s.name)))
	})
}

func (s *Servers) SetHealth(health bool) {
	s.internal.SetHealth(health)
}

func (s *Servers) SetReadiness(ready bool) {
	s.internal.SetReadiness(ready)
}
