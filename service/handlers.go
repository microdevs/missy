package service

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/microdevs/missy/log"
)

var pubkey *rsa.PublicKey

func init() {
	Config().RegisterOptionalParameter("TOKEN_CA_FILE", "", "service.token.ca.file", "Set the location to the certificate used to validate the JWT tokens")
	Config().Parse()
	initPublicKey()
}

func initPublicKey() {
	if pubkey != nil {
		return
	}

	pubkeyLocation := Config().Get("service.token.ca.file")
	if pubkeyLocation == "" {
		log.Warnf("No location for the public key for token auth was set - secure handlers will not work")
		return
	}
	pubkeyPEM, err := ioutil.ReadFile(pubkeyLocation)
	if err != nil {
		log.Errorf("Unable to load public key file for token auth: %s", err)
		return
	}

	pkey, err := jwt.ParseRSAPublicKeyFromPEM(pubkeyPEM)
	if err != nil {
		log.Errorf("Unable to parse public key for token auth: %s", err)
		return
	}
	pubkey = pkey
}

// StartTimerHandler is a middleware to start a timer for the request benchmark
func StartTimerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestTimer, NewTimer())
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}

// AuthHandler is a middleware to authenticating a user or machine by validating an JWT auth token passed in the header of the request
func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if pubkey == nil {
			log.Error("Secure handler was called but public ca file is missing")
			http.Error(w, "This handler is unavailable due to a configuration error", http.StatusInternalServerError)
			return
		}

		// get request bearer token
		reqToken, err := RawToken(r)
		if err != nil {
			http.Error(w, "No Authorization Bearer token found", http.StatusBadRequest)
			return
		}
		token, err := jwt.Parse(reqToken, func(t *jwt.Token) (interface{}, error) {
			return pubkey, nil
		})
		if err != nil {
			log.Warnf("Invalid token: %v", err)
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxToken, token)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}

// FinalHandler measures the time of the request with the help of the timestamp taken in StartTimerHandler
// and writes it to a Prometheus metric. It will also write a log line of the request in the log file
func FinalHandler(pattern string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mw := &ResponseWriter{ResponseWriter: w}
			// if Hijacker interface is implemented switch to our HijackedResponseWriter
			if _, ok := w.(http.Hijacker); ok {
				w = &HijackerResponseWriter{mw}
			} else {
				w = mw
			}

			h.ServeHTTP(w, r)

			log.Infof("%s \"%s %s %s %d\" - %s", r.RemoteAddr, r.Method, r.URL, r.Proto, mw.Status(), r.UserAgent())
			timer, ok := r.Context().Value(RequestTimer).(*Timer)
			if !ok {
				log.Errorf("FinalHandler: couldn't get timer from request's context: val=%+#v", timer)
				return
			}
			prometheus, ok := r.Context().Value(PrometheusInstance).(*PrometheusHolder)
			if !ok {
				log.Errorf("FinalHandler: couldn't get prometheus from request's context: val=%+#v", prometheus)
				return
			}
			prometheus.OnRequestFinished(r.Method, pattern, mw.Status(), timer.durationMillis())
		})
	}
}

// infoHandler
func (s *Service) infoHandler(w http.ResponseWriter, r *http.Request) {
	info := fmt.Sprintf("Name %s\nUptime %s", s.name, s.timer.Uptime())
	w.Write([]byte(info))
}

// healthHandler
func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	s.StateProbes.MuHealthy.Lock()
	defer s.StateProbes.MuHealthy.Unlock()

	if !s.StateProbes.IsHealthy {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Not OK"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// readinessHandler
func (s *Service) readinessHandler(w http.ResponseWriter, r *http.Request) {
	s.StateProbes.MuReady.Lock()
	defer s.StateProbes.MuReady.Unlock()

	if !s.StateProbes.IsReady {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Not Ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))

}
