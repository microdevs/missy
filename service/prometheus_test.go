package service

import (
	"reflect"
	"testing"
)

func TestNewPrometheusHolder(t *testing.T) {

	p := NewPrometheus("my-service")

	if ty := reflect.TypeOf(p).String(); ty != "*service.PrometheusHolder" {
		t.Errorf("NewPrometheus did not return a Pointer to Prometheus Holder but %s", ty)
	}
}
