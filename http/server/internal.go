package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/microdevs/missy/http/handlers/health"
	"github.com/microdevs/missy/http/handlers/readiness"
	"github.com/microdevs/missy/log"
)

type internal struct {
	server

	health    *health.Handler
	readiness *readiness.Handler
}

func newInternal(c Config, l log.FieldsLogger) *internal {
	s := &internal{
		// TODO what about TLS on internal external?
		server: server{
			server: &http.Server{
				Addr: c.InternalListen,
			},
			router: chi.NewRouter(),
			l:      l,
		},
		health:    health.NewHandler(),
		readiness: readiness.NewHandler(),
	}
	s.router.Get("/health", s.health.HandleGet)
	s.router.Get("/ready", s.readiness.HandleGet)
	if c.Metrics {
		s.router.Mount("/metrics", promhttp.Handler())
	}
	return s
}

func (i *internal) SetHealth(health bool) {
	i.health.Set(health)
}

func (i *internal) SetReadiness(ready bool) {
	i.readiness.Set(ready)
}
