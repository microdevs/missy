package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HTTPRequestsCollector allows to collect metrics about incoming HTTP requests.
type HTTPRequestsCollector struct {
	request *prometheus.HistogramVec
}

// NewHTTPRequestsCollector returns a new instance of a HTTPRequestsCollector.
func NewHTTPRequestsCollector(name string) *HTTPRequestsCollector {
	request := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: strings.Replace(name, "-", "_", -1) + "_http_handler_latency",
		Help: "HTTP Handler Latency by Endpoint",
	},
		[]string{"method", "path", "status_code"},
	)
	prometheus.MustRegister(request)
	return &HTTPRequestsCollector{
		request: request,
	}
}

// Send allows to send new metric about the HTTP request.
func (r *HTTPRequestsCollector) CollectDuration(method, name string, status int, duration time.Duration) {
	r.request.WithLabelValues(method, name, strconv.Itoa(status)).Observe(duration.Seconds())
}
