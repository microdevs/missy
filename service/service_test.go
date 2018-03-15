package service

import (
	"bytes"
	"github.com/microdevs/missy/config"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func runTestWithConfigFile(t *testing.T, f func(*testing.T)) {
	var yml = []byte(`
name: test
authorization:
  publicKeyFile: "test-fixtures/cert.pem"
environment:
  - envName: ENVVAR_A
    defaultValue: foo
    internalName: var.a
    mandatory: false
    usage: "This is the description for ENVVAR_A"
  - envName: ENVVAR_B
    defaultValue: "bar"
    internalName: var.b
    mandatory: false
    usage: "This is the description for ENVVAR_B"
`)
	ioutil.WriteFile(config.MissyConfigFile, yml, os.FileMode(0644))
	f(t)
	os.Remove(config.MissyConfigFile)
}

func TestNewService(t *testing.T) {
	runTestWithConfigFile(t, func(t *testing.T){
		s := New()
		if ty := reflect.TypeOf(s).String(); ty != "*service.Service" {
			t.Errorf("New() did not return a Pointer to Service but %s", ty)
		}
		if s.name != "test" {
			t.Errorf("Service's name was not set to \"test\", got %s", s.name)
		}
	})
}

func TestNewServiceWithDifferentHostPort(t *testing.T) {
	runTestWithConfigFile(t, func(t *testing.T) {
		testhost := "devil.hell"
		testport := "666"
		os.Setenv("LISTEN_HOST", testhost)
		os.Setenv("LISTEN_PORT", testport)

		s := New()

		if s.Host != testhost {
			t.Errorf("Expected Host set to %s got %s", testhost, s.Host)
		}

		if s.Port != testport {
			t.Errorf("Expected Port set to %s got %s", testport, s.Port)
		}
		http.DefaultServeMux = nil
	})
}

func TestInfoEndpoint(t *testing.T) {
	runTestWithConfigFile(t,func(t *testing.T){
		buf := new(bytes.Buffer)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://missy/info", nil)
		// Test /info endpoint
		s := New()
		s.Router.ServeHTTP(w,r)
		if w.Code != http.StatusOK {
			t.Errorf("Error calling /info endpoint")
		}

		buf.ReadFrom(w.Body)
		body0 := buf.Bytes()
		exp := `^Name ` + s.name + `\s*Uptime \d+\.\d+s|ms|µs`
		matches, errInfoRegex := regexp.Match(exp, body0)
		if errInfoRegex != nil || matches == false {
			t.Errorf("/info Response did not match expected response body, got %s, error: %v", string(body0), errInfoRegex)
		}
		http.DefaultServeMux = nil
	})
}

func TestHealthEndpoint(t *testing.T) {
	runTestWithConfigFile(t,func(t *testing.T){
		buf := new(bytes.Buffer)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://missy/health", nil)
		s := New()
		s.Router.ServeHTTP(w,r)
		if w.Code != http.StatusOK {
			t.Errorf("Error calling /info endpoint")
		}
		buf.ReadFrom(w.Body)

		if body1 := buf.String(); body1 != "OK" {
			t.Errorf("/health returned unexpected output, expected OK got %s", body1)
		}
		buf.Reset()
		http.DefaultServeMux = nil
	})
}

func TestPrometheusEndpoint(t *testing.T) {
	runTestWithConfigFile(t,func(t *testing.T) {
		buf := new(bytes.Buffer)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://missy/metrics", nil)
		s := New()
		s.Router.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("Error calling /info endpoint")
		}
		buf.ReadFrom(w.Body)
		if body2 := buf.String(); !strings.Contains(body2, "go_gc_duration_seconds") {
			t.Errorf("/metrics returned unexpedted output, got\n%s", body2)
		}
		http.DefaultServeMux = nil
	})

}

func TestFailingHandler(t *testing.T) {
	runTestWithConfigFile(t, func(t *testing.T){
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://missy.com/die", nil)
		s := New()
		s.HandleFunc("/die", func(w http.ResponseWriter, r *http.Request) {
			panic("triggering a panic!")
		})
		s.Router.ServeHTTP(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %v\n", w.Code)
		}
		http.DefaultServeMux = nil
	})
}
