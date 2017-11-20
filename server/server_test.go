package server

import (
	"testing"
	"reflect"
	"syscall"
	"net/http"
	"bytes"
	"regexp"
	"strings"
	"io/ioutil"
	"github.com/microdevs/missy/config"
	"os"
)

func TestMain(m *testing.M) {
	// setup
	var yml = []byte(`
name: test
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
	m.Run()
	//teardown
	os.Remove(config.MissyConfigFile)
}

func TestNewServer(t *testing.T) {

	s := NewServer()

	if ty := reflect.TypeOf(s).String(); ty != "*server.Server" {
		t.Errorf("NewServer did not return a Pointer to Server but %s", ty)
	}

	if s.name != "testname" {
		t.Errorf("Server's name was not set to \"testname\", got %s", s.name)
	}
}

func TestNewServerWithDifferentHostPort(t *testing.T) {
	testhost := "devil.hell"
	testport := "666"
	syscall.Setenv("LISTEN_HOST", testhost)
	syscall.Setenv("LISTEN_PORT", testport)

	s := NewServer()

	if s.Host != testhost {
		t.Errorf("Expected Host set to %s got %s", testhost, s.Host)
	}

	if s.Port != testport {
		t.Errorf("Expected Port set to %s got %s", testport, s.Port)
	}
}

func TestServerEndpoints(t *testing.T) {
	buf := new(bytes.Buffer)
	testhost := "localhost"
	testport := "8089"

	s := NewServer()
	s.Host = testhost
	s.Port = testport
	go s.StartServer()

	url := "http://" + s.Host + ":" + s.Port

	// Test /info endpoint
	response0, err0 := http.Get(url + "/info")

	if err0 != nil || response0.StatusCode != http.StatusOK {
		t.Errorf("Error calling /info endpoint: %s", err0)
	}

	buf.ReadFrom(response0.Body)
	body0 := buf.Bytes()
	exp := `^Name` + s.name + `\s*Uptime \d+\.\d{6}s|ms|µs`
	matches, errInfoRegex := regexp.Match(exp,body0)
	if errInfoRegex != nil || matches == false {
		t.Errorf("/info Response did not match expected response body, got %s, error: %v", string(body0), errInfoRegex)
	}
	buf.Reset()

	// Test /health endpoint
	response1, err1 := http.Get(url + "/health")
	if err1 != nil || response1.StatusCode != http.StatusOK {
		t.Errorf("Error calling /info endpoint, Status Code %d, error: %v", response1.StatusCode, err0)
	}
	buf.ReadFrom(response1.Body)

	if body1 := buf.String(); body1 != "OK" {
		t.Errorf("/health returned unexpected output, expected OK got %s", body1)
	}
	buf.Reset()

	// Test Prometheus output
	response2, err2 := http.Get(url + "/metrics")
	if err2 != nil || response2.StatusCode != http.StatusOK {
		t.Errorf("Error calling /metrics endpoint, Status Code %d, error: %v", response1.StatusCode, err0)
	}
	buf.ReadFrom(response2.Body)
	if body2 := buf.String(); !strings.Contains(body2, "go_gc_duration_seconds") {
		t.Errorf("/metrics returned unexpedted output, got\n%s",body2)
	}


}
