package service

import (
	"fmt"
	"net/http"
)

type PolicyRequestValidator struct {
}

// PolicyValidInRequest validates policy in request
func (p *PolicyRequestValidator) PolicyValidInRequest(policy string, req *http.Request) error {
	if !TokenHasAccess(req, policy) {
		return fmt.Errorf("token hasn't required policy")
	}
	return nil
}
