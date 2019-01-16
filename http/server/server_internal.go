package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/microdevs/missy/http/handlers/health"
	"github.com/microdevs/missy/http/handlers/readiness"
	"github.com/microdevs/missy/log"
)

type serverInternal struct {
	server

	health    *health.Handler
	readiness *readiness.Handler
}

func newServerInternal(c Config, l log.FieldsLogger) *serverInternal {
	s := &serverInternal{
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

func (si *serverInternal) SetHealth(health bool) {
	si.health.Set(health)
}

func (si *serverInternal) SetReadiness(ready bool) {
	si.readiness.Set(ready)
}
