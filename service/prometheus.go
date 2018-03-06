package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
)

var holder *PrometheusHolder

// PrometheusHolder is holds an instances of globally and internally used Prometheus metrics
type PrometheusHolder struct {
	httpLatency *prometheus.HistogramVec
}

// NewPrometheus returns a new instance of a Prometheus holder to bind to a service
func NewPrometheus(serviceName string) *PrometheusHolder {
	if holder != nil {
		return holder
	}

	prefix := strings.Replace(serviceName, "-", "_", -1)

	httpLatency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: prefix + "_http_handler_latency",
		Help: "HTTP Handler Latency by endpoint",
	},
		[]string{"method", "path", "status_code"},
	)

	prometheus.MustRegister(httpLatency)

	holder = &PrometheusHolder{
		httpLatency: httpLatency,
	}
	return holder
}

// OnRequestFinished will measure the http latency for the respective call to a missy service endpoint
func (p *PrometheusHolder) OnRequestFinished(method string, path string, statusCode int, processTimeMillis float64) {
	p.httpLatency.WithLabelValues(method, path, strconv.Itoa(statusCode)).Observe(processTimeMillis)
}
