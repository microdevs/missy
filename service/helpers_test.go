package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

func TestVars(t *testing.T) {
	req := &http.Request{}
	vars := Vars(req)
	ty := reflect.TypeOf(vars)
	if ty.String() != "map[string]string" {
		t.Errorf("Server's Var() function does not return a map[string]string, instead it returns a %s", ty)
	}
}

func TestTokenHasAccess(t *testing.T) {
	tests := []struct {
		Token          *jwt.Token
		Policy         string
		ExpectedResult bool
	}{
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"policy1", "policy2", "policy3"}}}}, "policy1", true},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"policy1", "policy2", "policy3"}}}}, "policy2", true},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"policy2", "policy3"}}}}, "policy1", false},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"policy1", "policy2", "policy3"}}}}, "policy4", false},
		{&jwt.Token{Claims: jwt.StandardClaims{}}, "policy1", false},
		{nil, "policy1", false},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": "not a map"}}, "policy1", false},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{"1", "2", "3"}}}}, "policy1", false},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"default": []interface{}{1, true, 0.283, "foo"}}}}, "policy1", false},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"": []interface{}{1, true, 0.283, "foo"}}}}, "policy1", false},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{}}}, "", false},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"": []interface{}{""}}}}, "", true},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"": []interface{}{""}}}}, "policy", false},
		{&jwt.Token{Claims: jwt.MapClaims{"policies": map[string]interface{}{"": []interface{}{"policy"}}}}, "policy", true},
	}
	for i, test := range tests {
		r := httptest.NewRequest(http.MethodGet, "/foo", strings.NewReader("foobar"))
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxToken, test.Token)
		r = r.WithContext(ctx)
		result := TokenHasAccess(r, test.Policy)
		if result != test.ExpectedResult {
			t.Logf("%d Result should be %t but was %t", i, test.ExpectedResult, result)
			t.Fail()
		}
	}
}

func TestTokenShouldNotPanicOnEmptyToken(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/foo", http.NoBody)
	// don't set a token
	IsRequestTokenValid(r) // check if this function won't panic
}

func TestIsRequestTokenValid(t *testing.T) {
	tests := []struct {
		token  *jwt.Token
		result bool
	}{
		{&jwt.Token{Valid: false}, false},
		{&jwt.Token{Valid: true}, true},
		{&jwt.Token{}, false},
		{nil, false},
	}
	for _, test := range tests {
		r := httptest.NewRequest(http.MethodGet, "/foo", http.NoBody)
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxToken, test.token)
		r = r.WithContext(ctx)
		result := IsRequestTokenValid(r)
		if result != test.result {
			t.Errorf("Result should be %t but was %t", test.result, result)
		}
	}
}

func TestIsSignedTokenValid(t *testing.T) {

	result := IsSignedTokenValid(generateSignedTokenString(t))
	if result != true {
		t.Error("Result should be true but was false")
	}

}

func TestIsSignedTokenNotValid(t *testing.T) {

	result := IsSignedTokenValid("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MjA4NzU1NjMsImlhdCI6MTUyMDg3NDk2MywiaXNhZG1pbiI6dHJ1ZSwicG9saWNpZXMiOlsiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5kb2N1bWVudHMuYWxsIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5kb2N1bWVudHMucGF5YWJsZXMiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmRvY3VtZW50cy5yZWNpZXZhYmxlcyIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQuYWNjb3VudGluZy5kZXRhaWxzIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5jb2NrcGl0LnByb2ZpdGFiaWxpdHkiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULnBheW1lbnRzLnBheW1lbnRSdW4iLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULnBheW1lbnRzLmJhbmtUcmFuc2FjdGlvbnMiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmNvY2twaXQuYm9va2luZ0pvdXJuYWwiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmNvY2twaXQuYmFsYW5jZSIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQuZGFzaGJvYXJkLndpZGdldC5teVRhc2tzIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5kb2N1bWVudHMucGFnZSIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQucGF5bWVudHMucGFnZSIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQucGF5bWVudHMuYWNjb3VudHNQYXlhYmxlIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5hY2NvdW50aW5nLmRldGFpbHMuZWRpdGFibGUiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmNvY2twaXQubGlxdWlkaXR5IiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5jb2NrcGl0LnN1c2EiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmRhc2hib2FyZC5wYWdlIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5nZW5lcmFsLnVwbG9hZCIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQuZG9jdW1lbnRzLm15VGFza3MiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmRvY3VtZW50cy5wZW5kaW5nIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5jb2NrcGl0LnBhZ2UiXSwidXNlcmlkIjoyLCJ1c2VybmFtZSI6InRlc3RAc21hY2MuaW8ifQ.mLfL0SKM5NWELiqJujKWUxpuXaApHX6q1QMqpPpMyB8R4CpRnK2w4NFxMZDWTIM28Dz1VyCOUFoJAk886Xe5L2t5-7LizZaXbKsrSQefvZj98D4SMMByCy95siQPJBBicwrPa4_FxjhtGn6gMR-NjKncZVCt64ry2n2Lr5BguXM")
	if result != false {
		t.Error("Result should be false but was true")
	}

}

func TestTokenClaims(t *testing.T) {
	tests := []struct {
		Token *jwt.Token
		Claim map[string]interface{}
	}{
		{&jwt.Token{Claims: jwt.MapClaims{}}, map[string]interface{}{}},
		{&jwt.Token{}, nil},
		{&jwt.Token{Claims: jwt.MapClaims{"test": "test"}}, map[string]interface{}{"test": "test"}},
		{&jwt.Token{Claims: jwt.MapClaims{"test": map[string]string{"t2": "v2"}}}, map[string]interface{}{"test": map[string]string{"t2": "v2"}}},
		{nil, nil},
	}
	// loop through the test cases specified above
	for _, test := range tests {
		r := httptest.NewRequest(http.MethodGet, "/foo", strings.NewReader("foobar"))
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxToken, test.Token)
		r = r.WithContext(ctx)
		result := TokenClaims(r)
		for i, c := range result {
			if reflect.DeepEqual(c, test.Claim[i]) != true {
				t.Logf("Result should be %v but was %v", test.Claim[i], c)
				t.Fail()
			}
		}
	}
}

func TestRawToken(t *testing.T) {

	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.Header.Add("Authorization", "Bearer token")

	token, err := RawToken(r)
	if err != nil {
		t.Log("error occoured calling raw token: ", err)
		t.Fail()
	}

	if token != "token" {
		t.Logf("token was \"%s\" expected \"%s\"", token, "token")
		t.Fail()
	}

}
