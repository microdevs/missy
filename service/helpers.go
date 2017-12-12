package service

import (
	"github.com/gorilla/mux"
	"net/http"
)

// Vars returns the gorilla/mux values from a request
func Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}
