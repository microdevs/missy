package service

import (
	"testing"
	"reflect"
)

func TestNewPrometheusHolder(t *testing.T) {

	p := NewPrometheus("my-service")

	if ty := reflect.TypeOf(p).String(); ty != "*server.PrometheusHolder" {
		t.Errorf("NewPrometheus did not return a Pointer to Prometheus Holder but ", ty)
	}
}