package service

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// Vars returns the gorilla/mux values from a request
func Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// Token returns the validated auth token from the request context
func Token(r *http.Request) *jwt.Token {
	return context.Get(r, "token").(*jwt.Token)
}
