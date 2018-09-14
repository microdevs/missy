package service

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestVars(t *testing.T) {
	req := &http.Request{}
	vars := Vars(req)
	ty := reflect.TypeOf(vars)
	if ty.String() != "map[string]string" {
		t.Errorf("Server's Var() function does not return a map[string]string, instead it returns a %s", ty)
	}
}

var tokenHasAccessTestCases = []struct {
	Token          *jwt.Token
	Context        string
	Policy         string
	ExpectedResult bool
}{
	{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"policy1", "policy2", "policy3"}}}}, "default", "policy1", true},
	{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"policy1", "policy2", "policy3"}}}}, "default", "policy2", true},
	{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"policy1", "policy2", "policy3"}}}}, "other", "policy1", false},
	{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"policy1", "policy2", "policy3"}}}}, "default", "policy4", false},
	{&jwt.Token{Claims: jwt.StandardClaims{}}, "default", "policy1", false},
	{nil, "default", "policy1", false},
	{&jwt.Token{Claims: jwt.MapClaims{"policies": "not a map"}}, "default", "policy1", false},
	{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []int{1, 2, 3}}}}, "default", "policy1", false},
	{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{1, true, 0.283, "foo"}}}}, "default", "policy1", false},
}

func TestTokenHasAccess(t *testing.T) {
	// loop through the test cases specified above
	for _, test := range tokenHasAccessTestCases {
		r := httptest.NewRequest(http.MethodGet, "/foo", strings.NewReader("foobar"))
		context.Set(r, "token", test.Token)
		result := TokenHasAccess(r, test.Context, test.Policy)
		if result != test.ExpectedResult {
			t.Logf("Result should be %t but was %t", test.ExpectedResult, result)
			t.Fail()
		}
	}
}
