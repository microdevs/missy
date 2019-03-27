package service

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	s := New("test")
	if ty := reflect.TypeOf(s).String(); ty != "*service.Service" {
		t.Errorf("New() did not return a Pointer to Service but %s", ty)
	}
	if s.name != "test" {
		t.Errorf("Service's name was not set to \"test\", got %s", s.name)
	}
}

func TestNewServiceWithDifferentHostPort(t *testing.T) {
	testhost := "devil.hell"
	testport := "666"
	os.Setenv("LISTEN_HOST", testhost)
	os.Setenv("LISTEN_PORT", testport)
	Config().ParseEnvironment(true)

	s := New("test")

	if s.Host != testhost {
		t.Errorf("Expected Host set to %s got %s", testhost, s.Host)
	}

	if s.Port != testport {
		t.Errorf("Expected Port set to %s got %s", testport, s.Port)
	}
	http.DefaultServeMux = nil
}

func TestInfoEndpoint(t *testing.T) {

	buf := new(bytes.Buffer)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://missy/info", nil)
	// Test /info endpoint
	s := New("test")
	s.MetricsRouter.ServeHTTP(w, r)
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

}

func TestHealthEndpoint(t *testing.T) {
	buf := new(bytes.Buffer)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://missy/health", nil)
	s := New("test")
	s.MetricsRouter.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Error calling /info endpoint")
	}
	buf.ReadFrom(w.Body)

	if body1 := buf.String(); body1 != "OK" {
		t.Errorf("/health returned unexpected output, expected OK got %s", body1)
	}
	buf.Reset()
	http.DefaultServeMux = nil
}

func TestPrometheusEndpoint(t *testing.T) {
	buf := new(bytes.Buffer)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://missy/metrics", nil)
	s := New("test")
	s.MetricsRouter.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Error calling /info endpoint")
	}
	buf.ReadFrom(w.Body)
	if body2 := buf.String(); !strings.Contains(body2, "go_gc_duration_seconds") {
		t.Errorf("/metrics returned unexpedted output, got\n%s", body2)
	}
	http.DefaultServeMux = nil

}

func TestFailingHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://missy.com/die", nil)
	s := New("test")
	s.UnsafeHandleFunc("/die", func(w http.ResponseWriter, r *http.Request) {
		panic("triggering a panic!")
	})
	s.Router.ServeHTTP(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %v\n", w.Code)
	}
	http.DefaultServeMux = nil
}

type TLSTest struct {
	Certfile         string
	Keyfile          string
	WriteFile        bool
	Mode             os.FileMode
	ExpectedCertfile string
	ExpectedKeyfile  string
	ExpectedUseTLS   bool
}

var prepareTLSTestData = []TLSTest{
	{"/tmp/testcertfile", "/tmp/testkeyfile", true, 0644, "/tmp/testcertfile", "/tmp/testkeyfile", true},
	{"/tmp/testcertfile", "/tmp/testkeyfile", true, 0000, "", "", false},
	{"/tmp/testcertfile    ", "   /tmp/testkeyfile", true, 0644, "/tmp/testcertfile", "/tmp/testkeyfile", true},
	{"/tmp/testcertfile", "/tmp/testkeyfile", false, 0644, "", "", false},
}

func TestPrepareTLSSuccess(t *testing.T) {

	for i, test := range prepareTLSTestData {

		testCertFile := strings.Trim(test.Certfile, " ")
		testKeyFile := strings.Trim(test.Keyfile, " ")

		// cleanup before new test
		os.Chmod(testCertFile, 0644)
		os.Remove(testCertFile)
		os.Chmod(testKeyFile, 0644)
		os.Readlink(testKeyFile)

		t.Logf("Running Test #%d", i+1)

		if test.WriteFile == true {
			ioutil.WriteFile(testCertFile, []byte("certfile"), 0644)
			os.Chmod(testCertFile, test.Mode)
			ioutil.WriteFile(testKeyFile, []byte("keyfile"), 0644)
			os.Chmod(testKeyFile, test.Mode)
		}

		os.Setenv("TLS_CERTFILE", test.Certfile)
		os.Setenv("TLS_KEYFILE", test.Keyfile)

		certFile, keyFile, useTLS := prepareTLS()

		if certFile != test.ExpectedCertfile {
			t.Logf("Expected certfile to be %s but was %s", test.ExpectedCertfile, certFile)
			t.Fail()
		}

		if keyFile != test.ExpectedKeyfile {
			t.Logf("Expected keyfile to be %s but was %s", test.ExpectedKeyfile, keyFile)
			t.Fail()
		}

		if useTLS != test.ExpectedUseTLS {
			t.Logf("Expected useTLS to be %t but is %t", test.ExpectedUseTLS, useTLS)
			t.Fail()
		}

	}
}
