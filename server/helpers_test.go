package server

import (
	"testing"
	"net/http"
	"reflect"
)

func TestVars(t *testing.T) {
	req := &http.Request{}
	vars := Vars(req)
	ty := reflect.TypeOf(vars)
	if ty.String() != "map[string]string" {
		t.Errorf("Server's Var() function does not return a map[string]string, instead it returns a %s", ty)
	}
}
