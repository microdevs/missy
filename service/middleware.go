package service

import (
	"fmt"
	"net/http"
)

const (
	tokenClaimsPoliciesKey = "policies"
)

// PolicyManager allows to validate and fetch policies from the HTTP request.
type PolicyManager struct {
}

// PolicyValidInRequest validates policy in HTTP request.
func (p *PolicyManager) PolicyValidInRequest(policy string, req *http.Request) error {
	if !TokenHasAccess(req, policy) {
		return fmt.Errorf("token hasn't required policy")
	}
	return nil
}

// TokenPolicies returns all policies from the token in HTTP request.
func (p *PolicyManager) TokenPolicies(req *http.Request) (map[string]interface{}, bool) {
	policiesRaw, found := TokenClaims(req)[tokenClaimsPoliciesKey]
	if !found {
		return nil, false
	}
	policies, ok := policiesRaw.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return policies, true
}
