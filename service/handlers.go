package service

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/microdevs/missy/log"
	"net/http"
	"strings"
)

// StartTimerHandler is a middleware to start a timer for the request benchmark
func StartTimerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, RequestTimer, NewTimer())
		h.ServeHTTP(w, r)
	})
}

func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// todo: pubkey := config.Get("auth.pubkey")
		pubkeyPEM := []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDJSZnimuTDZsnXFHB7fW5iPfCy
TpE4rO6ES/Loi3e3H7nxPwDe8YFSH5NgKHWJjV+iGHRhDXQJKPBHN+WR/O2+iyMs
+3f8nEdKgJKaZrNMEJva2zcfl+/tlTksNgQFij0DJ2NIjkNJHXW1IJa/d3iZSyo/
8KcYKuXuvlfNALmKSQIDAQAB
-----END PUBLIC KEY-----`)

		pubkey, err := jwt.ParseRSAPublicKeyFromPEM(pubkeyPEM)
		if err != nil {
			log.Error("Unable to parse public key for token auth")
			http.Error(w, "Authorization error", http.StatusInternalServerError)
		}

		reqToken := r.Header.Get("Authorization")
		if reqToken == "" {
			http.Error(w,"Unauthorized", http.StatusUnauthorized)
			return
		}
		splitToken := strings.Split(reqToken, "Bearer ")
		reqToken = splitToken[1]
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(reqToken, &claims, func(t *jwt.Token) (interface{}, error) {
			return pubkey, nil
		})

		if !token.Valid {
			log.Warn("Invalid token")
			http.Error(w,"Unauthorized", http.StatusUnauthorized)
			return
		}

		context.Set(r, "token", token)

		h.ServeHTTP(w, r)
	})
}

// AccessLogHandler writes an acccess logon after the response has been sent
func AccessLogHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw := &ResponseWriter{w, http.StatusOK}
		h.ServeHTTP(mw, r)
		log.Infof("%s \"%s %s %s\" %d - %s", r.RemoteAddr, r.Method, r.URL, r.Proto, mw.Status, r.UserAgent())
	})
}

// StopTimerHandler measures the time of the request with the help of the timestamp taken in StartTimerHandler and writes it to a Prometheus metric
func StopTimerHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw := &ResponseWriter{w, http.StatusOK}
		h.ServeHTTP(mw, r)
		timer := context.Get(r, RequestTimer).(*Timer)
		prometheus := context.Get(r, PrometheusInstance).(*PrometheusHolder)
		prometheus.OnRequestFinished(r.Method, r.URL.Path, mw.Status, timer.durationMillis())
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
