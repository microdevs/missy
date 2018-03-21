package service

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/microdevs/missy/config"
	"github.com/microdevs/missy/log"
)

var pubkey *rsa.PublicKey

func initPublicKey() {
	if pubkey != nil {
		return
	}

	pubkeyLocation := config.GetInstance().Authorization.PublicKeyFile
	pubkeyPEM, err := ioutil.ReadFile(pubkeyLocation)
	if err != nil {
		log.Fatal("Unable to load public key file for token auth: ", err)
	}

	pkey, err := jwt.ParseRSAPublicKeyFromPEM(pubkeyPEM)
	if err != nil {
		log.Fatal("Unable to parse public key for token auth: ", err)
	}
	pubkey = pkey
}

// StartTimerHandler is a middleware to start a timer for the request benchmark
func StartTimerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, RequestTimer, NewTimer())
		h.ServeHTTP(w, r)
	})
}

// AuthHandler is a middleware to authenticating a user or machine by validating an JWT auth token passed in the header of the request
func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		if reqToken == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		splitToken := strings.Split(reqToken, "Bearer ")
		reqToken = splitToken[1]
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(reqToken, &claims, func(t *jwt.Token) (interface{}, error) {
			return pubkey, nil
		})
		if err != nil {
			http.Error(w, "invalid token format", http.StatusBadRequest)
			return
		}

		if !token.Valid {
			log.Warn("Invalid token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		context.Set(r, "token", token)
		context.Set(r, "claims", claims)

		h.ServeHTTP(w, r)
	})
}

// FinalHandler measures the time of the request with the help of the timestamp taken in StartTimerHandler
// and writes it to a Prometheus metric. It will also write a log line of the request in the log file
func FinalHandler(h http.Handler) http.Handler {
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
		timer := context.Get(r, RequestTimer).(*Timer)
		prometheus := context.Get(r, PrometheusInstance).(*PrometheusHolder)
		prometheus.OnRequestFinished(r.Method, r.URL.Path, mw.Status(), timer.durationMillis())
	})
}

// infoHandler
func (s *Service) infoHandler(w http.ResponseWriter, r *http.Request) {
	info := fmt.Sprintf("Name %s\nUptime %s", s.name, s.timer.Uptime())
	w.Write([]byte(info))
}

// healthHandler
func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
