package server

import (
	"net/http"

	"github.com/microdevs/missy/http/middleware"
)

type MuxWithMetrics struct {
	Router

	m middleware.MetricRequest
}

func RouterWithMetrics(r Router, m middleware.MetricRequest) *MuxWithMetrics {
	return &MuxWithMetrics{
		Router: r,
		m:      m,
	}
}

func (m *MuxWithMetrics) Connect(pattern string, h http.HandlerFunc) {
	m.router(pattern).Connect(pattern, h)
}

func (m *MuxWithMetrics) Delete(pattern string, h http.HandlerFunc) {
	m.router(pattern).Delete(pattern, h)
}

func (m *MuxWithMetrics) Get(pattern string, h http.HandlerFunc) {
	m.router(pattern).Get(pattern, h)
}

func (m *MuxWithMetrics) Head(pattern string, h http.HandlerFunc) {
	m.router(pattern).Head(pattern, h)
}

func (m *MuxWithMetrics) Options(pattern string, h http.HandlerFunc) {
	m.router(pattern).Options(pattern, h)
}

func (m *MuxWithMetrics) Patch(pattern string, h http.HandlerFunc) {
	m.router(pattern).Patch(pattern, h)
}

func (m *MuxWithMetrics) Post(pattern string, h http.HandlerFunc) {
	m.router(pattern).Post(pattern, h)
}

func (m *MuxWithMetrics) Put(pattern string, h http.HandlerFunc) {
	m.router(pattern).Put(pattern, h)
}

func (m *MuxWithMetrics) router(pattern string) Router {
	if m.m != nil {
		return m.Router.With(middleware.Metrics(pattern, m.m))
	}
	return m.Router
}
