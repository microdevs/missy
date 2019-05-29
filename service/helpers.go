package service

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/microdevs/missy/log"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type ctxKey string

const (
	ctxToken ctxKey = "token"
)

// Vars returns the gorilla/mux values from a request
func Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// Token returns the validated auth token from the request context
func Token(r *http.Request) *jwt.Token {
	ti := r.Context().Value(ctxToken)
	if ti, ok := ti.(*jwt.Token); ok {
		return ti
	}
	return nil
}

func RawToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")

	// check if there is a Bearer token (token will be at index 1)
	if len(splitToken) < 2 {
		return "", errors.New("error getting raw token: malformed authorization header")
	}

	return splitToken[1], nil
}

// TokenHasAccess checks if a valid access token contains a given policy in a context
func TokenHasAccess(r *http.Request, policy string) bool {
	token := Token(r)
	// return false if there is no token
	if token == nil {
		return false
	}
	// let's assume the claims are map claims because this is what our IAM delivers
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Warn("Invalid token format: token claims are not of type jwt.MapClaims")
		return false
	}
	// if the claims do not contain policies return false
	policiesInterfaces, ok := claims["policies"].(map[string]interface{})
	if !ok {
		log.Warn("Invalid token format: policies inside claims are not of type map[string]interface{}")
		return false
	}

	tokenPolicies := make(map[string][]interface{})

	for k := range policiesInterfaces {
		tokenPolicies[k], ok = policiesInterfaces[k].([]interface{})
		if !ok {
			log.Warn("Invalid token format: policies inside claims are not of type map[string][]interface{}")
			return false
		}
	}
	for _, policies := range tokenPolicies {
		for _, p := range policies {
			ps, ok := p.(string)
			if !ok {
				log.Warnf("Invalid token format: Policy is not of type string but %s", reflect.TypeOf(p).String())
				continue
			}
			if ps == policy {
				return true
			}
		}
	}
	// if the requested policy was not found in the context, the function returns false
	return false
}

// IsRequestTokenValid checks if request has a valid token
func IsRequestTokenValid(r *http.Request) bool {
	token := Token(r)
	// return false if there is no token
	if token == nil {
		return false
	}

	return token.Valid
}

// IsSignedTokenValid checks if provided signed token string is valid
func IsSignedTokenValid(signedToken string) bool {
	initPublicKey()
	if pubkey == nil {
		log.Error("No public key is set to validate the token.")
		return false
	}

	token, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return pubkey, nil
	})

	if err != nil {
		log.Warnf("Cannot parse jwt token: %v", err)
		return false
	}

	return token.Valid
}

// TokenClaims get token claims
func TokenClaims(r *http.Request) map[string]interface{} {
	token := Token(r)
	// return false if there is no token
	if token == nil {
		return nil
	}

	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Warn("Cannot cast token claims to jwt.MapClaims")
		return nil
	}

	return claimsMap
}
