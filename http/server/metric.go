package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/microdevs/missy/http/handlers/health"
	"github.com/microdevs/missy/http/handlers/readiness"
	"github.com/microdevs/missy/log"
)

type metric struct {
	server

	health    *health.Handler
	readiness *readiness.Handler
}

func newMetric(c Config, l log.FieldsLogger) *metric {
	s := &metric{
		// TODO what about TLS on metric external?
		server: server{
			server: &http.Server{
				Addr: c.MetricListen,
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

func (m *metric) SetHealth(health bool) {
	m.health.Set(health)
}

func (m *metric) SetReadiness(ready bool) {
	m.readiness.Set(ready)
}
