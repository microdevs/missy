package service

import (
	"github.com/microdevs/missy/log"
	"net/http"
	"reflect"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// Vars returns the gorilla/mux values from a request
func Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// token returns the validated auth token from the request context
func token(r *http.Request) *jwt.Token {
	t := context.Get(r, "token")
	if t == (*jwt.Token)(nil) {
		return nil
	}
	return t.(*jwt.Token)
}

// TokenHasAccess checks if a valid access token contains a given policy in a context
func TokenHasAccess(r *http.Request, context string, policy string) bool {
	token := token(r)
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
	contextPolicies, ok := claims["policies"].(map[string]interface{})
	if !ok {
		log.Warn("Invalid token format: policies inside claims are not of type map[string]interface{}")
		return false
	}
	// if the policies do not contain the context return false
	if contextPolicies[context] == nil {
		return false
	}
	policies, ok := contextPolicies[context].([]interface{})
	if !ok {
		log.Warn("Invalid token format: Policy context value does not match type []interface{}")
		return false
	}
	// look for the policy that matches the requested policy
	// if found return true
	for _, p := range policies {
		ps, ok := p.(string)
		// if policy cannot asserted to type string we silently skip this entry
		if !ok {
			log.Warnf("Invalid token format: Policy is not of type string but %s", reflect.TypeOf(p).String())
			continue
		}
		if ps == policy {
			return true
		}
	}
	// if the requested policy was not found in the context, the function returns false
	return false
}

// IsRequestTokenValid checks if request has a valid token
func IsRequestTokenValid(r *http.Request) bool {
	token := token(r)
	// return false if there is no token
	if token == nil {
		return false
	}

	return token.Valid
}

// IsSignedTokenValid checks if provided signed token string is valid
func IsSignedTokenValid(signedToken string) bool {
	initPublicKey()
	token, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return pubkey, nil
	})

	if err != nil {
		log.Warnf("cannot pars jwt token: %v", err)
		return false
	}

	return token.Valid
}
