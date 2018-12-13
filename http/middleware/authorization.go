package middleware

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/microdevs/missy/http/token"
	"github.com/microdevs/missy/log"
	"github.com/pkg/errors"
)

type Config struct {
	PublicKey []byte `env:"HTTP_SERVER_AUTHORIZATION_PUBLIC_KEY"`
}

func (ac *Config) FromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Errorf("reading certificate file err: %s", err)
	}
	ac.PublicKey = data
	return nil
}

func Authorization(c Config, l log.FieldsLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		if c.PublicKey == nil {
			return h
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authSplit := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(authSplit) < 2 {
				http.Error(w, "No header 'Authorization' Bearer found.", http.StatusBadRequest)
				return
			}
			t, err := token.Parse(c.PublicKey, authSplit[1])
			if err != nil {
				l.Error(err)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			r.WithContext(token.ContextWithToken(r.Context(), t))
			h.ServeHTTP(w, r)
		})
	}
}
