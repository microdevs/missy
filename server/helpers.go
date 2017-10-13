package server

import (
	"net/http"
	"github.com/gorilla/mux"
)

func Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}