package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
)

func TestSecureHandlerAuth(t *testing.T) {
	runTestWithConfigFile(t, func(t *testing.T) {
		token := generateSignedTokenString(t)

		r := httptest.NewRequest("GET", "http://missy.com/test", nil)
		r.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()

		s := New()
		s.SecureHandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {

			// expect claims to be populated in context
			claims := context.Get(r, ContextClaims)

			username := claims.(jwt.MapClaims)[ClaimUsername].(string)
			userId := claims.(jwt.MapClaims)[ClaimUserId].(float64)

			if username != "test@test.de" {
				t.Log(fmt.Sprintf("claims username is expected to be test@test.de but is %s", username))
				t.Fail()
			}

			if userId != 526 {
				t.Log(fmt.Sprintf("claims userId is expected to be 526 but is %v", userId))
				t.Fail()
			}
			w.Write([]byte("test"))
		})
		s.Router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Log(fmt.Sprintf("Response code is expected to be 200 but is %d", w.Code))
			t.Fail()
		}
	})
}

func TestSecureHandlerAuthWithInvalidToken(t *testing.T) {
	runTestWithConfigFile(t, func(t *testing.T) {
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MjA4NzU1NjMsImlhdCI6MTUyMDg3NDk2MywiaXNhZG1pbiI6dHJ1ZSwicG9saWNpZXMiOlsiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5kb2N1bWVudHMuYWxsIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5kb2N1bWVudHMucGF5YWJsZXMiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmRvY3VtZW50cy5yZWNpZXZhYmxlcyIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQuYWNjb3VudGluZy5kZXRhaWxzIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5jb2NrcGl0LnByb2ZpdGFiaWxpdHkiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULnBheW1lbnRzLnBheW1lbnRSdW4iLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULnBheW1lbnRzLmJhbmtUcmFuc2FjdGlvbnMiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmNvY2twaXQuYm9va2luZ0pvdXJuYWwiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmNvY2twaXQuYmFsYW5jZSIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQuZGFzaGJvYXJkLndpZGdldC5teVRhc2tzIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5kb2N1bWVudHMucGFnZSIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQucGF5bWVudHMucGFnZSIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQucGF5bWVudHMuYWNjb3VudHNQYXlhYmxlIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5hY2NvdW50aW5nLmRldGFpbHMuZWRpdGFibGUiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmNvY2twaXQubGlxdWlkaXR5IiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5jb2NrcGl0LnN1c2EiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmRhc2hib2FyZC5wYWdlIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5nZW5lcmFsLnVwbG9hZCIsIjVhOGFhMzk5NmZhMjQ3MDAwNzZiMjFjOS5GVFQuZG9jdW1lbnRzLm15VGFza3MiLCI1YThhYTM5OTZmYTI0NzAwMDc2YjIxYzkuRlRULmRvY3VtZW50cy5wZW5kaW5nIiwiNWE4YWEzOTk2ZmEyNDcwMDA3NmIyMWM5LkZUVC5jb2NrcGl0LnBhZ2UiXSwidXNlcmlkIjoyLCJ1c2VybmFtZSI6InRlc3RAc21hY2MuaW8ifQ.mLfL0SKM5NWELiqJujKWUxpuXaApHX6q1QMqpPpMyB8R4CpRnK2w4NFxMZDWTIM28Dz1VyCOUFoJAk886Xe5L2t5-7LizZaXbKsrSQefvZj98D4SMMByCy95siQPJBBicwrPa4_FxjhtGn6gMR-NjKncZVCt64ry2n2Lr5BguXM"
		w := callWithToken(token)
		if w.Code != http.StatusBadRequest {
			t.Log(fmt.Sprintf("Response code is expected to be 400 but is %d", w.Code))
			t.Fail()
		}
	})
}

func TestSecureHandlerAuthWithNonToken(t *testing.T) {
	runTestWithConfigFile(t, func(t *testing.T) {
		token := "this is not a token"
		w := callWithToken(token)
		if w.Code != http.StatusBadRequest {
			t.Log(fmt.Sprintf("Response code is expected to be 400 but is %d", w.Code))
			t.Fail()
		}
	})
}

func callWithToken(token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "http://missy.com/test", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()

	s := New()
	s.SecureHandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})
	s.Router.ServeHTTP(w, r)
	return w
}

func generateSignedTokenString(t *testing.T) string {
	data, err := ioutil.ReadFile("test-fixtures/key.pem")
	if err != nil {
		t.Error("Unable to load private key: ", err)
	}
	pk, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		t.Error("Unable to parse data from private key file: ", err)
	}

	claims := jwt.MapClaims{}
	claims["username"] = "test@test.de"
	claims["userid"] = 526

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(pk)
	if err != nil {
		t.Error("Unable to sign token: ", err)
	}

	return tokenString
}
