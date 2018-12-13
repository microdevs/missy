package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/microdevs/missy/log"
)

type server struct {
	server *http.Server
	router Router

	certFile string
	keyFile  string

	l log.FieldsLogger
}

func newServer(c Config, l log.FieldsLogger) *server {
	return &server{
		certFile: c.TLSCertFile,
		keyFile:  c.TLSKeyFile,
		server: &http.Server{
			Addr: c.Listen,
		},
		router: chi.NewRouter(),
		l:      l,
	}
}

// Routes allow to modify created Router.
// You may setup here your endpoints wrapped with middlewares.
func (s *server) Routes(f func(r Router) error) error {
	return f(s.router)
}

// ListenAndServe allows to start the Server.
func (s *server) ListenAndServe() error {
	s.server.Handler = s.router
	tls := s.certFile != "" || s.keyFile != ""
	s.l.Infof("starting server at '%s', TLS=%v", s.server.Addr, tls)
	if tls {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	return s.server.ListenAndServe()
}

func (s *server) Shutdown(ctx context.Context) {
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.l.Errorf("stopping server (shutdown) err: %s", err)
	}
	s.l.Infof("server stopped listenning at '%s'", s.server.Addr)
}
