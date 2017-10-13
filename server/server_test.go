package server

import (
	"testing"
	"reflect"
	"syscall"
	"net/http"
	"bytes"
	"regexp"
	"strings"
)

func TestNewServer(t *testing.T) {

	s := NewServer("testname")

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

	s := NewServer("evil")

	if s.Host != testhost {
		t.Errorf("Expected Host set to %s got %s", testhost, s.Host)
	}

	if s.Port != testport {
		t.Errorf("Expected Port set to %s got %s", testport, s.Port)
	}
}

func TestServerEndpoints(t *testing.T) {
	buf := new(bytes.Buffer)
	testname := "testendpoint"
	testhost := "localhost"
	testport := "8089"

	s := NewServer(testname)
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
	exp := `^Name` + testname + `\s*Uptime \d+\.\d{6}s|ms|µs`
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