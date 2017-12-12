package service

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
)

var holder *PrometheusHolder

type PrometheusHolder struct {
	httpLatency *prometheus.SummaryVec
}

func NewPrometheus(serviceName string) *PrometheusHolder {
	if holder != nil {
		return holder
	}

	prefix := strings.Replace(serviceName, "-", "_", -1)

	httpLatency := prometheus.NewSummaryVec(prometheus.SummaryOpts{
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

func (p *PrometheusHolder) OnRequestFinished(method string, path string, statusCode int, processTimeMillis float64) {
	p.httpLatency.WithLabelValues(method, path, strconv.Itoa(statusCode)).Observe(processTimeMillis)
}
