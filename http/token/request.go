package token

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	tokenHeader   = "context"
	tokenPolicies = "policies"
)

// IsValidInRequest checks if request has a valid token
func IsValidInRequest(r *http.Request) bool {
	token := FromContext(r.Context())
	if token == nil {
		return false
	}
	return token.Valid
}

// RequestClaims get token claims.
func RequestClaims(r *http.Request) map[string]interface{} {
	token := FromContext(r.Context())
	if token == nil {
		return nil
	}
	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	return claimsMap
}

// RequestHasAccess checks if a valid access token contains a given policy in a context.
func RequestHasAccess(r *http.Request, policy string) bool {
	token := FromContext(r.Context())
	if token == nil {
		return false
	}
	// let's assume the claims are map claims because this is what our IAM delivers
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	// if the claims do not contain policies return false
	policiesInterfaces, ok := claims[tokenPolicies].(map[string]interface{})
	if !ok {
		return false
	}

	tokenPolicies := make(map[string][]interface{})

	for k := range policiesInterfaces {
		tokenPolicies[k], ok = policiesInterfaces[k].([]interface{})
		if !ok {
			return false
		}
	}
	// get header context for policy checks
	// if no context, will be empty sting
	ctx := r.Header.Get(tokenHeader)

	policies := tokenPolicies[ctx]
	if ctx != "" && len(tokenPolicies[""]) > 0 {
		policies = append(policies, tokenPolicies[""]...)
	}

	for _, p := range policies {
		ps, ok := p.(string)
		// if policy cannot asserted to type string we silently skip this entry
		if !ok {
			continue
		}
		if ps == policy {
			return true
		}
	}
	return false
}
